package seth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"testing"
)

func TestHashes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		h    Hash
		s    string
		desc string
	}{
		{ERC20Transfer, "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", "ERC20 transfer"},
		{ERC20Approve, "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925", "ERC20 approve"},
	}

	for _, c := range cases {
		cs := c.h.String()
		if cs != c.s {
			t.Errorf("%s: %q != %q", c.desc, cs, c.s)
		}
	}
}

func TestScanAddress(t *testing.T) {
	t.Parallel()
	const in = "hi 0x6810e776880c02933d47db1b9fc05908e5386b96 bye"

	var a, b string
	var addr Address
	_, err := fmt.Sscanf(in, "%s %a %s", &a, &addr, &b)
	if err != nil {
		t.Fatal(err)
	}
	if addr.String() != "0x6810e776880c02933d47db1b9fc05908e5386b96" {
		t.Fatal(addr.String(), "0x6810e776880c02933d47db1b9fc05908e5386b96")
	}
}

func TestScanHash(t *testing.T) {
	t.Parallel()
	const in = "hi 0x2801d9a7473b13e05282308e3006c22126503e2fb23212bffaef5567c5952494 bye"
	var a, b string
	var h Hash
	_, err := fmt.Sscanf(in, "%s %h %s", &a, &h, &b)
	if err != nil {
		t.Fatal(err)
	}
	if h.String() != "0x2801d9a7473b13e05282308e3006c22126503e2fb23212bffaef5567c5952494" {
		t.Fatal(h.String(), "0x2801d9a7473b13e05282308e3006c22126503e2fb23212bffaef5567c5952494")
	}
}

func TestScanHexInt(t *testing.T) {
	t.Parallel()
	const in = "hi 0x123 bye"
	var a, b string
	var i Int
	_, err := fmt.Sscanf(in, "%s %x %s", &a, &i, &b)
	if err != nil {
		t.Fatal(err)
	}
	if i.Int64() != 291 {
		t.Fatal(i.Int64(), 291)
	}
}

func TestHexParse(t *testing.T) {
	t.Parallel()
	// odd-length hex strings have an implicit
	// zero at the beginning, not the end (since
	// they represent big-endian numbers)
	in := []byte("0xe468e")
	out, err := hexparse(in)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x0e, 0x46, 0x8e}
	if !bytes.Equal(out, want) {
		t.Fatalf("%x != %x", want, out)
	}

	// test that we get the same string back out
	if !bytes.Equal(hexstring(want, true), in) {
		t.Errorf("%s != %s", hexstring(want, true), in)
	}

	// test that the encoding of 0 is 0x0
	zero := []byte{'0', 'x', '0'}
	if !bytes.Equal(hexstring([]byte{0}, true), zero) {
		t.Errorf("%s != 0x0", hexstring([]byte{0}, true))
	}
}

func TestSign(t *testing.T) {
	t.Parallel()
	var hash Hash

	for i := 0; i < 100; i++ {
		priv := GenPrivateKey()
		pub := priv.PublicKey()
		rand.Read(hash[:])

		// Test sign/recover with itself.
		sig := priv.Sign(&hash)
		if rec, err := sig.Recover(&hash); err != nil {
			t.Errorf("%s: could not verify: %v", sig, err)
		} else if *rec != *pub {
			t.Errorf("%s: recovered key doesn't match: %s != %s", sig, rec, pub)
		}

		if !sig.Valid() {
			t.Errorf("%s: not valid", sig)
		}

		// Test recover against ecdsa.Sign.
		r, s, _ := ecdsa.Sign(rand.Reader, priv.ToECDSA(), hash[:])
		if s.Cmp(halforder) == 1 {
			s.Sub(order, s)
		}
		sig2 := NewSignature(r, s, 0)
		if rec, err := sig2.Recover(&hash); err != nil || *rec != *pub {
			sig2[64] = 1
		}
		if rec, err := sig2.Recover(&hash); err != nil {
			t.Errorf("%s: could not verify: %v", sig2, err)
		} else if *rec != *pub {
			t.Errorf("%s: recovered key doesn't match: %s != %s", sig2, rec, pub)
		}

		// Test sign against ecdsa.Verify.
		r2, s2, _ := sig.Parts()
		if !ecdsa.Verify(pub.ToECDSA(), hash[:], &r2, &s2) {
			t.Errorf("%s: verification failed", sig)
		}
	}
}

func TestPubKeyToAddress(t *testing.T) {
	t.Parallel()
	pubkey, _ := ParsePublicKey("0x3f509f1ce5b0d2b255ba3c0a51ce36dcb06928904ebd8313f9e2e0a37cd5d60aefef6fb6a2e0a1708634a7d82df71bf103ab720247e215ced7643d9b1f85dc87")
	addr1, _ := ParseAddress("0x95f9aee97b55b06fb246a523bc36dbcf2cadf5f2")
	addr2 := pubkey.Address()
	if *addr1 != *addr2 {
		t.Errorf("addr mismatch: want: %s have: %s", addr1, addr2)
	}
}

func TestReceipt(t *testing.T) {
	t.Parallel()
	var (
		good, _ = ParseHash("0x97b25b137505e573d105be57dd4d3f7ddeb69b2c29608c86f055c5266a81c272")
		bad, _  = ParseHash("0x027aa484de0a14d84f91c4e355f5aeb966c91fd9a144a68b2a704ad882b9f52d")
		ugly, _ = ParseHash("0xf00ba4f00ba4f00ba4f00ba4f00ba4f00ba4f00ba4f00ba4f00ba4f00ba4f00b")
	)

	for _, c := range []*Client{
		NewHTTPClient("https://api.myetherapi.com/eth"),
		NewClientTransport(InfuraTransport{}),
	} {

		if r, err := c.GetReceipt(good); err != nil {
			t.Error(err)
		} else if r == nil {
			t.Error("receipt not found")
		} else if r.Threw() {
			t.Error("receipt indicates failure")
		}

		if r, err := c.GetReceipt(bad); err != nil {
			t.Error(err)
		} else if r == nil {
			t.Error("receipt not found")
		} else if !r.Threw() {
			t.Error("receipt indicates success")
		}

		if _, err := c.GetReceipt(ugly); err != ErrNotFound {
			t.Error("expected not found, but got:", err)
		}
	}
}

func TestIntUnmarshal(t *testing.T) {
	t.Parallel()
	for _, b := range []string{`"0x10"`, `16`, `"16"`} {
		var i Int
		if err := json.Unmarshal([]byte(b), &i); err != nil {
			t.Error(b+":", err)
		} else if i.Int64() != 16 {
			t.Error(b+": expected 16, got", i.Int64())
		}
	}
}

func TestGetNonce(t *testing.T) {
	t.Parallel()

	// Bittrex address
	addr, _ := ParseAddress("0xfbb1b73c4f0bda4f67dca266ce6ef42f520fbb98")
	client := NewHTTPClient("https://api.myetherapi.com/eth")

	const min = 4124441 // Nonce at time the test was added.

	nonce, err := client.GetNonce(addr)
	if err != nil {
		t.Fatal(err)
	} else if nonce < min {
		t.Fatal("nonce is too low:", nonce, "<", min)
	}
}
