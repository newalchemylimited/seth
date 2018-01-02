package seth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"reflect"
	"time"
	"unsafe"

	"github.com/newalchemylimited/seth/keccak"
	"github.com/tinylib/msgp/msgp"
)

// HashString produces a Hash that corresponds
// to the Keccak-256 hash of the given string
func HashString(s string) Hash {
	b := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: (*reflect.StringHeader)(unsafe.Pointer(&s)).Data,
		Len:  len(s),
		Cap:  len(s),
	}))
	return HashBytes(b)
}

// HashBytes produces the Keccak-256 hash of the given bytes
func HashBytes(b []byte) Hash {
	return Hash(keccak.Sum256(b))
}

var (
	// ERC20Transfer is hash of the canonical ERC20 Transfer event
	ERC20Transfer = HashString("Transfer(address,address,uint256)")
	// ERC20Approve is the hash of the canonical ERC20 Approve event
	ERC20Approve = HashString("Approval(address,address,uint256)")
)

type Transport interface {
	Execute(req *RPCRequest, res *RPCResponse) error
}

type Client struct {
	tport  Transport
	nextid uintptr
}

func NewClient(dial func() (io.ReadWriteCloser, error)) *Client {
	tp := &RPCTransport{
		dial:    dial,
		pending: make(map[int]*pending),
	}
	return NewClientTransport(tp)
}

func NewHTTPClient(url string) *Client {
	return NewClientTransport(&HTTPTransport{URL: url})
}

func NewClientTransport(tp Transport) *Client {
	return &Client{tport: tp}
}

// Data is just binary that can decode Ethereum's silly quoted hex
type Data []byte

func (d *Data) String() string {
	if d == nil {
		return "<nil>"
	}
	return string(hexstring(*d, false))
}

func (d Data) MarshalText() ([]byte, error) {
	return hexstring(d, false), nil
}

func (d *Data) UnmarshalText(b []byte) error {
	s, err := hexparse(b)
	if err != nil {
		return err
	}
	*d = s
	return nil
}

// Bytes is like Data, but can be be greater than 32 bytes in length
type Bytes []byte

func (b *Bytes) String() string {
	if b == nil {
		return "<nil>"
	}
	return string(hexstring(*b, false))
}

func (b Bytes) MarshalText() ([]byte, error) {
	return hexstring(b, false), nil
}

func (b *Bytes) UnmarshalText(bx []byte) error {
	s, err := hexparse(bx)
	if err != nil {
		return err
	}
	*b = s
	return nil
}

// Int is a big.Int that can decode Ethereum's silly quoted hex
type Int big.Int

// NewInt allocates and returns a new Int set to x.
func NewInt(x int64) *Int { return (*Int)(big.NewInt(x)) }

// Big is a convenience method for (*big.Int)(i).
func (i *Int) Big() *big.Int {
	return (*big.Int)(i)
}

// Scan implements fmt.Scanner
func (i *Int) Scan(s fmt.ScanState, verb rune) error {
	// don't let a leading '0x' cause %x to silently
	// return 0; use a more generic prefix and let
	// (*big.Int).Scan figure it out
	if verb == 'x' {
		verb = 'v'
	}
	return i.Big().Scan(s, verb)
}

func (i *Int) IsZero() bool {
	return i.Big().Sign() == 0
}

func (i *Int) Int64() int64 {
	return i.Big().Int64()
}

func (i *Int) SetInt64(v int64) {
	i.Big().SetInt64(v)
}

func (i *Int) Uint64() uint64 {
	return i.Big().Uint64()
}

func (i *Int) SetUint64(v uint64) {
	i.Big().SetUint64(v)
}

func (i *Int) Copy() Int {
	var out big.Int
	out.Set((*big.Int)(i))
	return Int(out)
}

func (i *Int) EncodeMsg(w *msgp.Writer) error {
	return w.WriteBytes(i.Big().Bytes())
}

func (i *Int) DecodeMsg(r *msgp.Reader) error {
	buf, err := r.ReadBytes(nil)
	if err != nil {
		return err
	}
	i.Big().SetBytes(buf)
	return nil
}

func (i *Int) MarshalMsg(b []byte) ([]byte, error) {
	return msgp.AppendBytes(b, i.Big().Bytes()), nil
}

func (i *Int) UnmarshalMsg(b []byte) ([]byte, error) {
	buf, rest, err := msgp.ReadBytesBytes(b, nil)
	if err != nil {
		return rest, err
	}
	i.Big().SetBytes(buf)
	return rest, nil
}

