package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var cmdpost = &cmd{
	desc: "post a raw transaction",
	do:   post,
}

func init() {

}

func post(args []string) {
	var out [][]byte

	for i, a := range args {
		var in io.ReadCloser
		if a == "-" {
			in = os.Stdin
		} else {
			var err error
			in, err = os.Open(a)
			if err != nil {
				fatalf("%s\n", err)
			}
		}
		hexbuf, err := ioutil.ReadAll(in)
		in.Close()
		if err != nil {
			fatalf("reading arg%d: %s\n", i, err)
		}
		if len(hexbuf) == 0 {
			fatalf("zero-length input can't possibly be a transaction\n")
		}
		if hexbuf[len(hexbuf)-1] == '\n' {
			hexbuf = hexbuf[:len(hexbuf)-1]
		}
		buf := make([]byte, hex.DecodedLen(len(hexbuf)))
		_, err = hex.Decode(buf, hexbuf)
		if err != nil {
			fatalf("decoding arg %d: %s", i, err)
		}
		out = append(out, buf)
	}

	c := client()
	for i, b := range out {
		tx, err := c.RawCall(b)
		if err != nil {
			fatalf("sending arg %d: %s\n", i, err)
		}
		fmt.Println(tx.String())
	}
}
