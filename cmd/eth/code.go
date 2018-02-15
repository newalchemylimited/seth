package main

import (
	"fmt"

	"github.com/newalchemylimited/seth"
)

var cmdcode = &cmd{
	desc: "get contract code",
	do:   code,
}

func code(args []string) {
	if len(args) != 1 {
		fatalf("usage: eth code <address>\n")
	}

	var addr seth.Address
	if err := addr.FromString(args[0]); err != nil {
		fatalf("bad address: %s\n", err)
	}
	b := getcode(client(), &addr)
	fmt.Printf("%x\n", b)
}
