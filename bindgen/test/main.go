package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"

	"github.com/newalchemylimited/seth"
	"github.com/newalchemylimited/seth/tevm"
)

//go:generate bindgen -b -c=Test -o generated.go compiletest.sol

func fatal(j ...interface{}) {
	fmt.Fprintln(os.Stderr, j...)
	os.Exit(1)
}

func check(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fatal(file, line, err)
	}
}

func assert(cond bool) {
	if !cond {
		_, file, line, _ := runtime.Caller(1)
		fatal(file, line, "assertion failed")
	}
}

func main() {
	// note: this program gets run from the parent directory
	bundle, err := seth.CompileGlob("test/*.sol")
	if err != nil {
		fatal(err)
	}

	c := tevm.NewChain()
	acct := c.NewAccount(1)

	contract := bundle.Contract("Test")
	ccode := contract.Code

	if !bytes.Equal(seth.StripBytecode(ccode), seth.StripBytecode(TestCode)) {
		fatal("compiled and precompiled code not identical")
	}

	addr, err := c.Create(&acct, TestCode)
	if err != nil {
		fatal("deploying the contract:", err)
	}

	cc := NewTest(&addr, c.Sender(&acct))

	v, err := cc.Value()
	check(err)
	assert(v.Int64() == 0)

	v, err = cc.Counter()
	check(err)
	assert(v.Int64() == 0)

	_, err = cc.MustThrow()
	assert(err != nil)

	_, err = cc.Inc()
	check(err)

	v, err = cc.Value()
	check(err)
	assert(v.Int64() == 1)

	v, err = cc.Counter()
	check(err)
	assert(v.Int64() == 1)
}
