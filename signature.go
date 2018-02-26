package seth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/newalchemylimited/seth/ec"
	"github.com/newalchemylimited/seth/keccak"
)

var (
	curve     = ec.S256()
	params    = curve.Params()
	order     = params.N
	halforder = new(big.Int).Rsh(order, 1)
	salt      = keccak.Sum512([]byte("github.com/newalchemylimited/seth/ec"))
)

// A Signature holds an ECDSA signature in Ethereum's compact representation.
type Signature [65]byte

// NewSignature constructs a signature from r, s, and v.
func NewSignature(r, s *big.Int, v int) *Signature {
	sig := new(Signature)
	if s.Cmp(halforder) > 0 {
		v ^= 1
		s = new(big.Int).Sub(order, s)
	}
	copypad(sig[0:32], r.Bytes())
	copypad(sig[32:64], s.Bytes())
	sig[64] = byte(v)
	return sig
}

// ParseSignature parses a signature from a string.
func ParseSignature(s string) (*Signature, error) {
	sig := new(Signature)
	if err := sig.FromString(s); err != nil {
		return nil, err
	}
	return sig, nil
}

// String returns a string representation of the signature.
func (s *Signature) String() string {
	return string(hexstring(s[:], false))
}

// FromString parses a signature from a string.
func (s *Signature) FromString(z string) error {
	return hexdecode(s[:], []byte(z))
}

// MarshalText implements encoding.TextMarshaler.
func (s *Signature) MarshalText() ([]byte, error) {
	return hexstring(s[:], false), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *Signature) UnmarshalText(b []byte) error {
	return hexdecode(s[:], b)
}

// Parts returns the r, s, and v parts of the signature.
func (z *Signature) Parts() (r, s big.Int, v int) {
	r, s, v, _ = z.parts()
	return
}

// parts returns r, s, and v and also validates them.
func (z *Signature) parts() (r, s big.Int, v int, ok bool) {
	v = int(z[64])
	if v != 0 && v != 1 {
		return
	}
	r.SetBytes(z[:32])
	if r.Sign() == 0 {
		return
	}
	s.SetBytes(z[32:64])
	if s.Sign() == 0 {
		return
	}
	ok = r.Cmp(order) < 0 && s.Cmp(halforder) <= 0
	return
}

// CurvePoint takes the x-coordinate of a secp256k1
// curve point and computes the corresponding y-coordinate.
// "v" should indicate the low bit of the y-coordinate.
func CurvePoint(x *big.Int, y *big.Int, v int) {
	// the curve is y^2 = x^3 + B
	y.Mul(x, x)
	y.Mul(y, x)
	y.Add(y, params.B)
	// for secp256k1, sqrt(x) = x^((p+1)/4)
	y.Exp(y, curve.QPlus1Div4(), params.P) // |y| = sqrt(x^3 + B)
	if y.Bit(0) != uint(v)&1 {
		// negate if the signedness is wrong
		y.Sub(params.P, y)
	}
}

// Recover validates the signature and returns a public key given the hash.
func (z *Signature) Recover(hash *Hash) (*PublicKey, error) {
	var e big.Int
	var ri, si big.Int
	var ry big.Int

	r, s, v, ok := z.parts()
	if !ok {
		return nil, errors.New("invalid signature")
	}

	CurvePoint(&r, &ry, v)

	ri.ModInverse(&r, order)
	si.Mul(&ri, &s)
	si.Mod(&si, order)
	sx, sy := curve.ScalarMult(&r, &ry, si.Bytes())

	e.SetBytes(hash[:])
	e.Neg(&e)
	e.Mod(&e, order)
	e.Mul(&e, &ri)
	e.Mod(&e, order)

	ex, ey := curve.ScalarBaseMult(e.Bytes())
	qx, qy := curve.Add(sx, sy, ex, ey)

	return NewPublicKey(qx, qy), nil
}

// Valid checks whether the signature is valid.
func (z *Signature) Valid() bool {
	_, _, _, ok := z.parts()
	return ok
}

// A PrivateKey holds a private key.
type PrivateKey [32]byte

// NewPrivateKey creates a private key from d.
func NewPrivateKey(d *big.Int) *PrivateKey {
	pk := new(PrivateKey)
	copypad(pk[:], d.Bytes())
	return pk
}

// GenPrivateKey generates a private key from the default entropy source.
func GenPrivateKey() *PrivateKey {
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}
	return NewPrivateKey(key.D)
}

// ToECDSA returns the private key as an ECDSA private key.
func (k *PrivateKey) ToECDSA() *ecdsa.PrivateKey {
	x, y := curve.ScalarBaseMult(k[:])
	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: new(big.Int).SetBytes(k[:]),
	}
}

