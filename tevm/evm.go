package tevm

import (
	"encoding/binary"
	"encoding/json"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/newalchemylimited/seth"
	"github.com/tinylib/msgp/msgp"
)

// an account is a tuple of (balance, nonce, suicided)
type Account [32 + 8 + 1]byte

func (a *Account) Balance() seth.Int {
	var b big.Int
	b.SetBytes(a[:32])
	return seth.Int(b)
}

func (a *Account) SetBalance(v *big.Int) {
	buf := v.Bytes()
	for i := range a[:32-len(buf)] {
		a[i] = 0
	}
	copy(a[32-len(buf):32], buf)
}

func (a *Account) Nonce() uint64 {
	return binary.BigEndian.Uint64(a[32:])
}

func (a *Account) SetNonce(n uint64) {
	binary.BigEndian.PutUint64(a[32:], n)
}

func (a *Account) Suicided() bool {
	return a[32+8] != 0
}

func (a *Account) SetSuicided(t bool) {
	if t {
		a[32+8] = 1
	} else {
		a[32+8] = 0
	}
}

// default vm.Config
var theconfig = vm.Config{
	Debug:                   false,
	EnableJit:               false,
	ForceJit:                false,
	Tracer:                  nil,
	NoRecursion:             false,
	DisableGasMetering:      false,
	EnablePreimageRecording: false,
}

var theparams = params.ChainConfig{
	ChainId:        new(big.Int).SetInt64(5),
	HomesteadBlock: new(big.Int),
	EIP150Block:    new(big.Int),
	EIP155Block:    new(big.Int),
	EIP158Block:    new(big.Int),
}

type CodeTree struct {
	Tree
}

// GetCode gets the code associated with an address
func (c *CodeTree) GetCode(addr *seth.Address) []byte {
	return c.Tree.Get(addr[:])
}

// PutCode sets the code associated with an address
func (c *CodeTree) PutCode(addr *seth.Address, code []byte) {
	c.Tree.Insert(addr[:], code)
}

type AccountTree struct {
	Tree
}

// GetAccount gets an account
func (a *AccountTree) GetAccount(addr *seth.Address) (Account, bool) {
	var acct Account
	v := a.Tree.Get(addr[:])
	copy(acct[:], v)
	return acct, len(v) == len(acct)
}

// SetAccount sets an account
func (a *AccountTree) SetAccount(addr *seth.Address, acct *Account) {
	a.Tree.Insert(addr[:], acct[:])
}

// State database for the EVM.
type State struct {
	Refund seth.Uint64
	Trace  func(fn string, args ...interface{}) `json:"-"`

	Pending *seth.Block

	Accounts AccountTree
	Code     CodeTree
	Storage  Tree // key = hash(address, pointer)
	Preimage Tree

	Transactions Tree // key = txhash, value = serialized tx
	Receipts     Tree // key = txhash, value = serialized rx

	Blocks Tree // key = n2h(blocknum) = hash, value = serialized block

	Logs []*types.Log
	snapshots
}

type snapshots struct {
	Snapshots []statesnap
}

// StateDB returns a view of s that implements vm.StateDB.
func (s *State) StateDB() vm.StateDB {
	return (*gethState)(s)
}

// Hide the implementation of geth's vm.StateDB so that we don't leak all of
// these methods into the documentation.
type gethState State

type statesnap struct {
	Refund   seth.Uint64
	Accounts int
	Code     int
	State    int
	LogLen   int
	TXs      int
	RXs      int
}

func (s *gethState) CreateAccount(addr common.Address) {
	if s.Trace != nil {
		s.Trace("CreateAccount", addr.String())
	}
	var empty Account
	a := seth.Address(addr)
	s.Accounts.SetAccount(&a, &empty)
}

func (s *gethState) SubBalance(addr common.Address, v *big.Int) {
	if s.Trace != nil {
		s.Trace("SubBalance", addr.String(), v.String())
	}
	a := seth.Address(addr)
	acct, _ := s.Accounts.GetAccount(&a)
	bal := acct.Balance()
	b := bal.Big()
	b.Sub(b, v)
	var newacct Account
	copy(newacct[:], acct[:])
	newacct.SetBalance(b)
	s.Accounts.SetAccount(&a, &newacct)
}

