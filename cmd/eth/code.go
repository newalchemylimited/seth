package main

import (
	"flag"
	"fmt"

	"github.com/newalchemylimited/seth"
)

var cmdcode = &cmd{
	desc:  "get contract code",
	usage: "eth code <addr>",
	do:    code,
}

func init() {
	cmdcode.fs.Init("code", flag.ExitOnError)
}

func code(fs *flag.FlagSet) {
	args := fs.Args()
	if len(args) != 1 {
		fs.Usage()
	}

	var addr seth.Address
	if err := addr.FromString(args[0]); err != nil {
		fatalf("bad address: %s\n", err)
	}
	b := getcode(client(), &addr)
	fmt.Printf("%x\n", b)
}
