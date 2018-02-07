package seth

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func unhex(t *testing.T, str string) []byte {
	b, err := hex.DecodeString(strings.Replace(str, " ", "", -1))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

type encodeTest struct {
	input  interface{}
	output string
	error  string
}

var encodeIntTests = []encodeTest{
	{input: uint64(0x00), output: "80"},
	{input: uint64(0x7F), output: "7F"},
	{input: uint64(0xFF), output: "81FF"},
	{input: uint64(0xFFF), output: "820FFF"},
	{input: uint64(0xFFFF), output: "82FFFF"},
	{input: uint64(0xFFFFF), output: "830FFFFF"},
	{input: uint64(0xFFFFFF), output: "83FFFFFF"},
	{input: uint64(0xFFFFFFF), output: "840FFFFFFF"},
	{input: uint64(0xFFFFFFFF), output: "84FFFFFFFF"},
}

func runEncTests(t *testing.T, d []encodeTest, f func(w io.Writer, val interface{})) {
	for i, test := range d {
		res := new(bytes.Buffer)

		f(res, test.input)

		if !bytes.Equal(res.Bytes(), unhex(t, test.output)) {
			t.Errorf("test %d: output mismatch:\ngot   %X\nwant  %s\nvalue %#v\ntype  %T",
				i, res.Bytes(), test.output, test.input, test.input)
		}
	}
}

func TestEncodeInt(t *testing.T) {
	runEncTests(t, encodeIntTests, func(w io.Writer, val interface{}) {
		rlpEncodeInt(w, Uint64(val.(uint64)))
		return
	})
}

var encodeBytesTests = []encodeTest{
	{input: []byte(""), output: "80"},
	{input: []byte{0x00}, output: "00"},
	{input: []byte("A"), output: "41"},
	{input: []byte("AA"), output: "824141"},
	{input: []byte("AAA"), output: "83414141"},
	{input: []byte("AAAA"), output: "8441414141"},
	{
		input:  []byte("aaabbbcccdddeeefffggghhhiiijjjkkklllmmmnnnooopppqqqrrrssstttuuuvvvwwwxxxyyyzzz"),
		output: "B84E6161616262626363636464646565656666666767676868686969696A6A6A6B6B6B6C6C6C6D6D6D6E6E6E6F6F6F7070707171717272727373737474747575757676767777777878787979797A7A7A",
	},
	{
		input: []byte("aaabbbcccdddeeefffggghhhiiijjjkkklllmmmnnnooopppqqqrrrssstttuuuvvvwwwxxxyyyzzz" +
			"aaabbbcccdddeeefffggghhhiiijjjkkklllmmmnnnooopppqqqrrrssstttuuuvvvwwwxxxyyyzzz" +
			"aaabbbcccdddeeefffggghhhiiijjjkkklllmmmnnnooopppqqqrrrssstttuuuvvvwwwxxxyyyzzz" +
			"aaabbbcccdddeeefffggghhhiiijjjkkklllmmmnnnooopppqqqrrrssstttuuuvvvwwwxxxyyyzzz" +
			"aaabbbcccdddeeefffggghhhiiijjjkkklllmmmnnnooopppqqqrrrssstttuuuvvvwwwxxxyyyzzz" +
			"aaabbbcccdddeeefffggghhhiiijjjkkklllmmmnnnooopppqqqrrrssstttuuuvvvwwwxxxyyyzzz" +
			"aaabbbcccdddeeefffggghhhiiijjjkkklllmmmnnnooopppqqqrrrssstttuuuvvvwwwxxxyyyzzz" +
			"aaabbbcccdddeeefffggghhhiiijjjkkklllmmmnnnooopppqqqrrrssstttuuuvvvwwwxxxyyyzzz" +
			"aaabbbcccdddeeefffggghhhiiijjjkkklllmmmnnnooopppqqqrrrssstttuuuvvvwwwxxxyyyzzz" +
			"aaabbbcccdddeeefffggghhhiiijjjkkklllmmmnnnooopppqqqrrrssstttuuuvvvwwwxxxyyyzzz"),

		output: "B9030C6161616262626363636464646565656666666767676868686969696A6A6A6B6B6B6C6C6C6D6D6D6E6E6E6F6F6F7070707171717272727373737474747575757676767777777878787979797A7A7A" +
			"6161616262626363636464646565656666666767676868686969696A6A6A6B6B6B6C6C6C6D6D6D6E6E6E6F6F6F7070707171717272727373737474747575757676767777777878787979797A7A7A" +
			"6161616262626363636464646565656666666767676868686969696A6A6A6B6B6B6C6C6C6D6D6D6E6E6E6F6F6F7070707171717272727373737474747575757676767777777878787979797A7A7A" +
			"6161616262626363636464646565656666666767676868686969696A6A6A6B6B6B6C6C6C6D6D6D6E6E6E6F6F6F7070707171717272727373737474747575757676767777777878787979797A7A7A" +
			"6161616262626363636464646565656666666767676868686969696A6A6A6B6B6B6C6C6C6D6D6D6E6E6E6F6F6F7070707171717272727373737474747575757676767777777878787979797A7A7A" +
			"6161616262626363636464646565656666666767676868686969696A6A6A6B6B6B6C6C6C6D6D6D6E6E6E6F6F6F7070707171717272727373737474747575757676767777777878787979797A7A7A" +
			"6161616262626363636464646565656666666767676868686969696A6A6A6B6B6B6C6C6C6D6D6D6E6E6E6F6F6F7070707171717272727373737474747575757676767777777878787979797A7A7A" +
			"6161616262626363636464646565656666666767676868686969696A6A6A6B6B6B6C6C6C6D6D6D6E6E6E6F6F6F7070707171717272727373737474747575757676767777777878787979797A7A7A" +
			"6161616262626363636464646565656666666767676868686969696A6A6A6B6B6B6C6C6C6D6D6D6E6E6E6F6F6F7070707171717272727373737474747575757676767777777878787979797A7A7A" +
			"6161616262626363636464646565656666666767676868686969696A6A6A6B6B6B6C6C6C6D6D6D6E6E6E6F6F6F7070707171717272727373737474747575757676767777777878787979797A7A7A",
	},
}

func TestEncodeBytes(t *testing.T) {
	runEncTests(t, encodeBytesTests, func(w io.Writer, val interface{}) {
		rlpEncodeString(w, val.([]byte))
		return
	})
}

type jstx struct {
	Data     string `json:"data"`
	GasPrice Int    `json:"gasprice"`
	Hex      string `json:"hex"`
	Nonce    uint64 `json:"nonce"`
	R        Int    `json:"r"`
	S        Int    `json:"s"`
	StartGas int64  `json:"startgas"`
	To       string `json:"to"`
	TxHash   string `json:"txhash"`
	V        byte   `json:"v"`
	Value    Int    `json:"value"`
}

type sTx struct {
	t *Transaction
	s *Signature
}

var signedTxTests []encodeTest
var unsignedTxTests []encodeTest

func signedTx(t *testing.T, p string) {
	if len(signedTxTests) > 0 {
		return
	}

	files, err := filepath.Glob(p)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		js, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		rtx := jstx{}
		if err = json.Unmarshal(js, &rtx); err != nil {
			t.Fatal(err)
		}

		addr, err := ParseAddress(rtx.To)
		if err != nil {
			addr = nil
		}

		var hash Hash
		copy(hash[:], unhex(t, rtx.TxHash))

		tx := Transaction{
			Hash:     hash,
			Nonce:    Uint64(rtx.Nonce),
			GasPrice: rtx.GasPrice,
			Gas:      Uint64(rtx.StartGas),
			Input:    unhex(t, rtx.Data), //dataBytes,
			To:       addr,
			TxIndex:  nil,
			From:     nil,
			Value:    rtx.Value,
		}

		signedTransaction := sTx{t: &tx, s: NewSignature(rtx.R.Big(), rtx.S.Big(), int(rtx.V)-37)}
		signedTxTests = append(signedTxTests, encodeTest{input: signedTransaction, output: rtx.Hex})
		unsignedTxTests = append(unsignedTxTests, encodeTest{input: tx, output: "3031"})
	}
}

func TestSignedTxMarshal(t *testing.T) {
	signedTx(t, "./_test/txs/*.json")

	runEncTests(t, signedTxTests, func(w io.Writer, val interface{}) {
		tx := val.(sTx)
		res, err := SignTransaction(tx.t, func(*Hash) (*Signature, error) {
			return tx.s, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		w.Write(res)
		return
	})
}
