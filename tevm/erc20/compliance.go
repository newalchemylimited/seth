package erc20

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/newalchemylimited/seth"
	"github.com/newalchemylimited/seth/tevm"
)

type Token struct {
	*testing.T
	C       *tevm.Chain
	Addr    *seth.Address // address of the contract in the chain
	Owner   *seth.Address // if there is an owner, this address
	Mint    func(t *Token, addr *seth.Address, amt *seth.Int) error
	HasBurn bool // token has 'burn(uint256)' defined
}

func assertln(t *testing.T, cond bool, text string) {
	if !cond {
		t.Helper()
		t.Error(text)
	}
}

func mustError(t *testing.T, err error, text string) {
	t.Helper()
	if err == nil {
		t.Fatal("no error", text)
	}
	t.Logf("%s: %q", text, err)
}

func isERC20Approve(l *seth.Log, from, to *seth.Address) bool {
	return bytes.Equal(l.Topics[0], seth.ERC20Approve[:]) &&
		bytes.Equal(l.Topics[1][12:], from[:]) &&
		bytes.Equal(l.Topics[2][12:], to[:])
}

func isERC20Transfer(l *seth.Log, from, to *seth.Address) bool {
	return bytes.Equal(l.Topics[0], seth.ERC20Transfer[:]) &&
		bytes.Equal(l.Topics[1][12:], from[:]) &&
		bytes.Equal(l.Topics[2][12:], to[:])
}

func please(t *testing.T, err error) {
	if err != nil {
		t.Helper()
		t.Fatal(err)
	}
}

// TestABICompliance tests that the ABI for the contract
// matches the public interface of an ERC20 token, including
// constant functions and events.
func TestABICompliance(t *testing.T, cc *seth.CompiledContract) {
	funcs := []struct {
		sig      string
		constant bool
	}{
		{"transfer(address,uint256)", false},
		{"approve(address,uint256)", false},
		{"transferFrom(address,address,uint256)", false},
		{"allowance(address,address)", true},
		{"balanceOf(address)", true},
	}
	events := []struct {
		sig     string
		indexed []int
	}{
		{"Transfer(address,address,uint256)", []int{0, 1}},
		{"Approval(address,address,uint256)", []int{0, 1}},
	}

	for i := range funcs {
		d := cc.Find(funcs[i].sig)
		if d == nil {
			t.Errorf("couldn't find %q", funcs[i].sig)
		} else if d.Type != "function" {
			t.Errorf("%q not a function", funcs[i].sig)
		} else if d.Constant != funcs[i].constant {
			t.Errorf("%q (const=%v) != %v", funcs[i].sig, d.Constant, funcs[i].constant)
		}
	}

	for i := range events {
		d := cc.Find(events[i].sig)
		if d == nil {
			t.Errorf("couldn't find %q", events[i].sig)
		} else if d.Type != "event" {
			t.Errorf("%q not an event", events[i].sig)
		} else {
			for _, j := range events[i].indexed {
				if !d.Inputs[j].Indexed {
					t.Errorf("%q: arg %d not indexed", events[i].sig, j)
				}
			}
		}
	}
}

