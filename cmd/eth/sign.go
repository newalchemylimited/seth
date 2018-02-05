package main

import (
	"encoding/hex"
	"encoding/json"
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
var sigjson bool      // output in json
var hashed bool       // input is already hashed

func bool2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func init() {
	cmdsign.fs.StringVar(&signprefix, "prefix", "", "signing prefix")
	cmdsign.fs.BoolVar(&hashed, "h", false, "input already hashed")
	cmdsign.fs.BoolVar(&sighex, "x", false, "output is in hex instead of binary")
	cmdsign.fs.BoolVar(&sigjson, "j", false, "output is in json instead of binary")
}

func sign(args []string) {
	if len(args) != 1 {
		fmt.Println("usage: eth sign [-h|-x|-j|-prefix] <infile>")
		os.Exit(1)
	}
	if hashed && signprefix != "" {
		fatalf("cannot add a prefix to hashed plaintext\n")
	}

	if bool2i(sighex)+bool2i(sigjson) > 1 {
		fatalf("eth sign: cannot specify more than one of -x or -j at a time\n")
	}

	f, err := os.Open(args[0])
	if err != nil {
		fatalf("cannot sign %s: %s\n", args[0], err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		fatalf("reading: %s\n", err)
	}
	var h seth.Hash
	if hashed {
		if len(buf) != len(h[:]) {
			fatalf("input length %d is not a keccak256 hash\n", len(buf))
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
	switch {
	case sighex:
		_, err := io.WriteString(os.Stdout, hex.EncodeToString(sig[:]))
		if err != nil {
			fatalf("%s\n", err)
		}
	case sigjson:
		// marshal r, s, and v like they would appear
		// in the JSON representation of a transaction
		r, s, v := sig.Parts()
		buf, _ := json.MarshalIndent(&struct {
			R seth.Int  `json:"r"`
			S seth.Int  `json:"s"`
			V *seth.Int `json:"v"`
		}{seth.Int(r), seth.Int(s), seth.NewInt(int64(v))}, "", "\t")
		buf = append(buf, '\n')
		_, err := os.Stdout.Write(buf)
		if err != nil {
			fatalf("%s\n", err)
		}
	default:
		_, err := os.Stdout.Write(sig[:])
		if err != nil {
			fatalf("%s\n", err)
		}
	}
	if err := os.Stdout.Close(); err != nil {
		fatalf("%s\n", err)
	}
}
