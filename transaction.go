package seth

import (
	"bytes"
)

// Transaction represents an ethereum transaction.
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

// Encode returns an RLP encoded representation of the transaction. If a
// signature is provided, this will return an encoded representation containing
// the signature.
func (t *Transaction) Encode(sig *Signature) []byte {
	var e rlpEncoder
	if sig == nil {
		e.EncodeTransaction(t)
	} else {
		e.EncodeSignedTx(t, sig)
	}
	return e.Bytes()
}

// HashToSign returns a hash which can be used to sign the transaction.
func (t *Transaction) HashToSign() *Hash {
	var data, res rlpEncoder

	data.EncodeTransaction(t)
	data.EncodeInt(1)
	data.EncodeInt(0)
	data.EncodeInt(0)

	res.EncodeList(data.Bytes())

	hash := HashBytes(res.Bytes())

	return &hash
}

// An rlpEncoder is a byte buffer that can RLP encode values.
type rlpEncoder bytes.Buffer

// Write proxies to bytes.Buffer.Write.
func (e *rlpEncoder) Write(b []byte) (n int, err error) {
	return (*bytes.Buffer)(e).Write(b)
}

// Bytes proxies to bytes.Buffer.Bytes.
func (e *rlpEncoder) Bytes() []byte {
	return (*bytes.Buffer)(e).Bytes()
}

func intBytes(n uint64) []byte {
	var buf bytes.Buffer
	shift := uint64(0)

	for (n >> shift) > 0xFF {
		shift += 8
	}

	for shift > 0 {
		buf.Write([]byte{byte(n >> shift)})
		shift -= 8
	}
	buf.Write([]byte{byte(n)})
	return buf.Bytes()
}

// EncodeInt encodes an int.
func (e *rlpEncoder) EncodeInt(n uint64) {
	if n == 0 {
		e.Write([]byte{0x80})
	} else {
		e.EncodeBytes(intBytes(n), 0x80, 0xB7)
	}
}

// EncodeString encodes a string.
func (e *rlpEncoder) EncodeString(b []byte) {
	if len(b) == 0 {
		e.Write([]byte{0x80})
	} else if bytes.Equal(b, []byte{0x00}) {
		e.Write([]byte{0x00})
	} else {
		e.EncodeBytes(b, 0x80, 0xB7)
	}
}

// EncodeList encodes a list of bytes.
func (e *rlpEncoder) EncodeList(b []byte) {
	if len(b) == 1 && b[0] == 0x00 {
		e.Write([]byte{0x80})
	} else {
		e.EncodeBytes(b, 0xC0, 0xF7)
	}
}

// EncodeBytes encodes bytes using sh or lh as headers depending on len(b).
func (e *rlpEncoder) EncodeBytes(b []byte, sh byte, lh byte) {
	if len(b) == 1 && b[0] < 0x80 {
		e.Write(b)
		return
	}

	if len(b) < 56 {
		e.Write([]byte{byte(int(sh) + len(b))})
	} else {
		blen := intBytes(uint64(len(b)))
		e.Write([]byte{byte(int(lh) + len(blen))})
		e.Write(blen)
	}

	e.Write(b)
}

// EncodeTransaction encodes the given transaction.
func (e *rlpEncoder) EncodeTransaction(t *Transaction) {
	e.EncodeInt(uint64(t.Nonce))
	e.EncodeString(t.GasPrice.Big().Bytes())
	e.EncodeInt(uint64(t.Gas))

	if t.To == nil {
		e.EncodeString(nil)
	} else {
		e.EncodeString(t.To[:])
	}

	e.EncodeString(t.Value.Big().Bytes())
	e.EncodeString(t.Input)

	return
}

// EncodeSignedTx encodes a transaction with the given signature.
func (e *rlpEncoder) EncodeSignedTx(t *Transaction, sig *Signature) {
	var buf rlpEncoder
	buf.EncodeTransaction(t)

	r, s, v := sig.Parts()

	buf.EncodeInt(uint64(v) + 37)
	buf.EncodeString(r.Bytes())
	buf.EncodeString(s.Bytes())

	e.EncodeList(buf.Bytes())
}

// A Signer is a function capable of signing a hash.
type Signer func(*Hash) (*Signature, error)

// SignTransaction produces a signed, serialized 'raw' transaction
// from the given transaction and signer.
func SignTransaction(t *Transaction, sign Signer) ([]byte, error) {
	hash := t.HashToSign()
	sig, err := sign(hash)
	if err != nil {
		return nil, err
	}
	return t.Encode(sig), nil
}