func (s *gethState) AddBalance(addr common.Address, v *big.Int) {
	if s.Trace != nil {
		s.Trace("AddBalance", addr.String(), v.String())
	}
	a := seth.Address(addr)
	acct, _ := s.Accounts.GetAccount(&a)
	bal := acct.Balance()
	b := bal.Big()
	b.Add(b, v)
	var newacct Account
	copy(newacct[:], acct[:])
	newacct.SetBalance(b)
	s.Accounts.SetAccount(&a, &newacct)
}

func (s *gethState) GetBalance(addr common.Address) *big.Int {
	if s.Trace != nil {
		s.Trace("GetBalance", addr.String())
	}
	a := seth.Address(addr)
	acct, _ := s.Accounts.GetAccount(&a)
	bal := acct.Balance()
	return bal.Big()
}

func (s *gethState) GetNonce(addr common.Address) uint64 {
	if s.Trace != nil {
		s.Trace("GetNonce", addr.String())
	}
	a := seth.Address(addr)
	acct, _ := s.Accounts.GetAccount(&a)
	return acct.Nonce()
}

func (s *gethState) SetNonce(addr common.Address, n uint64) {
	if s.Trace != nil {
		s.Trace("SetNonce", addr.String(), n)
	}
	a := seth.Address(addr)
	acct, ok := s.Accounts.GetAccount(&a)
	if !ok {
		panic("SetNonce called on account that doesn't exist")
	}
	acct.SetNonce(n)
	s.Accounts.SetAccount(&a, &acct)
}

func (s *gethState) GetCodeHash(addr common.Address) common.Hash {
	if s.Trace != nil {
		s.Trace("GetCodeHash", addr.String())
	}
	return common.Hash(seth.HashBytes(s.GetCode(addr)))
}

func (s *gethState) GetCode(addr common.Address) []byte {
	if s.Trace != nil {
		s.Trace("GetCode", addr.String())
	}
	a := seth.Address(addr)
	return s.Code.GetCode(&a)
}

func (s *gethState) SetCode(addr common.Address, data []byte) {
	if s.Trace != nil {
		s.Trace("SetCode", addr.String(), data)
	}
	a := seth.Address(addr)
	s.Code.PutCode(&a, data)
}

func (s *gethState) GetCodeSize(addr common.Address) int {
	if s.Trace != nil {
		s.Trace("GetCodeSize", addr.String())
	}
	return len(s.GetCode(addr))
}

func (s *gethState) AddRefund(v uint64) {
	s.Refund += seth.Uint64(v)
}

func (s *gethState) GetRefund() uint64 {
	return uint64(s.Refund)
}

func stateKey(addr *common.Address, hash *common.Hash) seth.Hash {
	var v [20 + 32]byte
	copy(v[:], addr[:])
	copy(v[20:], hash[:])
	return seth.HashBytes(v[:])
}

func (s *gethState) GetState(addr common.Address, hash common.Hash) common.Hash {
	if s.Trace != nil {
		s.Trace("GetState", addr.String(), hash.String())
	}
	h := stateKey(&addr, &hash)
	var out common.Hash
	v := s.Storage.Get(h[:])
	copy(out[:], v)
	return out
}

var zerohash common.Hash

func (s *gethState) SetState(addr common.Address, hash, value common.Hash) {
	if s.Trace != nil {
		s.Trace("SetState", addr.String(), hash.String(), value.String())
	}
	h := stateKey(&addr, &hash)
	if value == zerohash {
		s.Storage.Delete(h[:])
	} else {
		s.Storage.Insert(h[:], value[:])
	}
}

func (s *gethState) Exist(addr common.Address) bool {
	if s.Trace != nil {
		s.Trace("Exist", addr.String())
	}
	a := seth.Address(addr)
	_, ok := s.Accounts.GetAccount(&a)
	return ok
}

func (s *gethState) Empty(addr common.Address) bool {
	if s.Trace != nil {
		s.Trace("Empty", addr.String())
	}
	a := seth.Address(addr)
	acct, ok := s.Accounts.GetAccount(&a)
	bal := acct.Balance()
	return !ok || (acct.Nonce() == 0 && bal.IsZero() && len(s.Code.GetCode(&a)) == 0)
}

func (s *gethState) Suicide(addr common.Address) bool {
	if s.Trace != nil {
		s.Trace("Suicide", addr.String())
	}
	a := seth.Address(addr)
	acct, ok := s.Accounts.GetAccount(&a)
	if !ok || acct.Suicided() {
		return false
	}
	acct.SetSuicided(true)
	s.Accounts.SetAccount(&a, &acct)
	return true
}

