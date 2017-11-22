package tevm

import (
	"math/big"
	"testing"

	"github.com/newalchemylimited/seth"
)

func TestCompileAndRun(t *testing.T) {
	bundle, err := seth.CompileGlob("*.sol")
	if err != nil {
		t.Fatal(err)
	}
	cc := bundle.Contract("Test")
	if cc == nil {
		t.Fatal("didn't see contract Test in output")
	}
	if len(cc.Code) == 0 {
		t.Fatal("no output bytecode for contract Test")
	}
	if len(cc.Sourcemap) == 0 {
		t.Error("no sourcemap for Test")
	}

	if d := cc.Find("counter()"); d == nil {
		t.Error("couldn't find counter()")
	}
	if d := cc.Find("inc()"); d == nil {
		t.Error("couldn't find inc()")
	}
	if d := cc.Find("mustThrow()"); d == nil {
		t.Error("couldn't find mustThrow()")
	}

	c := NewChain()
	me := c.NewAccount(1)
	s := c.Sender(&me)

	var contract seth.Address

	contract, err = c.Create(&me, cc.Code)
	if err != nil {
		t.Fatalf("couldn't create contract: %s", err)
	}

	_, err = c.Call(&me, &contract, "inc()")
	if err != nil {
		t.Fatalf("inc(): %q", err)
	}

	bits, err := c.StaticCall(&me, &contract, "counter()")
	if err != nil {
		t.Fatalf("counter(): %q", err)
	}

	var v big.Int
	v.SetBytes(bits)
	if v.Int64() != 1 {
		t.Errorf("expected counter to be 1; found %d", v.Int64())
	}

	// Try with the sender.
	_, err = s.Send(&contract, "inc()")
	if err != nil {
		t.Fatalf("sender: inc(): %q", err)
	}
	if err := s.ConstCall(&contract, "counter()", (*seth.Int)(&v)); err != nil {
		t.Fatalf("sender: counter(): %q", err)
	}
	if v.Int64() != 2 {
		t.Errorf("sender: expected counter to be 2; found %d", v.Int64())
	}

	_, err = c.Call(&me, &contract, "mustThrow()")
	if err == nil {
		t.Fatal("expected mustThrow() to return an error")
	}
	t.Logf("calling mustThrow() returns %q", err)

	_, err = c.Call(&me, &contract, "inc()")
	if err != nil {
		t.Fatalf("inc(): %q", err)
	}

	bits, err = c.StaticCall(&me, &contract, "counter()")
	if err != nil {
		t.Fatalf("counter(): %q", err)
	}
	v.SetBytes(bits)
	if v.Int64() != 3 {
		t.Errorf("expected counter to be 3; found %d", v.Int64())
	}
}
