package seth

import (
	"bytes"
	"io"
)

func intBytes(n Uint64) []byte {
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

func rlpEncodeInt(w io.Writer, n Uint64) {
	if n == 0 {
		w.Write([]byte{0x80})
	} else {
		encodeBytes(w, intBytes(n), 0x80, 0xB7)
	}
	return
}

func rlpEncodeString(w io.Writer, b []byte) {
	if bytes.Equal(b, []byte("")) {
		w.Write([]byte{0x80})
	} else if bytes.Equal(b, []byte{0x00}) {
		w.Write([]byte{0x00})
	} else {
		encodeBytes(w, b, 0x80, 0xB7)
	}
	return
}

func rlpEncodeList(w io.Writer, b []byte) {
	if len(b) == 1 && b[0] == 0x00 {
		w.Write([]byte{0x80})
	} else {
		encodeBytes(w, b, 0xC0, 0xF7)
	}
	return
}

func encodeBytes(w io.Writer, b []byte, sh byte, lh byte) {
	if len(b) == 1 && b[0] < 0x80 {
		w.Write(b)
		return
	}

	if len(b) < 56 {
		w.Write([]byte{byte(int(sh) + len(b))})
	} else {
		blen := intBytes(Uint64(len(b)))
		w.Write([]byte{byte(int(lh) + len(blen))})
		w.Write(blen)
	}

	w.Write(b)
	return
}

func encodeTransaction(w io.Writer, t *Transaction) {
	rlpEncodeInt(w, t.Nonce)
	rlpEncodeString(w, t.GasPrice.Big().Bytes())
	rlpEncodeInt(w, t.Gas)

	if t.To == nil {
		rlpEncodeString(w, []byte(""))
	} else {
		rlpEncodeString(w, t.To[:])
	}

	rlpEncodeString(w, t.Value.Big().Bytes())
	rlpEncodeString(w, t.Input)

	return
}

func rlpEncodeSignedTx(t *Transaction, sig *Signature) []byte {
	var res, buf bytes.Buffer
	encodeTransaction(&buf, t)

	r, s, v := sig.Parts()

	rlpEncodeInt(&buf, Uint64(v)+37)
	rlpEncodeString(&buf, r.Bytes())
	rlpEncodeString(&buf, s.Bytes())

	rlpEncodeList(&res, buf.Bytes())
	return res.Bytes()
}

type Signer func(*Hash) (*Signature, error)

// KeySigner is sugar that returns a Signer from a PrivateKey.
func KeySigner(k *PrivateKey) Signer {
	return func(h *Hash) (*Signature, error) {
		return k.Sign(h), nil
	}
}

// SignTransaction produces a signed, serialized 'raw' transaction
// from the given transaction and signer.
func SignTransaction(t *Transaction, sign Signer) ([]byte, error) {
	var data, res bytes.Buffer
	encodeTransaction(&data, t)
	rlpEncodeInt(&data, 1)
	rlpEncodeInt(&data, 0)
	rlpEncodeInt(&data, 0)

	rlpEncodeList(&res, data.Bytes())

	hash := HashBytes(res.Bytes())
	sig, err := sign(&hash)
	if err != nil {
		return nil, err
	}
	return rlpEncodeSignedTx(t, sig), nil
}
