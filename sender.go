package seth

import (
	"errors"
	"fmt"
	"math/big"
	"time"
)

// ErrCannotCancel is returned when attempting to cancel a transaction that has
// already been mined.
var ErrCannotCancel = errors.New("seth: cannot cancel")

type Sender interface {
	// From Client... TODO: Clean this up
	FilterTopics(topics []*Hash, addr *Address, start, end int64) (*Filter, error)
	GetBlock(num int64, txs bool) (*Block, error)
	GetReceipt(tx *Hash) (*Receipt, error)

	Create(code []byte, value *Int, constructor string, args ...interface{}) (*Receipt, error)
	Send(to *Address, method string, args ...interface{}) (Hash, error)
	ConstCall(to *Address, method string, out interface{}, args ...interface{}) error
	Cancel(h *Hash) (Hash, error)
	Wait(h *Hash) error
	Drain(prompt ...func(t *Transaction)) error
}

type SenderOptions struct {
	// A Signer can be used to sign raw transactions for this sender. If
	// this is set, all transactions will be sent as raw transactions.
	Signer Signer

	// GasRatio is the ratio of the gas estimate
	// to use as the gas offered for a transaction,
	// expressed as a rational number.
	// For instance, Num=5,Denom=4 would offer 5/4ths of
	// the gas estimate as the gas for a transaction.
	GasRatio struct {
		Num, Denom int
	}

	// GasPrice is the gas price offered for each transaction.
	GasPrice Int

	// Pending decides if "pending" is sent instead of "latest" for the defaultBlock parameter
	// See: https://github.com/ethereum/wiki/wiki/JSON-RPC#the-default-block-parameter
	Pending bool
}

var DefaultSenderOptions SenderOptions

func init() {

	if DefaultSenderOptions.GasRatio.Num == 0 {
		DefaultSenderOptions.GasRatio.Num = 10
		DefaultSenderOptions.GasRatio.Denom = 5
	}

	if DefaultSenderOptions.GasPrice.Int64() == 0 {
		(*big.Int)(&DefaultSenderOptions.GasPrice).SetString("50000000000", 10) // 50 Gwei
	}

}

// Sender is a client that sends transactions
// from a particular address.
type sender struct {
	*Client
	Addr *Address

	options SenderOptions
}

// NewSender constructs a Sender
func NewSender(c *Client, from *Address, options SenderOptions) Sender {
	return &sender{Client: c, Addr: from, options: options}
}

func (s *sender) pad(gas *Int) *Int {
	if gas == nil {
		return nil
	}
	ob := new(big.Int)
	gb := (*big.Int)(gas)
	var num, denom big.Int
	num.SetInt64(int64(s.options.GasRatio.Num))
	denom.SetInt64(int64(s.options.GasRatio.Denom))
	ob.Set(gb)
	ob.Mul(ob, &num)
	ob.Div(ob, &denom)

	return (*Int)(ob)
}

func (s *sender) ConstCall(to *Address, method string, out interface{}, args ...interface{}) error {
	opts := CallOpts{To: to, From: s.Addr, GasPrice: &s.options.GasPrice}
	opts.EncodeCall(method, args...)
	return s.Client.ConstCall(&opts, out, s.options.Pending)
}

// Create creates a new contract with the given contract code.
// This call blocks until the transaction posts, and then returns
// the contract's address.
func (s *sender) Create(code []byte, value *Int, constructor string, args ...interface{}) (*Receipt, error) {

	if constructor != "" && constructor != "()" {
		ethargs := outgoingArgConvert(constructor, args)
		argsAbi := ABIEncode(true, constructor, ethargs...)
		code = append(code, argsAbi...)
	}

	opts := CallOpts{From: s.Addr, GasPrice: &s.options.GasPrice, Value: value}
	opts.Data = Data(code)
	gas, err := s.EstimateGas(&opts, s.options.Pending)
	if err != nil {
		return nil, fmt.Errorf("error estimating gas: %s", err)
	}
	opts.Gas = s.pad(&gas)
	h, err := s.Call(&opts)
	if err != nil {
		return nil, fmt.Errorf("error sending create call: %s", err)
	}
	err = s.Wait(&h)
	if err != nil {
		return nil, fmt.Errorf("error waiting for create transaction: %s", err)
	}
	r, err := s.GetReceipt(&h)
	if err != nil {
		return nil, fmt.Errorf("error waiting for create receipt: %s", err)
	}
	if r.Address == nil {
		return nil, fmt.Errorf("txhash %s: contract not created", &h)
	}
	return r, nil
}

