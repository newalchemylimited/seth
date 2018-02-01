package main

import (
	"fmt"
)

var cmdaccounts = &cmd{
	desc: "show available accounts",
	do:   accounts,
}

func accounts(args []string) {
	kds := keys()
	for _, kd := range kds {
		uuid := "none"
		if kd.kf != nil {
			uuid = kd.kf.ID
		}
		fmt.Println(uuid, kd.addr.String())
	}
}
