package tevm

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/newalchemylimited/seth"
)

func TestCompileAndRun(t *testing.T) {
	bundle, err := CompileGlob("*.sol")
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
	opts := &seth.CallOpts{
		From: &me,
		To:   &contract,
	}
	opts.EncodeCall("counter()")
	if err := s.ConstCall(opts, (*seth.Int)(&v), true); err != nil {
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

func TestParseInfo(t *testing.T) {
	var c0, c1 CompiledContract
	c0.Sourcemap = "1:2:1;1:9:1;2:9:2;2:9:2;2:9:2"
	c0.compileSourcemap()
	c1.Sourcemap = "1:2:1;:9;2::2;;"
	c1.compileSourcemap()

	if !reflect.DeepEqual(c0.srcmap, c1.srcmap) {
		t.Errorf("%v != %v", c0.srcmap, c1.srcmap)
	}
}