// Call makes a transaction call using the given CallOpts. Omitted fields are
// populated with default values.
func (s *sender) Call(opts *CallOpts) (Hash, error) {
	if opts.From == nil {
		opts.From = s.Addr
	}

	if opts.GasPrice == nil {
		opts.GasPrice = &s.options.GasPrice
	}

	if opts.Gas == nil {
		gas, err := s.EstimateGas(opts, s.options.Pending)
		if err != nil {
			return Hash{}, err
		}
		opts.Gas = s.pad(&gas)
	}

	if s.options.Signer == nil {
		return s.Client.Call(opts)
	}

	tx := opts.Transaction()

	// if no nonce was specified, try to select it
	if opts.Nonce == nil {
		if tx.From == nil {
			return Hash{}, fmt.Errorf("Sender.Call: unspecified nonce, and no from address provided")
		}
		n, err := s.GetNonceAt(tx.From, Pending)
		if err != nil {
			return Hash{}, err
		}
		tx.Nonce = Uint64(n)
	}
	hash := tx.HashToSign()

	sig, err := s.options.Signer(hash)
	if err != nil {
		return Hash{}, err
	}

	// If a from address was provided, verify that the signer produced a
	// signature for the correct address.
	if opts.From != nil {
		pub, err := sig.Recover(hash)
		if err != nil {
			return Hash{}, err
		}
		from := pub.Address()
		if *from != *opts.From {
			return Hash{}, fmt.Errorf(
				"sender: address mismatch: expected %v, got %v",
				opts.From, pub.Address())
		}
	}

	return s.RawCall(tx.Encode(sig))
}

// Send makes a contract call from the sender address.
// It automatically handles gas estimation and padding.
func (s *sender) Send(to *Address, method string, args ...interface{}) (Hash, error) {
	opts := CallOpts{To: to}
	opts.EncodeCall(method, args...)
	return s.Call(&opts)
}

// Cancel a transaction with the given hash.
func (s *sender) Cancel(h *Hash) (Hash, error) {
	tx, err := s.GetTransaction(h)
	if err != nil {
		return Hash{}, err
	} else if tx.TxIndex != nil {
		return Hash{}, ErrCannotCancel
	}
	opts := CallOpts{To: s.Addr, From: s.Addr, Nonce: &tx.Nonce}
	if s.options.GasPrice.Cmp(&tx.GasPrice) > 0 {
		opts.GasPrice = &s.options.GasPrice
	} else {
		opts.GasPrice = NewInt(tx.GasPrice.Int64() + 1)
	}
	return s.Call(&opts)
}

// Wait waits for a transaction hash to be mined into the canonical chain.
func (s *sender) Wait(h *Hash) error {
	for {
		t, err := s.GetTransaction(h)

		if err != nil {
			return err
		}
		if t.TxIndex != nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
}

// Drain waits for the pending transaction pool to
// contain no transactions from this account.
func (s *sender) Drain(prompt ...func(t *Transaction)) error {
	for {
		p, err := s.GetBlock(-1, true)
		if err != nil {
			return err
		}
		txs, err := p.ParseTransactions()
		if err != nil {
			return err
		}
		var t *Transaction
		for i := range txs {
			if txs[i].From == s.Addr {
				t = &txs[i]
				break
			}
		}
		if t == nil {
			return nil
		}
		for _, p := range prompt {
			p(t)
		}
		if err := s.Wait(&t.Hash); err != nil {
			return err
		}
	}
}
