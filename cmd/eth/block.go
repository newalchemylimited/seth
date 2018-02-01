package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/newalchemylimited/seth"
)

var cmdblock = &cmd{
	desc: "print a block as json",
	do:   block,
}

func showblock(b *seth.Block) {
	buf, err := json.MarshalIndent(b, "", "\t")
	if err != nil {
		fatalf("fatal error: %s", err)
	}
	fmt.Printf("%s\n", buf)
}

func blockrange(c *seth.Client, spec string) {
	var split []string
	var opstr string
	for _, sep := range []string{"+", "-"} {
		if strings.Contains(spec, sep) {
			split = strings.Split(spec, sep)
			opstr = sep
			break
		}
	}
	if len(split) != 2 {
		fatalf("bad block range specifier %q\n", spec)
	}

	// resolve this block number into
	// an absolute number; we may actually
	// have to fetch the block to do this
	start := blocknum(split[0])
	if start < 0 {
		start = int64(*(getblock(c, start, false).Number))
	}

	diff, err := strconv.ParseInt(split[1], 0, 64)
	if err != nil {
		fatalf("couldn't parse %q as an integer: %s\n", split[1], err)
	}
	op := func(a, b int64) int64 { return a + b }
	if opstr == "-" {
		op = func(a, b int64) int64 { return a - b }
	}

	for i := int64(0); i < diff; i++ {
		showblock(getblock(c, op(start, i), true))
	}
}

func block(args []string) {
	if len(args) == 0 {
		args = append(args, "pending")
	}

	c := client()
	for i := range args {
		if strings.ContainsAny(args[i], "+-") {
			blockrange(c, args[i])
		} else {
			showblock(getblock(c, blocknum(args[i]), true))
		}
	}
}
