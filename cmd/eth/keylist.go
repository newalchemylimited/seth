package main

import (
	"fmt"
)

var cmdkeylist = &cmd{
	desc: "show available key files",
	do:   keylist,
}

func keylist(args []string) {
	kds := keys()
	for _, kd := range kds {
		fmt.Printf("%36s %s\n", kd.id, kd.addr.String())
	}
}
