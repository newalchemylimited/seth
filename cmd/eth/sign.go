package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/newalchemylimited/seth"
)

var cmdsign = &cmd{
	desc: "sign input",
	do:   sign,
}

var signprefix string // prefix to add to signature
var sighex bool       // output hex
var hashed bool       // input is already hashed

func init() {
	cmdsign.fs.StringVar(&signprefix, "prefix", "", "signing prefix")
	cmdsign.fs.BoolVar(&hashed, "h", false, "input already hashed")
	cmdsign.fs.BoolVar(&sighex, "x", false, "output is in hex instead of binary")
}

func sign(args []string) {
	if len(args) != 1 {
		fmt.Println("usage: eth sign [-h|-x|-prefix] <infile>")
		os.Exit(1)
	}
	if hashed && signprefix != "" {
		fatalf("cannot add a prefix to hashed plaintext")
	}

	f, err := os.Open(args[0])
	if err != nil {
		fatalf("cannot sign %s: %s", args[0], err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		fatalf("reading: %s", err)
	}
	var h seth.Hash
	if hashed {
		if len(buf) != len(h[:]) {
			fatalf("input length %d is not a keccak256 hash", len(buf))
		}
		copy(h[:], buf)
	} else {
		if signprefix != "" {
			buf = append([]byte(signprefix), buf...)
		}
		h = seth.HashBytes(buf)
	}

	fn := signer()
	sig := fn(&h)
	if sighex {
		_, err := io.WriteString(os.Stdout, hex.EncodeToString(sig[:]))
		if err != nil {
			fatalf("%s\n", err)
		}
	} else {
		_, err := os.Stdout.Write(sig[:])
		if err != nil {
			fatalf("%s\n", err)
		}
	}
	if err := os.Stdout.Close(); err != nil {
		fatalf("%s\n", err)
	}
}
