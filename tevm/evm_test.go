package tevm

import (
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func tracefn(t *testing.T) func(s string, args ...interface{}) {
	return func(s string, args ...interface{}) {
		t.Helper()
		t.Log(s, args)
	}
}

func please(t *testing.T, cond bool) {
	t.Helper()
	if !cond {
		t.Fatal("nope")
	}
}

func TestStateBasic(t *testing.T) {
	var st State
	var s gethState
	s.State = &st
	var addr common.Address
	rand.Read(addr[:])

	s.CreateAccount(addr)
	please(t, s.Exist(addr))
	please(t, s.Empty(addr))
	please(t, s.GetNonce(addr) == 0)
	please(t, s.GetCodeSize(addr) == 0)
	please(t, !s.HasSuicided(addr))
	please(t, s.Suicide(addr))
	please(t, s.HasSuicided(addr))
}

func TestChainBasic(t *testing.T) {
	c := NewChain()
	// c.State.Trace = tracefn(t) // -- for debugging
	me := c.NewAccount(1)
	please(t, c.BalanceOf(&me).Int64() == 1e18)

	s := c.Sender(&me)
	bal, err := s.GetBalance(&me)
	if err != nil {
		t.Fatal("couldn't get balance:", err)
	}
	please(t, bal.Int64() == 1e18)

	// create a contract that just returns
	dumb, err := c.Create(&me, []byte{0x0, 0x0, 0x0, 0x0})
	if err != nil {
		t.Fatal("couldn't create contract:", err)
	}

	// same but with sender
	if _, err := s.Create([]byte{0x0, 0x0, 0x0, 0x0}); err != nil {
		t.Fatal("sender couldn't create contract:", err)
	}

	err = c.Send(&me, &dumb, c.BalanceOf(&me))
	if err != nil {
		t.Fatal("couldn't send balance to contract:", err)
	}

	please(t, c.BalanceOf(&dumb).Int64() == 1e18)
	please(t, c.BalanceOf(&me).Int64() == 0)
}
