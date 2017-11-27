package tevm

import (
	"bytes"
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
	t.Parallel()
	var st State
	s := st.StateDB()
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
	t.Parallel()
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

	n := uint64(*c.State.Pending.Number)
	h := c.State.Pending.Hash
	c.Seal()
	n2 := uint64(*c.State.Pending.Number)
	h2 := c.State.Pending.Hash
	if n2 != n+1 {
		t.Errorf("Seal() took block number from %d to %d", n, n2)
	}
	if bytes.Equal(h[:], h2[:]) {
		t.Errorf("blocks have equal hashes...?")
	}

	// make some state modifications
	err = c.Send(&dumb, &me, c.BalanceOf(&dumb))
	if err != nil {
		t.Fatal("couldn't send balance:", err)
	}

	please(t, c.BalanceOf(&me).Int64() == 1e18)
	please(t, c.BalanceOf(&dumb).Int64() == 0)

	// now revert the chain state and see that
	// we end up in the appropriate reverted state
	c = c.AtBlock(int64(n2 - 1))
	if c == nil {
		t.Fatalf("didn't revert to block %d", n2-1)
	}
	if c.State.Pending == nil {
		t.Fatal("didn't repopulate pending block?")
	}
	if uint64(*c.State.Pending.Number) != n {
		t.Errorf("reverted block has number %d instead of %d", *c.State.Pending.Number, n)
	}

	please(t, c.BalanceOf(&me).Int64() == 0)
	please(t, c.BalanceOf(&dumb).Int64() == 1e18)
}