func (i *Int) String() string {
	if i == nil {
		return "<nil>"
	}
	return string(hexstring(i.Big().Bytes(), true))
}

// MarshalText implements encoding.TextMarshaler.
func (i Int) MarshalText() ([]byte, error) {
	return hexstring(i.Big().Bytes(), true), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *Int) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, rawnull) {
		return nil
	}
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return i.UnmarshalText(b[1 : len(b)-1])
	}
	return i.UnmarshalText(b)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (i *Int) UnmarshalText(b []byte) error {
	if !hexprefix(b) {
		return i.Big().UnmarshalText(b)
	}
	buf, err := hexparse(b)
	if err != nil {
		return err
	}
	i.Big().SetBytes(buf)
	return nil
}

func (i *Int) FromString(s string) error {
	return i.UnmarshalText([]byte(s))
}

func (i *Int) Msgsize() int {
	n := (i.Big().BitLen() + 7) / 8
	return msgp.BytesPrefixSize + n
}

// Uint64 is a uint64 that marshals as a hex-encoded number.
type Uint64 uint64

// MarshalText implements encoding.TextMarshaler.
func (i Uint64) MarshalText() ([]byte, error) {
	var n Int
	n.SetUint64(uint64(i))
	return n.MarshalText()
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *Uint64) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, rawnull) {
		return nil
	}
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return i.UnmarshalText(b[1 : len(b)-1])
	}
	return i.UnmarshalText(b)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (i *Uint64) UnmarshalText(b []byte) error {
	var n Int
	if err := n.UnmarshalText(b); err != nil {
		return err
	}
	if !n.Big().IsUint64() {
		return errors.New("overflow: " + string(b))
	}
	*i = Uint64(n.Uint64())
	return nil
}

//go:generate msgp

//msgp:shim json.RawMessage as:string using:string/json.RawMessage

// Address represent an Ethereum address
type Address [20]byte

// ParseAddress parses an address.
func ParseAddress(s string) (*Address, error) {
	a := new(Address)
	if err := a.FromString(s); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Address) String() string {
	return string(hexstring(a[:], false))
}

func (a *Address) FromString(s string) error {
	return hexdecode(a[:], []byte(s))
}

// Scan implements fmt.Scanner (uses the verb %a)
func (a *Address) Scan(s fmt.ScanState, x rune) error {
	if x != 'a' {
		return fmt.Errorf("rune %q not valid verb for address", x)
	}
	tok, err := s.Token(false, nil)
	if err != nil {
		return err
	}
	return hexdecode(a[:], tok)
}

func (a Address) MarshalText() ([]byte, error) {
	return hexstring(a[:], false), nil
}

func (a *Address) UnmarshalText(b []byte) error {
	return hexdecode(a[:], b)
}

// Zero returns whether this is the zero address.
func (a *Address) Zero() bool {
	return a == nil || *a == Address{}
}

// Hash represents a Keccak256 hash
type Hash [32]byte

// ParseHash parses a hash.
func ParseHash(s string) (*Hash, error) {
	h := new(Hash)
	if err := h.FromString(s); err != nil {
		return nil, err
	}
	return h, nil
}

// String produces the hash as a 0x-prefixed hex string
func (h *Hash) String() string {
	return string(hexstring(h[:], false))
}

func (h *Hash) FromString(s string) error {
	return hexdecode(h[:], []byte(s))
}

func (h Hash) MarshalText() ([]byte, error) {
	return hexstring(h[:], false), nil
}

func (h *Hash) UnmarshalText(b []byte) error {
	return hexdecode(h[:], b)
}

// Scan implements fmt.Scanner (uses %h verb)
func (h *Hash) Scan(s fmt.ScanState, verb rune) error {
	if verb != 'h' {
		return fmt.Errorf("verb %q invalid for hash", verb)
	}
	tok, err := s.Token(false, nil)
	if err != nil {
		return err
	}
	return hexdecode(h[:], tok)
}