func (s *gethState) HasSuicided(addr common.Address) bool {
	if s.Trace != nil {
		s.Trace("HasSuicided", addr.String())
	}
	a := seth.Address(addr)
	acct, ok := s.Accounts.GetAccount(&a)
	return ok && acct.Suicided()
}

func (s *gethState) RevertToSnapshot(v int) {
	if s.Trace != nil {
		s.Trace("RevertToSnapshot", v)
	}
	snaps := s.Snapshots
	if len(snaps) <= v || v < 0 {
		panic("no such snapshot")
	}
	ns := snaps[v]
	s.Refund = ns.Refund
	s.Accounts.Rollback(ns.Accounts)
	s.Code.Rollback(ns.Code)
	s.Storage.Rollback(ns.State)
	s.Transactions.Rollback(ns.TXs)
	s.Receipts.Rollback(ns.RXs)
	s.Logs = s.Logs[:ns.LogLen]

	// make sure we can't roll forward
	snaps = snaps[:v]
}

func (s *gethState) Snapshot() int {
	if s.Trace != nil {
		s.Trace("Snapshot")
	}
	snap := statesnap{
		Refund:   s.Refund,
		Accounts: s.Accounts.Snapshot(),
		Code:     s.Code.Snapshot(),
		State:    s.Storage.Snapshot(),
		TXs:      s.Transactions.Snapshot(),
		RXs:      s.Receipts.Snapshot(),
		LogLen:   len(s.Logs),
	}
	s.Snapshots = append(s.Snapshots, snap)
	return len(s.Snapshots) - 1
}

// atSnap returns a copy of the state at the given snapshot
func (s *State) atSnap(n int, dst *State) {
	dst.Trace = s.Trace
	if n < 0 {
		return
	}
	ns := s.Snapshots[n]
	dst.Trace = s.Trace
	dst.Refund = s.Refund
	dst.Accounts = AccountTree{s.Accounts.CopyAt(ns.Accounts)}
	dst.Code = CodeTree{s.Code.CopyAt(ns.Code)}
	dst.Storage = s.Storage.CopyAt(ns.State)
	dst.Transactions = s.Transactions.CopyAt(ns.TXs)
	dst.Receipts = s.Receipts.CopyAt(ns.RXs)
	// prevent any updates to this new state
	// from clobbering the receiver
	dst.Logs = s.Logs[:ns.LogLen:ns.LogLen]
	dst.Snapshots = s.Snapshots[:n:n]
}

func (s *gethState) AddLog(l *types.Log) {
	if s.Trace != nil {
		s.Trace("AddLog", l)
	}
	s.Logs = append(s.Logs, l)
}

func (s *gethState) AddPreimage(h common.Hash, b []byte) {
	if s.Trace != nil {
		s.Trace("AddPreimage", h.String(), b)
	}
	s.Preimage.Insert(h[:], b)
}

func (s *gethState) ForEachStorage(addr common.Address, fn func(a, v common.Hash) bool) {
	if s.Trace != nil {
		s.Trace("ForEachStorage", addr.String())
	}
	panic("ForEachStorage not implemented")
}

// A Chain is a model of the state of the blockchain. The fields in this type
// are not threadsafe and must not be accessed concurrently. The methods on
// this type are threadsafe.
type Chain struct {
	State      State
	block2snap map[int64]int
	mu         sync.Mutex
}

// AtBlock returns the chain state at a given
// block number. As a special case, -1 is interpreted
// as the pending block (i.e. the current chain state),
// and -2 is interpreted as the latest block (i.e. the
// chain state just before the pending block).
func (c *Chain) AtBlock(n int64) *Chain {
	var snap int
	pending := int64(*c.State.Pending.Number)
	switch n {
	case pending - 1, -2: // latest
		s, ok := c.block2snap[pending-1]
		if ok {
			n = pending - 1
			snap = s
			break
		}
		fallthrough
	case pending, -1: // pending
		n = pending
		return c
	default:
		s, ok := c.block2snap[n]
		if !ok {
			return nil
		}
		snap = s
	}

	h := seth.Hash(n2h(uint64(n)))
	buf := c.State.Blocks.Get(h[:])
	if buf == nil {
		return nil
	}

	nb := new(seth.Block)
	if _, err := nb.UnmarshalMsg(buf); err != nil {
		panic(err)
	}

	cc := new(Chain)
	c.State.atSnap(snap, &cc.State)
	cc.State.Pending = nb
	cc.block2snap = c.block2snap
	return cc
}

