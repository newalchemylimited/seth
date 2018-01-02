package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/newalchemylimited/seth/tevm"
)

var addr string

func init() {
	flag.StringVar(&addr, "a", ":8043", "bind address to listen on")
}

func main() {
	flag.Parse()
	c := tevm.NewChain()
	acct := c.NewAccount(10)
	log.Println("default account:", acct.String())
	log.Printf("binding to %s...", addr)
	log.Fatal(http.ListenAndServe(addr, c))
}
