// +build yubihsm

package seth

import (
	"crypto/rand"
	"testing"

	"github.com/newalchemylimited/seth/yubihsm"
)

func TestYubiHSM(t *testing.T) {
	yubihsm.SetVerbosity(100)

	hsm := FindHSM("yubihsm")
	if hsm == nil {
		t.Skip("didn't find yubihsm device; skipping")
	}
	t.Log("hsm probe succeeded")

	if err := hsm.Unlock([]byte("password")); err != nil {
		t.Fatalf("couldn't unlock hsm: %s", err)
	}

	keys, err := hsm.Pubkeys()
	if err != nil {
		t.Fatalf("listing hsm keys: %s", err)
	}

	for i := range keys {
		t.Logf("found address %s", keys[i].Public.Address())
	}

	var h Hash
	for i := range keys {
		rand.Read(h[:])

		sign, err := hsm.Signer(&keys[i])
		if err != nil {
			t.Errorf("getting signer %s: %s", keys[i].ID, err)
			continue
		}

		sig, err := sign(&h)
		if err != nil {
			t.Errorf("key %s: signing: %s", keys[i].ID, err)
			continue
		}

		addr, err := sig.Recover(&h)
		if err != nil {
			t.Errorf("key %s: recover: %s", keys[i].ID, err)
			continue
		}
		if *addr != keys[i].Public {
			t.Errorf("keys: %s != %s", addr, &keys[i].Public)
		}
	}
}