// Block represents and Ethereum block
type Block struct {
	Number          *Uint64           `json:"number"`     // block number, or nil if pending
	Hash            *Hash             `json:"hash"`       // block hash, or nil if pending
	Parent          Hash              `json:"parentHash"` // parent block hash
	Nonce           Uint64            `json:"nonce"`
	UncleHash       Hash              `json:"sha3Uncles"`          // hash of uncles in block
	Bloom           Data              `json:"logsBloom,omitempty"` // bloom filter of logs, or nil if pending
	TxRoot          Hash              `json:"transactionsRoot"`    // root of transaction trie of block
	StateRoot       Hash              `json:"stateRoot"`           // root of final state trie of block
	ReceiptsRoot    Hash              `json:"receiptsRoot"`        // root of receipts trie of block
	Miner           Address           `json:"miner"`
	GasLimit        Uint64            `json:"gasLimit"`
	GasUsed         Uint64            `json:"gasUsed"`
	Transactions    []json.RawMessage `json:"transactions"` // transactions; either hex strings of hashes, or actual tx bodies
	Uncles          []Hash            `json:"uncles"`       // array of uncle hashes
	Difficulty      *Int              `json:"difficulty"`
	TotalDifficulty *Int              `json:"totalDifficulty"`
	Timestamp       Uint64            `json:"timestamp"`
	Extra           Data              `json:"extraData,omitempty"`
}

// Time turns the block timestamp into a time.Time
func (b *Block) Time() time.Time {
	return time.Unix(int64(b.Timestamp), 0)
}