func l2l(l *types.Log, sl *seth.Log) {
	sl.Address = seth.Address(l.Address)
	sl.Topics = make([]seth.Data, len(l.Topics))
	for i := range l.Topics {
		sl.Topics[i] = seth.Data(l.Topics[i][:])
	}
	sl.Data = seth.Data(l.Data)
	sl.BlockHash = (*seth.Hash)(&l.BlockHash)
	sl.TxHash = (*seth.Hash)(&l.TxHash)
	index := seth.Uint64(l.Index)
	sl.LogIndex = &index
	txindex := seth.Uint64(l.TxIndex)
	sl.TxIndex = &txindex
	bn := seth.Uint64(l.BlockNumber)
	sl.BlockNumber = &bn
	sl.Removed = l.Removed
}

func lconv(l []*types.Log) []seth.Log {
	out := make([]seth.Log, len(l))
	for i := range l {
		l2l(l[i], &out[i])
	}
	return out
}

func (c *Chain) Logs() []seth.Log {
	return lconv(c.State.Logs)
}

func n2h(u uint64) common.Hash {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], u)
	return common.Hash(seth.HashBytes(buf[:]))
}

const (
	defaultBlock      = 100
	defaultBlockTime  = 30
	defaultGasPrice   = 4000000000 // 4 gwei
	defaultGasLimit   = 6000000
	defaultDifficulty = 100
)

func now() uint64 {
	return uint64(time.Now().Unix())
}

// NewChain creates a new fake blockchain.
// In its initial state, the chain has no accounts
// with non-zero balances, and no deployed contracts.
func NewChain() *Chain {
	n := seth.Uint64(defaultBlock)
	h := seth.Hash(n2h(defaultBlock))
	c := &Chain{
		block2snap: make(map[int64]int),
		State: State{
			Pending: &seth.Block{
				Number:     &n,
				Hash:       &h,
				Timestamp:  seth.Uint64(time.Now().Unix()),
				GasLimit:   seth.Uint64(defaultGasLimit),
				Difficulty: seth.NewInt(defaultDifficulty),
			},
		},
	}
	return c
}

// NewAccount creates a new account with some ether in it.
// The balance of the new account will be 'ether' * 10**18
func (c *Chain) NewAccount(ether int) seth.Address {
	var addr seth.Address
	rand.Read(addr[:])
	if ether == 0 {
		c.State.StateDB().CreateAccount(common.Address(addr))
		return addr
	}
	var b big.Int
	b.SetInt64(int64(ether))
	var mul big.Int
	var et big.Int
	et.SetInt64(18)
	mul.SetInt64(10)
	mul.Exp(&mul, &et, nil)
	b.Mul(&b, &mul)

	var acct Account
	acct.SetBalance(&b)
	c.State.Accounts.SetAccount(&addr, &acct)
	return addr
}

func cantransfer(s vm.StateDB, addr common.Address, v *big.Int) bool {
	return s.GetBalance(addr).Cmp(v) >= 0
}

func dotransfer(s vm.StateDB, from, to common.Address, v *big.Int) {
	st := s.(*gethState)
	if st.Trace != nil {
		st.Trace("Transfer", from.String(), to.String(), v.String())
	}
	if v.Sign() == 0 {
		return
	}

	aaddr, baddr := seth.Address(from), seth.Address(to)
	facct, _ := st.Accounts.GetAccount(&aaddr)
	fbcct, _ := st.Accounts.GetAccount(&baddr)

	var ov big.Int
	fb, tb := facct.Balance(), fbcct.Balance()
	fbb, tbb := fb.Big(), tb.Big()

	ov.Set(v)
	fbb.Sub(fbb, v)
	tbb.Add(tbb, &ov)

	facct.SetBalance(fbb)
	fbcct.SetBalance(tbb)

	st.Accounts.SetAccount(&aaddr, &facct)
	st.Accounts.SetAccount(&baddr, &fbcct)
}

