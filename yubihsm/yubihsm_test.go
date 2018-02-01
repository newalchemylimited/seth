package yubihsm

import (
	"fmt"
	"testing"

	"github.com/newalchemylimited/seth"
)

const ConnectorURL = "http://127.0.0.1:12345"

func init() {
	if err := SetVerbosity(100); err != nil {
		panic(err)
	}
}

func TestConnect(t *testing.T) {
	conn, err := Connect(ConnectorURL)
	if err != nil {
		t.Fatal(err)
	}

	di, err := conn.DeviceInfo()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(di)

	ctx := new(Context)

	sess, err := conn.NewDerivedSession(1, []byte("password"), false, ctx)
	if err != nil {
		t.Fatal(err)
	}

	defer sess.Destroy()

	if err := sess.Authenticate(ctx); err != nil {
		t.Fatal(err)
	}

	objs, err := sess.ListObjects(nil)
	if err != nil {
		t.Fatal(err)
	}

	for i := range objs {
		fmt.Println(objs[i])
	}

	// Make a key.
	/*
		domains := Domains(1, 2, 3)

		caps, err := CapabilitiesByName("asymmetric_sign_ecdsa")
		if err != nil {
			t.Fatal(err)
		}

		algo := AlgoECK256

		id, err := sess.GenerateECKey("foo", domains, caps, algo)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(id)
	*/

	id := 40487

	obj, err := sess.GetObject(id, TypeAsymmetric)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(obj)

	pkbytes, err := sess.GetPublicKey(id)
	if err != nil {
		t.Fatal(err)
	}

	pk := new(seth.PublicKey)

	copy(pk[:], pkbytes)

	fmt.Println("Address:", pk.Address())

	signer := func(sum *seth.Hash) *seth.Signature {
		r, s, err := sess.SignECDSA(id, sum[:])
		if err != nil {
			t.Fatal(err)
		}

		sig := seth.NewSignature(r, s, 0)

		if pk0, err := seth.NewSignature(r, s, 0).Recover(sum); err != nil {
			t.Fatal(err)
		} else if *pk0 != *pk {
			sig = seth.NewSignature(r, s, 1)
		}

		return sig
	}

	addr, err := seth.ParseAddress("0x0000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}

	tx := &seth.Transaction{
		To:       addr,
		Value:    *seth.NewInt(0.0008e18),
		GasPrice: *seth.NewInt(2e9),
		Gas:      21000,
	}

	raw, err := seth.SignTransaction(tx, signer)
	if err != nil {
		t.Fatal(err)
	}

	client := seth.NewHTTPClient("http://api.myetherapi.com/eth")

	hash, err := client.RawCall(raw)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(hash)
}
