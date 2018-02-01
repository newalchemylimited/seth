package main

import (
	"flag"
	"fmt"
	"math/big"

	"github.com/newalchemylimited/seth"
	"github.com/newalchemylimited/seth/cc"
)

var cmdbal = &cmd{
	desc: "get the balance of an account",
	do:   getbal,
}

var hexbalance bool
var decbalance bool

func init() {
	cmdbal.fs.Init("balance", flag.ExitOnError)
	cmdbal.fs.BoolVar(&hexbalance, "x", false, "print balance in hex")
	cmdbal.fs.BoolVar(&decbalance, "d", false, "print balance as an integer")
}

func getbal(args []string) {
	if len(args) == 0 {
		fatalf("usage: eth balance <addresses...>\n")
	}
	addrs := make([]seth.Address, len(args))
	for i := range args {
		if err := addrs[i].FromString(args[i]); err != nil {
			fatalf("cannot convert %q to address: %s\n", args[i], err)
		}
	}
	c := client()
	for i := range addrs {
		bal, err := c.GetBalance(&addrs[i])
		if err != nil {
			fatalf("fatal error getting balance: %s\n", err)
		}
		if hexbalance {
			fmt.Println(bal.String())
			continue
		} else if decbalance {
			b := ((big.Int)(bal))
			fmt.Println(b.String())
			continue
		}
		amt := cc.Amount{"ETH", (big.Int)(bal)}
		fmt.Println(amt.String())
	}
}