func (c *Chain) context(sender [20]byte) vm.Context {
	b := c.State.Pending
	return vm.Context{
		CanTransfer: cantransfer,
		Transfer:    dotransfer,
		GetHash:     n2h,
		Origin:      common.Address(sender),
		Coinbase:    common.Address(b.Miner),
		GasLimit:    uint64(b.GasLimit),
		BlockNumber: new(big.Int).SetInt64(int64(*b.Number)),
		Time:        new(big.Int).SetInt64(int64(b.Timestamp)),
		Difficulty:  new(big.Int).Set((*big.Int)(b.Difficulty)),
	}
}

type acctref seth.Address

var zero big.Int

func (a *acctref) Address() common.Address { return common.Address(*a) }

func s2r(sender *seth.Address) vm.ContractRef {
	return (*acctref)(sender)
}

func (c *Chain) evm(sender [20]byte) *vm.EVM {
	return vm.NewEVM(c.context(sender), c.State.StateDB(), &theparams, theconfig)
}

// Create executes a transation that deploys the given
// code to a new contract address, and returns the address
// of the newly created contract.
func (c *Chain) Create(sender *seth.Address, code []byte) (seth.Address, error) {
	c.mu.Lock()
	_, addr, _, err := c.evm(*sender).Create(s2r(sender), code, defaultGasLimit, &zero)
	c.mu.Unlock()
	return seth.Address(addr), err
}

// CreateAt creates a new contract at the given address. This does not do
// bookkeeping in the same way that Create does. In particular, it does not
// increment the sender nonce or enforce callstack limits.
func (c *Chain) CreateAt(addr, sender *seth.Address, code []byte) error {
	c.mu.Lock()

	evm := c.evm(*sender)
	evm.StateDB.CreateAccount(common.Address(*addr))
	evm.StateDB.SetNonce(common.Address(*addr), 1)

	contract := vm.NewContract(s2r(sender), s2r(addr), &zero, defaultGasLimit)
	contract.Code = code
	contract.CodeHash = crypto.Keccak256Hash(code)
	contract.CodeAddr = (*common.Address)(addr)

	ret, err := evm.Interpreter().Run(contract, nil)
	if err != nil {
		c.mu.Unlock()
		return err
	}

	evm.StateDB.SetCode(common.Address(*addr), ret)

	c.mu.Unlock()

	return nil
}

// Call executes a transaction that represents
// a call initiated by 'sender' to the destination
// address.
//
// 'sig' must be in the canonical method signature encoding.
func (c *Chain) Call(sender, dst *seth.Address, sig string, args ...seth.EtherType) ([]byte, error) {
	c.mu.Lock()
	ret, _, err := c.evm(*sender).Call(s2r(sender), common.Address(*dst), seth.ABIEncode(sig, args...), defaultGasLimit, &zero)
	c.mu.Unlock()
	return ret, err
}

// StaticCall yields the result of the given transaction in
// the pending block without comitting the state changes to the chain.
func (c *Chain) StaticCall(sender, dst *seth.Address, sig string, args ...seth.EtherType) ([]byte, error) {
	c.mu.Lock()
	ret, _, err := c.evm(*sender).StaticCall(s2r(sender), common.Address(*dst), seth.ABIEncode(sig, args...), defaultGasLimit)
	c.mu.Unlock()
	return ret, err
}

// Send creates a transaction that sends ether from one address to another.
func (c *Chain) Send(sender, dst *seth.Address, value *big.Int) error {
	c.mu.Lock()
	_, _, err := c.evm(*sender).Call(s2r(sender), common.Address(*dst), nil, defaultGasLimit, value)
	c.mu.Unlock()
	return err
}

// Client creates a seth.Client that talks to
// the fake chain. The client can be used to test
// unmodified code using the seth library against
// the mock chain.
func (c *Chain) Client() *seth.Client {
	return seth.NewClientTransport(c)
}

// Sender creates a Sender from a sending address.
// This can be used to test unmodified Go code using
// the seth library against a synthetic blockchain.
func (c *Chain) Sender(from *seth.Address) *seth.Sender {
	return seth.NewSender(c.Client(), from)
}

// SubBalance subtracts from the balance of an account.
func (c *Chain) SubBalance(addr *seth.Address, v *big.Int) {
	c.mu.Lock()
	c.State.StateDB().SubBalance(common.Address(*addr), v)
	c.mu.Unlock()
}

// AddBalance adds to the balance of an account.
func (c *Chain) AddBalance(addr *seth.Address, v *big.Int) {
	c.mu.Lock()
	c.State.StateDB().AddBalance(common.Address(*addr), v)
	c.mu.Unlock()
}

