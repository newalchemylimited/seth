package seth

import (
	"reflect"
	"testing"
)

func TestCompileABI(t *testing.T) {
	bundle, err := Compile([]Source{{
		Body: `pragma solidity >=0.4.10;
contract Test {
	uint public counter;

	function Test() {
	}

	function value() constant returns(uint) {
		return counter;
	}

	function mustThrow() {
		require(false);
	}

	function inc() {
		counter = counter + 1;
	}
}
`,
		Filename: "test.sol",
	}})
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

	stripped := StripBytecode(cc.Code)
	if len(stripped) != len(cc.Code)-43 {
		t.Error("didn't strip bytecode")
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
