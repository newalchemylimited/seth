package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/newalchemylimited/seth"
	"github.com/newalchemylimited/seth/tevm"
)

var addr string

func init() {
	flag.StringVar(&addr, "a", ":8043", "bind address to listen on")
}

func main() {
	flag.Parse()

	var c *tevm.Chain
	args := flag.Args()
	if len(args) > 0 && args[0] == "fork" {
		if len(args) != 2 {
			log.Fatalln("expected args 'tevmd fork <blocknum>'")
		}
		bn, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			log.Fatalf("can't parse %q as block number: %s", args[1], err)
		}
		log.Println("forking main chain at block %d", bn)
		c = tevm.NewFork(seth.NewClientTransport(seth.InfuraTransport{}), bn)
	} else {
		c = tevm.NewChain()
	}

	acct := c.NewAccount(10)
	log.Println("default account:", acct.String())
	log.Printf("binding to %s...", addr)
	log.Fatal(http.ListenAndServe(addr, c))
}
