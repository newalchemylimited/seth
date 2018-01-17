package tevm

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/newalchemylimited/seth"
	"github.com/newalchemylimited/seth/cc"
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

// Test that a chain can be JSON marshaled and recovered.
func TestChainSerialization(t *testing.T) {
	t.Parallel()
	chain := NewChain()

	me := chain.NewAccount(1)
	chain.Create(&me, []byte{0xf0, 0x00, 0xba, 0x40})
	*chain.State.Pending.Number += 42
	chain.Seal()

	b, err := json.Marshal(chain)
	if err != nil {
		t.Fatal(err)
	}

	chain2 := new(Chain)

	if err := json.Unmarshal(b, chain2); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(chain, chain2) {
		t.Fatal("chain state did not match:\n", chain, "\n", chain2)
	}
}

// Test that creating a contract at an address works.
func TestCreateAt(t *testing.T) {
	t.Parallel()
	chain := NewChain()

	me := chain.NewAccount(1)
	addr, err := seth.ParseAddress("0x0123456789abcdef0123456789abcdef0123456")
	if err != nil {
		t.Fatal(err)
	}

	bundle, err := seth.CompileString(`
		pragma solidity ^0.4.18;
		contract Foo {
			address public owner;
			uint public a;
			function Foo() public {
				owner = msg.sender;
				a = 100;
			}
			function b(uint x) returns (uint) {
				require(msg.sender == owner);
				return a + x;
			}
		}
		contract Bad {
			function Bad() public {
				require(false);
			}
		}
	`)
	if err != nil {
		t.Fatal(err)
	}

	code := bundle.Contract("Foo").Code

	if err := chain.CreateAt(addr, &me, code); err != nil {
		t.Fatal(err)
	}

	sender := chain.Sender(&me)
	var out seth.Int
	in := seth.NewInt(50)

	if gas, err := chain.EstimateGas(&me, addr, "b(uint256)", &out, in); err != nil {
		t.Fatal(err)
	} else {
		if gas > 6000000 {
			t.Errorf("strange gas value: %d", gas)
		}
		t.Logf("call uses %d gas", gas)
	}

	if err := sender.ConstCall(addr, "b(uint256)", &out, in); err != nil {
		t.Fatal(err)
	} else if v := out.Int64(); v != 150 {
		t.Fatal("expected 150, got", v)
	}

	// Make sure the throwing case works.
	code = bundle.Contract("Bad").Code
	if err := chain.CreateAt(addr, &me, code); err == nil {
		t.Fatal("expected error, got nothing")
	}
}

func TestForkedChain(t *testing.T) {
	t.Parallel()
	chain := NewFork(seth.NewClientTransport(seth.InfuraTransport{}), 4876654)
	me := chain.NewAccount(1)

	// Check that the total supply of OMG is 140245398245132780789239631

	omg, _ := cc.CurrencyByName("OMG")
	var out, ts big.Int
	ts.SetString("140245398245132780789239631", 10)

	ret, err := chain.StaticCall(&me, omg.Addr(), "totalSupply()")
	if err != nil {
		t.Fatal(err)
	}
	out.SetBytes(ret)

	if out.Cmp(&ts) != 0 {
		t.Fatalf("%d != %d", &out, &ts)
	}

	chain2 := chain.Copy()

	me2 := chain2.NewAccount(1)

	ret, err = chain2.StaticCall(&me2, omg.Addr(), "totalSupply()")
	if err != nil {
		t.Fatal(err)
	}
	out.SetBytes(ret)

	if out.Cmp(&ts) != 0 {
		t.Fatalf("%d != %d", &out, &ts)
	}

	bal := chain.BalanceOf(&me2)
	if bal.Sign() != 0 {
		t.Errorf("non-zero balance (%d) on the other side of the chain copy...?", bal)
	}
}
