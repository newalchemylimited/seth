package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/newalchemylimited/seth"
)

var cmdrecover = &cmd{
	desc: "recover the address from a signature",
	do:   ecrecover,
}

var ecrpub bool // print public key

func init() {
	cmdrecover.fs.BoolVar(&ecrpub, "pub", false, "print public key")
}

func ecrecover(args []string) {
	if len(args) < 1 || len(args) > 2 {
		fmt.Println("usage: eth recover [-pub] <sig> [hash]")
		os.Exit(1)
	}

	sig, err := seth.ParseSignature(args[0])
	if err != nil {
		fatal("signature:", err)
	}

	var hash seth.Hash

	if len(args) == 1 {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fatal("failed reading:", err)
		}
		hash = seth.HashBytes(b)
	} else if err := hash.FromString(args[1]); err != nil {
		fatal("hash:", err)
	}

	pk, err := sig.Recover(&hash)
	if err != nil {
		fatal("recover:", err)
	}

	if ecrpub {
		fmt.Println(pk)
	} else {
		fmt.Println(pk.Address())
	}
}
