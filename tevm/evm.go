package tevm

import (
	"encoding/binary"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/newalchemylimited/seth"
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

// State implements the garbage vm.StateDB interface
type State struct {
	// hide the implementation of geth's vm.StateDB
	// so that we don't leak all of these types into
	// the documentation
	impl gethState

	refund   seth.Int
	Trace    func(fn string, args ...interface{})
	Accounts AccountTree
	Code     CodeTree
	State    Tree // key = hash(address, pointer)
	Preimage Tree

	logs      []*types.Log
	snapshots []statesnap
}

type gethState struct {
	*State
}

type statesnap struct {
	refund   seth.Int
	accounts int
	code     int
	state    int
	loglen   int
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

func (s *gethState) AddRefund(v *big.Int) {
	b := (*big.Int)(&s.refund)
	b.Add(b, v)
}

func (s *gethState) GetRefund() *big.Int {
	return (*big.Int)(&s.refund)
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
	v := s.State.State.Get(h[:])
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
		s.State.State.Delete(h[:])
	} else {
		s.State.State.Insert(h[:], value[:])
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
	snaps := s.snapshots
	if len(snaps) <= v || v < 0 {
		panic("no such snapshot")
	}
	ns := snaps[v]
	s.refund = ns.refund
	s.Accounts.Rollback(ns.accounts)
	s.Code.Rollback(ns.code)
	s.State.State.Rollback(ns.state)
	s.logs = s.logs[:ns.loglen]

	// make sure we can't roll forward
	snaps = snaps[:v-1]
}

func (s *gethState) Snapshot() int {
	if s.Trace != nil {
		s.Trace("Snapshot")
	}
	snap := statesnap{
		refund:   s.refund.Copy(),
		accounts: s.Accounts.Snapshot(),
		code:     s.Code.Snapshot(),
		state:    s.State.State.Snapshot(),
		loglen:   len(s.logs),
	}
	s.snapshots = append(s.snapshots, snap)
	return len(s.snapshots) - 1
}

func (s *gethState) AddLog(l *types.Log) {
	if s.Trace != nil {
		s.Trace("AddLog", l)
	}
	s.logs = append(s.logs, l)
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

type Chain struct {
	State State
	vm    *vm.EVM

	transactions map[seth.Hash]*Transaction
	receipts     map[seth.Hash]*types.Receipt
	mu           sync.Mutex

	// Block is the pending block in the chain.
	Block *Block

	// Blocks are previous blocks in the chain.
	Blocks map[int64]*Block
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
	sl.LogIndex = seth.NewInt(int64(l.Index))
	sl.TxIndex = seth.NewInt(int64(l.TxIndex))
	sl.BlockNumber = seth.NewInt(int64(l.BlockNumber))
	sl.Removed = l.Removed
}

func (c *Chain) Logs() []seth.Log {
	out := make([]seth.Log, len(c.State.logs))
	for i := range c.State.logs {
		l2l(c.State.logs[i], &out[i])
	}
	return out
}

type Block struct {
	Coinbase     seth.Address
	Number       int64
	Time         uint64
	GasPrice     int64
	GasLimit     int64
	GasUsed      int64
	Transactions []seth.Hash
	Difficulty   int64
}

func (b *Block) Hash() seth.Hash {
	return seth.Hash(n2h(uint64(b.Number)))
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

func NewChain() *Chain {
	c := &Chain{
		transactions: make(map[seth.Hash]*Transaction),
		receipts:     make(map[seth.Hash]*types.Receipt),
		Blocks:       make(map[int64]*Block),
		Block: &Block{
			Time:       now(),
			GasPrice:   defaultGasPrice,
			GasLimit:   defaultGasLimit,
			Difficulty: defaultDifficulty,
		},
	}
	c.State.impl.State = &c.State
	return c
}

// NewAccount creates a new account with some ether in it
func (c *Chain) NewAccount(ether int) seth.Address {
	var addr seth.Address
	rand.Read(addr[:])
	if ether == 0 {
		c.State.impl.CreateAccount(common.Address(addr))
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
	st := s.(*gethState).State
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
	return vm.Context{
		CanTransfer: cantransfer,
		Transfer:    dotransfer,
		GetHash:     n2h,
		Origin:      common.Address(sender),
		GasPrice:    new(big.Int).SetInt64(c.Block.GasPrice),
		Coinbase:    common.Address(c.Block.Coinbase),
		GasLimit:    new(big.Int).SetInt64(c.Block.GasLimit),
		BlockNumber: new(big.Int).SetInt64(c.Block.Number),
		Time:        new(big.Int).SetInt64(int64(c.Block.Time)),
		Difficulty:  new(big.Int).SetInt64(c.Block.Difficulty),
	}
}

type acctref seth.Address

var zero big.Int

func (a *acctref) Address() common.Address { return common.Address(*a) }

func s2r(sender *seth.Address) vm.ContractRef {
	return (*acctref)(sender)
}

func (c *Chain) evm(sender [20]byte) *vm.EVM {
	return vm.NewEVM(c.context(sender), &c.State.impl, &theparams, theconfig)
}

func (c *Chain) Create(sender *seth.Address, code []byte) (seth.Address, error) {
	_, addr, _, err := c.evm(*sender).Create(s2r(sender), code, defaultGasLimit, &zero)
	return seth.Address(addr), err
}

func (c *Chain) Call(sender, dst *seth.Address, sig string, args ...seth.EtherType) ([]byte, error) {
	ret, _, err := c.evm(*sender).Call(s2r(sender), common.Address(*dst), seth.ABIEncode(sig, args...), defaultGasLimit, &zero)
	return ret, err
}

func (c *Chain) StaticCall(sender, dst *seth.Address, sig string, args ...seth.EtherType) ([]byte, error) {
	ret, _, err := c.evm(*sender).StaticCall(s2r(sender), common.Address(*dst), seth.ABIEncode(sig, args...), defaultGasLimit)
	return ret, err
}

func (c *Chain) Send(sender, dst *seth.Address, value *big.Int) error {
	_, _, err := c.evm(*sender).Call(s2r(sender), common.Address(*dst), nil, defaultGasLimit, value)
	return err
}

func (c *Chain) Client() *seth.Client {
	return seth.NewClientTransport(c)
}

func (c *Chain) Sender(from *seth.Address) *seth.Sender {
	return seth.NewSender(c.Client(), from)
}

func (c *Chain) BalanceOf(addr *seth.Address) *big.Int {
	acct, _ := c.State.Accounts.GetAccount(addr)
	bal := acct.Balance()
	return bal.Big()
}

// A Transaction as it appears in an RPC message.
type Transaction struct {
	BlockHash   seth.Hash     `json:"blockHash"`
	BlockNumber seth.Uint64   `json:"blockNumber"`
	From        seth.Address  `json:"from"`
	Gas         seth.Uint64   `json:"gas"`
	GasPrice    seth.Uint64   `json:"gasPrice"`
	Hash        seth.Hash     `json:"hash"`
	Input       seth.Data     `json:"input"`
	Nonce       seth.Uint64   `json:"nonce"`
	To          *seth.Address `json:"to"`
	Index       seth.Uint64   `json:"transactionIndex"`
	Value       seth.Int      `json:"value"`
	V           seth.Int      `json:"v"`
	R           seth.Int      `json:"r"`
	S           seth.Int      `json:"s"`
}
