package tevm

import (
	"encoding/json"
	"errors"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/newalchemylimited/seth"
)

type callArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      seth.Uint64     `json:"gas"`
	GasPrice seth.Uint64     `json:"gasPrice"`
	Value    seth.Int        `json:"value"`
	Data     seth.Data       `json:"data"`
}

func (a *callArgs) Ref() vm.ContractRef {
	return (*acctref)(&a.From)
}

// Execute implements seth.Transport.
func (c *Chain) Execute(method string, params []json.RawMessage, res interface{}) error {
	c.mu.Lock()
	if len(params) == 0 {
		c.mu.Unlock()
		return errors.New(method + ": not enough params")
	}
	ret, err := c.execute(method, params[0])
	if err != nil {
		c.mu.Unlock()
		return err
	}
	c.mu.Unlock()
	return marshal(ret, res)
}

func (c *Chain) execute(method string, param json.RawMessage) (interface{}, error) {
	switch method {
	case "eth_call":
		a := new(callArgs)
		if err := marshal(param, a); err != nil {
			return nil, err
		}
		return c.constCall(a)
	case "eth_sendTransaction":
		a := new(callArgs)
		if err := marshal(param, a); err != nil {
			return nil, err
		}
		return c.send(a)
	case "eth_getTransactionReceipt":
		var h seth.Hash
		if err := marshal(param, &h); err != nil {
			return nil, err
		}
		return c.receipt(h)
	case "eth_getTransactionByHash":
		var h seth.Hash
		if err := marshal(param, &h); err != nil {
			return nil, err
		}
		return c.transaction(h)
	case "eth_getBalance":
		var addr seth.Address
		if err := marshal(param, &addr); err != nil {
			return nil, err
		}
		return c.balance(&addr)
	case "eth_estimateGas":
		a := new(callArgs)
		if err := marshal(param, a); err != nil {
			return nil, err
		}
		return c.estimate(a)
	default:
		return nil, errors.New(method + ": unsupported method")
	}
}

// constCall handles eth_call.
func (c *Chain) constCall(a *callArgs) (seth.Data, error) {
	evm := c.evm(a.From)
	gas := uint64(c.Block.GasLimit)
	if a.Gas != 0 {
		gas = uint64(a.Gas)
	}
	var to common.Address
	if a.To != nil {
		to = *a.To
	}
	ret, _, err := evm.StaticCall(a.Ref(), to, a.Data, gas)
	if err != nil {
		return nil, err
	}
	return seth.Data(ret), nil
}

// send handles eth_sendTransaction.
func (c *Chain) send(a *callArgs) (seth.Hash, error) {
	evm := c.evm(a.From)
	nonce := evm.StateDB.GetNonce(a.From)
	used := int64(a.Gas)
	status := 1

	if a.GasPrice != 0 {
		evm.GasPrice.SetUint64(uint64(a.GasPrice))
	}

	var contract common.Address

	if a.To == nil {
		_, addr, rem, err := evm.Create(a.Ref(), a.Data, uint64(a.Gas), a.Value.Big())
		if err != nil {
			status = 0
		}
		used -= int64(rem)
		contract = addr
	} else {
		_, rem, err := evm.Call(a.Ref(), *a.To, a.Data, uint64(a.Gas), a.Value.Big())
		if err != nil {
			status = 0
		}
		used -= int64(rem)
	}

	tx := &Transaction{
		BlockHash:   c.Block.Hash(),
		BlockNumber: seth.Uint64(c.Block.Number),
		From:        seth.Address(a.From),
		Gas:         seth.Uint64(used),
		GasPrice:    seth.Uint64(evm.Context.GasPrice.Uint64()),
		Input:       a.Data,
		Nonce:       seth.Uint64(nonce),
		To:          (*seth.Address)(a.To),
		Index:       1,
		Value:       a.Value,
	}
	rand.Read(tx.Hash[:])

	c.Block.GasUsed += used
	c.Block.Transactions = append(c.Block.Transactions, tx.Hash)
	c.Blocks[c.Block.Number] = c.Block

	c.Block = &Block{
		Coinbase:   c.Block.Coinbase,
		Number:     c.Block.Number + 1,
		Time:       now(),
		GasPrice:   c.Block.GasPrice,
		GasLimit:   c.Block.GasLimit,
		Difficulty: c.Block.Difficulty,
	}

	rt := &types.Receipt{
		Status:            uint(status),
		CumulativeGasUsed: big.NewInt(c.Block.GasUsed),
		TxHash:            (common.Hash)(tx.Hash),
		ContractAddress:   contract,
		GasUsed:           big.NewInt(used),
	}

	c.transactions[tx.Hash] = tx
	c.receipts[tx.Hash] = rt

	return tx.Hash, nil
}

// receipt handles eth_getTransactionReceipt.
func (c *Chain) receipt(h seth.Hash) (*types.Receipt, error) {
	r := c.receipts[h]
	if r == nil {
		return nil, errors.New("not found")
	}
	return r, nil
}

// transaction handles eth_getTransactionByHash.
func (c *Chain) transaction(h seth.Hash) (*Transaction, error) {
	tx := c.transactions[h]
	if tx == nil {
		return nil, errors.New("not found")
	}
	return tx, nil
}

// balance handles eth_getBalance.
func (c *Chain) balance(addr *seth.Address) (*seth.Int, error) {
	acct, _ := c.State.Accounts.GetAccount(addr)
	bal := acct.Balance()
	return &bal, nil
}

// estimate handles eth_estimateGas.
func (c *Chain) estimate(a *callArgs) (seth.Uint64, error) {
	evm := c.evm(a.From)
	evm.Context.GasPrice.SetInt64(0)
	snap := evm.StateDB.Snapshot()

	gas := uint64(c.Block.GasLimit)
	if a.Gas != 0 {
		gas = uint64(a.Gas)
	}

	if a.To == nil {
		_, _, rem, err := evm.Create(a.Ref(), a.Data, gas, a.Value.Big())
		if err != nil {
			return 0, err
		}
		gas -= rem
	} else {
		_, rem, err := evm.Call(a.Ref(), *a.To, a.Data, gas, a.Value.Big())
		if err != nil {
			return 0, err
		}
		gas -= rem
	}

	evm.StateDB.RevertToSnapshot(snap)

	return seth.Uint64(gas), nil
}

func marshal(from, to interface{}) error {
	b, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, to)
}