// PublicKey returns the public key corresponding to this private key.
func (k *PrivateKey) PublicKey() *PublicKey {
	x, y := curve.ScalarBaseMult(k[:])
	return NewPublicKey(x, y)
}

// Address returns the address corresponding to this private key.
func (k *PrivateKey) Address() *Address {
	return k.PublicKey().Address()
}

// Sign a hash with this private key.
func (k *PrivateKey) Sign(hash *Hash) *Signature {
	var d, e big.Int
	var kb [32]byte
	var ki big.Int
	var s big.Int

	d.SetBytes(k[:])
	e.SetBytes(hash[:])

	md := keccak.New256()
	md.Write(k[:])
	md.Write(salt[:])
	md.Write(hash[:])
	md.Sum(kb[:0])

	block, _ := aes.NewCipher(kb[:])
	stream := cipher.NewCTR(block, salt[:16])

	for {
		stream.XORKeyStream(kb[:], kb[:])
		ki.SetBytes(kb[:])
		if ki.Sign() == 0 || ki.Cmp(order) >= 0 {
			continue
		}

		// ecdsa.fermatInverse is not used here because this package opts for
		// performance and clarity over constant time operations.
		ki.ModInverse(&ki, order)

		r, ry := curve.ScalarBaseMult(kb[:])
		if r.Sign() == 0 || r.Cmp(order) >= 0 {
			continue
		}
		v := int(ry.Bit(0))

		s.Mul(&d, r)
		s.Add(&s, &e)
		s.Mul(&s, &ki)
		s.Mod(&s, order)
		if s.Cmp(halforder) > 0 {
			v ^= 1
			s.Sub(order, &s)
		}
		if s.Sign() == 0 {
			continue
		}
		return NewSignature(r, &s, v)
	}
}

// Signer returns a Signer for this private key.
func (k *PrivateKey) Signer() Signer {
	return func(h *Hash) (*Signature, error) {
		return k.Sign(h), nil
	}
}

// ParsePrivateKey parses a private key.
func ParsePrivateKey(s string) (*PrivateKey, error) {
	k := new(PrivateKey)
	if err := k.FromString(s); err != nil {
		return nil, err
	}
	return k, nil
}

// String returns a string representation of the private key.
func (k *PrivateKey) String() string {
	return string(hexstring(k[:], false))
}

// FromString parses a private key from a string.
func (k *PrivateKey) FromString(s string) error {
	return hexdecode(k[:], []byte(s))
}

// MarshalText implements encoding.TextMarshaler.
func (k *PrivateKey) MarshalText() ([]byte, error) {
	return hexstring(k[:], false), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (k *PrivateKey) UnmarshalText(b []byte) error {
	return hexdecode(k[:], b)
}

// A PublicKey holds a public key.
type PublicKey [64]byte

// NewPublicKey creates a public key from X and Y.
func NewPublicKey(x, y *big.Int) *PublicKey {
	pk := new(PublicKey)
	copypad(pk[0:32], x.Bytes())
	copypad(pk[32:64], y.Bytes())
	return pk
}

// ToECDSA returns the public key as an ECDSA public key.
func (k *PublicKey) ToECDSA() *ecdsa.PublicKey {
	return &ecdsa.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(k[0:32]),
		Y:     new(big.Int).SetBytes(k[32:64]),
	}
}

// ParsePublicKey parses a public key.
func ParsePublicKey(s string) (*PublicKey, error) {
	k := new(PublicKey)
	if err := k.FromString(s); err != nil {
		return nil, err
	}
	return k, nil
}

// String returns a string representation of the public key.
func (k *PublicKey) String() string {
	return string(hexstring(k[:], false))
}

// FromString parses a public key from a string.
func (k *PublicKey) FromString(s string) error {
	return hexdecode(k[:], []byte(s))
}

// MarshalText implements encoding.TextMarshaler.
func (k *PublicKey) MarshalText() ([]byte, error) {
	return hexstring(k[:], false), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (k *PublicKey) UnmarshalText(b []byte) error {
	return hexdecode(k[:], b)
}

// Address returns the address corresponding to this public key.
func (k *PublicKey) Address() *Address {
	addr := new(Address)
	sum := keccak.Sum256(k[:])
	copy(addr[:], sum[12:])
	return addr
}

func copypad(dst, src []byte) {
	if len(src) > len(dst) {
		panic("copypad: src too big for dst")
	}
	b := len(dst) - len(src)
	for i := 0; i < b; i++ {
		dst[i] = 0
	}
	copy(dst[b:], src)
}
