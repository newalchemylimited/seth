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
	var err error
	var start, diff int64
	switch split[0] {
	case "pending":
		start = pending(c)
	case "latest":
		start = latest(c)
	default:
		start, err = strconv.ParseInt(split[0], 0, 64)
		if err != nil {
			fatalf("couldn't parse %q as an integer: %s\n", split[0], err)
		}
	}
	diff, err = strconv.ParseInt(split[1], 0, 64)
	if err != nil {
		fatalf("couldn't parse %q as an integer: %s\n", split[1], err)
	}
	op := func(a, b int64) int64 { return a + b }
	if opstr == "-" {
		op = func(a, b int64) int64 { return a - b }
	}

	for i := int64(0); i < diff; i++ {
		b, err := c.GetBlock(op(start, i), true)
		if err != nil {
			fatalf("getting block: %s\n", err)
		}
		showblock(b)
	}
}

func block(args []string) {
	if len(args) == 0 {
		args = append(args, "pending")
	}

	c := client()
	for i := range args {
		var bn int64
		var err error

		if strings.ContainsAny(args[i], "+-") {
			blockrange(c, args[i])
			continue
		}

		switch args[i] {
		case "pending":
			bn = seth.Pending
		case "latest":
			bn = seth.Latest
		default:
			bn, err = strconv.ParseInt(args[i], 0, 64)
			if err != nil {
				fatalf("can't parse %q: %s", args[i], err)
			}
		}
		b, err := c.GetBlock(bn, true)
		if err != nil {
			fatalf("getting block %d: %s", bn, err)
		}
		showblock(b)
	}
}
