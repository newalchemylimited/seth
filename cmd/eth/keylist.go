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
		uuid := "none"
		if kd.kf != nil {
			uuid = kd.kf.ID
		}
		fmt.Println(uuid, kd.addr.String())
	}
}
