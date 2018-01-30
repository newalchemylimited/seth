package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/newalchemylimited/seth"
)

var cmdblock = &cmd{
	desc: "print a block as json",
	do:   block,
}

func block(args []string) {
	if len(args) == 0 {
		args = append(args, "pending")
	}

	c := client()

	for i := range args {
		var bn int64
		var err error
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
		buf, err := json.MarshalIndent(b, "", "\t")
		if err != nil {
			fatalf("fatal error: %s", err)
		}
		fmt.Printf("%s\n", buf)
	}
}
