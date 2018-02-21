package main

import (
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/newalchemylimited/seth"
)

var cmdkeygen = &cmd{
	desc: "generate a keyfile",
	do:   keygen,
}

// -o file
var keygenout string

func init() {
	flag.StringVar(&keygenout, "o", "", "output file")
}

func keygen(args []string) {
	if len(args) > 1 {
		fatalf("usage: eth keygen <name>\n")
	}
	name := ""
	if len(args) == 1 && args[0] != "" {
		name = args[0]
	}

	priv := seth.GenPrivateKey()
	kf := priv.ToKeyfile(name, passpromptf("enter unlock passphrase:"))
	buf, err := json.MarshalIndent(kf, "", "\t")
	if err != nil {
		panic(err)
	}

	var out io.WriteCloser
	if keygenout != "" {
		var err error
		out, err = os.Create(keygenout)
		if err != nil {
			fatalf("creating %s: %s\n", keygenout, err)
		}
	} else {
		out = os.Stdout
	}

	_, err = out.Write(buf)
	if err != nil {
		fatalf("%s\n", err)
	}
	err = out.Close()
	if err != nil {
		fatalf("%s\n", err)
	}
}