func (c *Chain) balanceOf(addr *seth.Address) *big.Int {
	acct, _ := c.State.Accounts.GetAccount(addr)
	bal := acct.Balance()
	return bal.Big()
}

// BalanceOf returns the balance of an address, in Wei.
func (c *Chain) BalanceOf(addr *seth.Address) *big.Int {
	c.mu.Lock()
	b := c.balanceOf(addr)
	c.mu.Unlock()
	return b
}

func encode(v msgp.Marshaler) []byte {
	b, err := v.MarshalMsg(nil)
	if err != nil {
		panic(err)
	}
	return b
}

// Mine executes a transaction and returns
// the return value of the transaction (if any) and the
// transaction hash. Unlike the other methods of executing
// a transaction on a Chain, this method updates the pending
// block and saves the transaction and its receipt in the state
// tree so that they can be retrieved later. Additionally,
// this method respects the amount of gas sent in the transaction,
// rather than offering all of the gas in the block to the transaction,
// which more faithfully mimics the behavior of an actual ethereum node.
func (c *Chain) Mine(tx *seth.Transaction) (ret []byte, h seth.Hash, err error) {
	l0 := len(c.State.Logs)

	var gas uint64
	var addr common.Address
	status := 1
	vm := c.evm(*tx.From)
	if tx.To == nil {
		ret, addr, gas, err = vm.Create(s2r(tx.From), []byte(tx.Input), uint64(tx.Gas), tx.Value.Big())
	} else {
		ret, gas, err = vm.Call(s2r(tx.From), common.Address(*tx.To), []byte(tx.Input), uint64(tx.Gas), tx.Value.Big())
	}

	if err != nil {
		status = 0
	}

	used := uint64(tx.Gas) - gas
	b := c.State.Pending
	b.GasUsed += seth.Uint64(used)
	idx := new(seth.Uint64)
	*idx = seth.Uint64(len(b.Transactions))
	tx.TxIndex = idx
	tx.Block = *b.Hash
	tx.BlockNumber = *b.Number

	// tx hash is txidx as high 16 bits and block number as low 48 bits
	bh := n2h(uint64(*b.Number) | (uint64(len(b.Transactions)) << 48))
	tx.Hash = seth.HashBytes(bh[:])
	h = tx.Hash

	rx := &seth.Receipt{
		Hash:       tx.Hash,
		Index:      *tx.TxIndex,
		GasUsed:    seth.Uint64(used),
		Cumulative: b.GasUsed,
		Logs:       lconv(c.State.Logs[l0:]),
		Status:     seth.Uint64(status),
	}
	if tx.To == nil {
		rx.Address = new(seth.Address)
		copy(rx.Address[:], addr[:])
	}
	b.Transactions = append(b.Transactions, js(&tx.Hash))
	c.State.Transactions.Insert(tx.Hash[:], encode(tx))
	c.State.Receipts.Insert(rx.Hash[:], encode(rx))
	return
}

func js(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// Seal seals the current block (c.Pending) and
// replaces it with a new pending block with the
// same parameters (but with an update block number and hash,
// and zeroed gas used).
func (c *Chain) Seal() {
	b := c.State.Pending

	// seal the current state
	c.block2snap[int64(*b.Number)] = (*gethState)(&c.State).Snapshot()

	c.State.Blocks.Insert(b.Hash[:], encode(b))
	n := seth.Uint64(uint64(*b.Number) + 1)
	h := seth.Hash(n2h(uint64(n)))
	c.State.Pending = &seth.Block{
		Number:          &n,
		Parent:          *b.Hash,
		Hash:            &h,
		GasLimit:        b.GasLimit,
		Difficulty:      seth.NewInt(0),
		TotalDifficulty: seth.NewInt(0),
		Timestamp:       seth.Uint64(time.Now().Unix()),
	}
}

// MarshalJSON implements json.Marshaler.
func (c *Chain) MarshalJSON() ([]byte, error) {
	c.mu.Lock()
	b, err := json.Marshal(&struct {
		State      State
		Block2snap map[int64]int
	}{c.State, c.block2snap})
	c.mu.Unlock()
	return b, err
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Chain) UnmarshalJSON(b []byte) error {
	var s struct {
		State      State
		Block2snap map[int64]int
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	c.mu.Lock()
	c.State = s.State
	c.block2snap = s.Block2snap
	c.mu.Unlock()
	return nil
}