// TestERC20Compliance is a function that helps verify
// compliance with the ERC20 standard.
//
// 'token' should be the address of the token contract on the
// chain, and 'owner' should be the owner of the token contract.
// 'mint' should mint tokens to the given address in the given quantity
func TestCompliance(t *Token) {
	lc := len(t.C.Logs())
	addr := t.C.NewAccount(1)
	t.Mint(t, &addr, seth.NewInt(1000))

	// there should be a transfer from 0 in here,
	// since we minted some tokens
	logs := t.C.Logs()
	assertln(t.T, len(logs) == lc+1, "no Transfer emitted for minting")
	transfer := &logs[lc]
	var zeroaddr seth.Address
	assertln(t.T, isERC20Transfer(transfer, &zeroaddr, &addr), "didn't emit a proper transfer event")
	assertln(t.T, bytes.Equal(transfer.Address[:], t.Addr[:]), "event not emitted by token")

	balance := t.BalanceOf(&addr)
	assertln(t.T, balance.Int64() == 1000, "incorrect balance after mint")

	addr2 := t.C.NewAccount(1)
	ok, err := t.Transfer(&addr, &addr2, seth.NewInt(100))
	please(t.T, err)
	assertln(t.T, ok, "transfer() returned false")
	balance2 := t.BalanceOf(&addr2)
	if balance2.Int64() != 100 {
		t.Errorf("expected a balance of 100; found %d", balance2.Int64())
	}

	logs = t.C.Logs()
	assertln(t.T, len(logs) == lc+2, "no Transfer emitted after transfer() called")
	transfer = &logs[lc+1]
	assertln(t.T, isERC20Transfer(transfer, &addr, &addr2), "didn't emit a proper transfer event")
	assertln(t.T, bytes.Equal(transfer.Address[:], t.Addr[:]), "transfer event not emitted by token")

	balance = t.BalanceOf(&addr)
	if balance.Int64() != (1000 - 100) {
		t.Errorf("expected a balance of 900; found %d", balance.Int64())
	}

	ok, err = t.Approve(&addr, &addr2, seth.NewInt(100))
	please(t.T, err)
	assertln(t.T, ok, "approve() returned false")
	logs = t.C.Logs()
	assertln(t.T, len(logs) == lc+3, "no Approval event emitted after approve()")
	approve := &logs[lc+2]
	assertln(t.T, isERC20Approve(approve, &addr, &addr2), "didn't emit proper approve event")
	assertln(t.T, bytes.Equal(approve.Address[:], t.Addr[:]), "approve event not emitted by token")

	// this one should return false; the allowance isn't high enough
	ok, err = t.TransferFrom(&addr2, &addr, &addr2, seth.NewInt(200))

	// throwing or returning false are both acceptable behaviors
	assertln(t.T, !ok || err != nil, "transferFrom() that was too large succeeded anyway...?")
	assertln(t.T, len(t.C.Logs()) == lc+3, "event emitted even though transferFrom failed?")

	ok, err = t.TransferFrom(&addr2, &addr, &addr2, seth.NewInt(100))
	please(t.T, err)
	assertln(t.T, ok, "transferFrom() failed")
	logs = t.C.Logs()
	assertln(t.T, len(logs) == lc+4, "no Transfer emitted")
	transfer = &logs[lc+3]
	assertln(t.T, isERC20Transfer(transfer, &addr, &addr2), "didn't emit proper transfer event")
	assertln(t.T, bytes.Equal(transfer.Address[:], t.Addr[:]), "transfer event not emitted by token")

	balance = t.BalanceOf(&addr)
	assertln(t.T, balance.Int64() == 800, "balance is whack")

	// may as well test burning while we're here...
	if !t.HasBurn {
		return
	}
	supply := t.TotalSupply()
	_, err = t.C.Call(&addr2, t.Addr, "burn(uint256)", seth.NewInt(200))
	please(t.T, err)
	balance = t.BalanceOf(&addr2)
	assertln(t.T, balance.IsZero(), "balance isn't zero after burning 200 tokens...?")
	logs = t.C.Logs()
	assertln(t.T, len(logs) == lc+5, "no Transfer event emitted after burning")
	transfer = &logs[lc+4]
	assertln(t.T, isERC20Transfer(transfer, &addr2, &zeroaddr), "burn didn't emit proper transfer event")
	assertln(t.T, bytes.Equal(transfer.Address[:], t.Addr[:]), "transfer event not emitted by token")

	supply2 := t.TotalSupply()
	sp := (*big.Int)(supply)
	sp.Sub(sp, (*big.Int)(supply2))
	assertln(t.T, sp.Int64() == 200, "supply didn't fall by 200 after burning")
}

func ret2i(b []byte) *seth.Int {
	var i big.Int
	i.SetBytes(b)
	return (*seth.Int)(&i)
}

// BalanceOf calls token.balanceOf(addr)
func (t *Token) BalanceOf(addr *seth.Address) *seth.Int {
	t.Helper()
	bytes, err := t.C.StaticCall(t.Owner, t.Addr, "balanceOf(address)", addr)
	if err != nil {
		t.Fatalf("token.balanceOf(): %s", err)
	}
	return ret2i(bytes)
}

func ret2b(b []byte) bool {
	for i := range b {
		if b[i] != 0 {
			return true
		}
	}
	return false
}

func (t *Token) TotalSupply() *seth.Int {
	t.Helper()
	bytes, err := t.C.StaticCall(t.Owner, t.Addr, "totalSupply()")
	if err != nil {
		t.Fatal("totalSupply():", err)
	}
	return ret2i(bytes)
}

func (t *Token) Transfer(sender, to *seth.Address, amt *seth.Int) (bool, error) {
	bytes, err := t.C.Call(sender, t.Addr, "transfer(address,uint256)", to, amt)
	if err != nil {
		return false, err
	}
	return ret2b(bytes), nil
}

func (t *Token) TransferFrom(sender, from, to *seth.Address, amt *seth.Int) (bool, error) {
	bytes, err := t.C.Call(sender, t.Addr, "transferFrom(address,address,uint256)", from, to, amt)
	if err != nil {
		return false, err
	}
	return ret2b(bytes), nil
}

func (t *Token) Approve(sender, to *seth.Address, amt *seth.Int) (bool, error) {
	bytes, err := t.C.Call(sender, t.Addr, "approve(address,uint256)", to, amt)
	if err != nil {
		return false, err
	}
	return ret2b(bytes), nil
}
