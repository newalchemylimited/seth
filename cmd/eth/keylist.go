package main

import (
	"flag"
	"fmt"
)

var cmdkeylist = &cmd{
	desc:  "show available key files",
	usage: "eth keys",
	do:    keylist,
}

func init() {
	cmdkeylist.fs.Init("keys", flag.ExitOnError)
}

func keylist(fs *flag.FlagSet) {
	if fs.NArg() != 0 {
		fs.Usage()
	}
	kds := keys()
	for _, kd := range kds {
		fmt.Printf("%36s %s\n", kd.id, kd.addr.String())
	}
}