// Transactions returns the list of block transactions, given
// that b.Transactions is a set of serialized transactions, and
// not just a set of tx hashes.
func (b *Block) ParseTransactions() ([]Transaction, error) {
	out := make([]Transaction, len(b.Transactions))
	for i := range b.Transactions {
		err := json.Unmarshal(b.Transactions[i], &out[i])
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Transaction represents an ethereum transaction
type Transaction struct {
	Hash        Hash     `json:"hash"`             // tx hash
	Nonce       Uint64   `json:"nonce"`            // sender nonce
	Block       Hash     `json:"blockHash"`        // hash of parent block
	BlockNumber Uint64   `json:"blockNumber"`      //
	To          *Address `json:"to"`               // receiver, or nil for contract creation
	TxIndex     *Uint64  `json:"transactionIndex"` // transaction index, or nil if pending
	From        *Address `json:"from"`             // from
	Value       Int      `json:"value"`            // value in wei
	GasPrice    Int      `json:"gasPrice"`         // gas price
	Gas         Uint64   `json:"gas"`              // gas spent on transaction
	Input       Data     `json:"input"`            // input data
}

// RPCRequest is a request to be sent to an RPC server.
type RPCRequest struct {
	Version string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
	ID      int               `json:"id"`
}

// RPCError is an error returned by a server
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("%s (code %d) %s", e.Message, e.Code, e.Data)
}

// RPCResponse is a response returned by an RPC server.
type RPCResponse struct {
	ID      int             `json:"id"`
	Version string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   RPCError        `json:"error"`
}

// Pending returns the list of pending transactions.
func (c *Client) Pending() ([]Transaction, error) {
	b, err := c.GetBlock(-1, true)
	if err != nil {
		return nil, err
	}
	return b.ParseTransactions()
}

var rawtrue = json.RawMessage("true")
var rawfalse = json.RawMessage("false")
var rawpending = json.RawMessage(`"pending"`)
var rawlatest = json.RawMessage(`"latest"`)
var rawnull = json.RawMessage("null")

// GetBalance gets the balance of an address in wei at the latest block.
func (c *Client) GetBalance(addr *Address) (Int, error) {
	params := []json.RawMessage{nil, rawlatest}
	buf, _ := json.Marshal(addr)
	params[0] = json.RawMessage(buf)
	wei := Int{}
	err := c.Do("eth_getBalance", params, &wei)
	return wei, err
}

// GetBlock gets a block by block number. If 'txs' is true,
// the block includes all the transactions in the block; otherwise
// it only includes the transaction hashes. As a special case,
// a negative block number is interpreted to mean the pending block.
func (c *Client) GetBlock(num int64, txs bool) (*Block, error) {
	params := make([]json.RawMessage, 2)
	if num >= 0 {
		buf, _ := json.Marshal((*Int)(big.NewInt(num)))
		params[0] = json.RawMessage(buf)
	} else {
		params[0] = rawpending
	}
	if txs {
		params[1] = rawtrue
	} else {
		params[1] = rawfalse
	}
	out := Block{}
	err := c.Do("eth_getBlockByNumber", params, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTransaction gets a transaction by its hash
func (c *Client) GetTransaction(h *Hash) (*Transaction, error) {
	buf, _ := json.Marshal(h)
	o := new(Transaction)
	err := c.Do("eth_getTransactionByHash", []json.RawMessage{buf}, o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// Latest returns the latest block
func (c *Client) Latest(txs bool) (*Block, error) {
	out := Block{}
	params := make([]json.RawMessage, 2)
	params[0] = rawlatest
	if txs {
		params[1] = rawtrue
	} else {
		params[1] = rawfalse
	}
	err := c.Do("eth_getBlockByNumber", params, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// BlockIterator manages a channel that
// yields blocks in block number order.
type BlockIterator struct {
	c    *Client
	out  chan *Block
	done chan struct{}
}

// Stop causes the block itertation to stop.
// It should only be called once.
func (b *BlockIterator) Stop() {
	close(b.done)
}

func (b *BlockIterator) getLoop(block int64, txs bool) {
	for {
		select {
		case <-b.done:
			close(b.out)
			return
		default:
			v, err := b.c.GetBlock(block, txs)
			if err != nil {
				if err != ErrNotFound {
					log.Printf("error getting block %d: %s", block, err)
				}
				time.Sleep(1 * time.Second)
				continue
			}
			b.out <- v
			block++
		}
	}
}

// Next returns the next block in the chain. The channel will
// be closed when Stop() is called.
func (b *BlockIterator) Next() <-chan *Block { return b.out }

// IterateBlocks creates a BlockIterator that starts at the
// given block number.
func (c *Client) IterateBlocks(from int64, txs bool) *BlockIterator {
	b := &BlockIterator{c: c, out: make(chan *Block, 64), done: make(chan struct{})}
	go b.getLoop(from, txs)
	return b
}

// Receipt is a transaction receipt
type Receipt struct {
	Hash        Hash     `json:"transactionHash"`
	Index       Uint64   `json:"transactionIndex"`
	BlockHash   Hash     `json:"blockHash"`
	BlockNumber Uint64   `json:"blockNumber"`
	GasUsed     Uint64   `json:"gasUsed"`
	Cumulative  Uint64   `json:"cumulativeGasUsed"`
	Address     *Address `json:"contractAddress"` // contract created, or none if not a contract creation
	Status      Uint64   `json:"status"`
	Logs        []Log    `json:"logs"`
}

// Threw returns whether the transaction threw.
func (r *Receipt) Threw() bool {
	return r.Status == 0
}

func (c *Client) GetCode(addr *Address) ([]byte, error) {
	buf, _ := json.Marshal(addr)
	var out Data
	err := c.Do("eth_getCode", []json.RawMessage{buf, rawlatest}, &out)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

// GetReceipt gets a receipt for a given transaction hash.
func (c *Client) GetReceipt(tx *Hash) (*Receipt, error) {
	buf, _ := json.Marshal(tx)
	out := &Receipt{}
	err := c.Do("eth_getTransactionReceipt", []json.RawMessage{buf}, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Log is an Ethereum log (or, in Solidity, an "event")
type Log struct {
	Removed     bool    `json:"removed"`
	LogIndex    *Uint64 `json:"logIndex"` // nil if pending; same for following fields
	TxIndex     *Uint64 `json:"transactionIndex"`
	TxHash      *Hash   `json:"transactionHash"`
	BlockHash   *Hash   `json:"blockHash"`
	BlockNumber *Uint64 `json:"blockNumber"`
	Address     Address `json:"address"`
	Data        Data    `json:"data"`   // serialized log arguments
	Topics      []Data  `json:"topics"` // indexed log arguments
}

// TokenTransfer represents an ERC20 token transfer event
type TokenTransfer struct {
	Block    int64   // block number
	TxHeight int     // index of transaction in block
	Token    Address // address of contract
	From     Address // 'from' argument in transfer
	To       Address // 'to' argument in transfer
	Amount   Int     // value amount
}

func flatten(i *Int) int64 {
	return (*big.Int)(i).Int64()
}

func setdata(i *Int, b []byte) {
	(*big.Int)(i).SetBytes(b)
}

// ParseTransfer tries to parse this log as
// an ERC20 token transfer event, with the signature
//    event Transfer(address indexed from, address indexed to, uint256 value);
func (r *Receipt) ParseTransfer(l *Log) (TokenTransfer, bool) {
	if len(l.Topics) != 3 || len(l.Topics[1]) != 32 || len(l.Topics[2]) != 32 {
		return TokenTransfer{}, false
	}
	if !bytes.Equal(l.Topics[0], ERC20Transfer[:]) {
		return TokenTransfer{}, false
	}
	tt := TokenTransfer{Block: int64(r.BlockNumber), TxHeight: int(r.Index), Token: l.Address}
	copy(tt.From[:], l.Topics[1][12:]) // these are addresses; first 12 bytes should be zeros
	copy(tt.To[:], l.Topics[2][12:])
	setdata(&tt.Amount, l.Data)
	return tt, true
}
