package seth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"strings"

	"github.com/newalchemylimited/seth/keccak"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

// Keyfile represents an Ethereum key file
// as defined in the "Web3 Secret Storage Definition"
type Keyfile struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Crypto  struct {
		Cipher       string          `json:"cipher"`
		CipherParams json.RawMessage `json:"cipherparams"`
		Ciphertext   string          `json:"ciphertext"`
		KDF          string          `json:"kdf"`
		KDFParams    json.RawMessage `json:"kdfparams"`
		MAC          string          `json:"mac"`
	} `json:"crypto"`
	Address Address           `json:"address,omitempty"`
	Name    string            `json:"name"`
	Meta    map[string]string `json:"meta"`
}

func (k *Keyfile) ciphertext() ([]byte, error) {
	return hex.DecodeString(k.Crypto.Ciphertext)
}

func pbkdf2Derive(pass []byte, jsparams json.RawMessage) ([]byte, error) {
	type params struct {
		Iters  int    `json:"c"`
		Keylen int    `json:"dklen"`
		Hashfn string `json:"prf"`
		Salt   string `json:"salt"`
	}
	p := new(params)
	if err := json.Unmarshal(jsparams, p); err != nil {
		return nil, fmt.Errorf("parsing pbkdf2 params: %q", err)
	}
	var h func() hash.Hash
	switch strings.ToLower(p.Hashfn) {
	case "hmac-sha512":
		h = sha512.New
	case "hmac-sha384":
		h = sha512.New384
	case "hmac-sha256":
		h = sha256.New
	case "hmac-sha224":
		h = sha256.New224
	case "hmac-sha1":
		h = sha1.New
	default:
		return nil, fmt.Errorf("bad prf %q for pbkdf2", p.Hashfn)
	}

	salt, err := hex.DecodeString(p.Salt)
	if err != nil {
		return nil, fmt.Errorf("pbkdf2 salt: %s", err)
	}

	return pbkdf2.Key(pass, salt, p.Iters, p.Keylen, h), nil
}

func scryptDerive(pass []byte, jsparams json.RawMessage) ([]byte, error) {
	type params struct {
		Keylen int    `json:"dklen"`
		N      int    `json:"n"`
		R      int    `json:"r"`
		P      int    `json:"p"`
		Salt   string `json:"salt"`
	}
	p := new(params)
	if err := json.Unmarshal(jsparams, p); err != nil {
		return nil, fmt.Errorf("parsing scrypt params: %q", err)
	}
	salt, err := hex.DecodeString(p.Salt)
	if err != nil {
		return nil, err
	}
	return scrypt.Key(pass, salt, p.N, p.R, p.P, p.Keylen)
}

func (k *Keyfile) kdf(pass []byte) ([]byte, error) {
	switch strings.ToLower(k.Crypto.KDF) {
	case "pbkdf2":
		return pbkdf2Derive(pass, k.Crypto.KDFParams)
	case "scrypt":
		return scryptDerive(pass, k.Crypto.KDFParams)
	default:
		return nil, fmt.Errorf("unimplemented KDF %q", k.Crypto.KDF)
	}
}

func aes128ctrDecipher(key, ciphertext []byte, jsparams json.RawMessage) error {
	type params struct {
		IV string `json:"iv"`
	}
	p := new(params)
	err := json.Unmarshal(jsparams, p)
	if err != nil {
		return fmt.Errorf("getting params for aes-128-ctr: %s", err)
	}
	iv, err := hex.DecodeString(p.IV)
	if err != nil {
		return fmt.Errorf("bad aes-128-ctr iv: %s", err)
	}
	block, err := aes.NewCipher(key[:16])
	if err != nil {
		return err
	}
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return nil
}

// the keyfile MAC is the keccak256 of the last 16 bytes
// of the derived key, concatenated with the ciphertext
func (k *Keyfile) checkmac(key, ciphertext []byte) error {
	h := keccak.New256()
	h.Write(key[len(key)-16:])
	h.Write(ciphertext)
	sum := h.Sum(nil)
	want, err := hex.DecodeString(k.Crypto.MAC)
	if err != nil {
		return err
	}
	if !bytes.Equal(sum, want) {
		return fmt.Errorf("bad mac %q (bad passphrase?)", k.Crypto.MAC)
	}
	return nil
}

// Private uses a passphrase to unlock a keyfile
// and produce its private key.
func (k *Keyfile) Private(passphrase []byte) (*PrivateKey, error) {
	key, err := k.kdf(passphrase)
	if err != nil {
		return nil, err
	}
	ciphertext, err := k.ciphertext()
	if err != nil {
		return nil, err
	}
	err = k.checkmac(key, ciphertext)
	if err != nil {
		return nil, err
	}

	var decipher func(key, ciphertext []byte, params json.RawMessage) error
	switch strings.ToLower(k.Crypto.Cipher) {
	case "aes-128-ctr":
		decipher = aes128ctrDecipher
	default:
		return nil, fmt.Errorf("unimplemented cipher %q", k.Crypto.Cipher)
	}

	err = decipher(key, ciphertext, k.Crypto.CipherParams)
	if err != nil {
		return nil, err
	}
	priv := new(PrivateKey)
	copy(priv[:], ciphertext)
	var zeroaddr Address
	if k.Address != zeroaddr {
		addr := priv.Address()
		if !bytes.Equal(addr[:], k.Address[:]) {
			return nil, fmt.Errorf("derived address %q; want address %q", addr, &k.Address)
		}
	}
	return priv, nil
}
