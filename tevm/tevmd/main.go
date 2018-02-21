package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/newalchemylimited/seth"
	"github.com/newalchemylimited/seth/tevm"
)

var addr string
var src string
var verbose bool

func init() {
	flag.StringVar(&addr, "a", ":8043", "bind address to listen on")
	flag.StringVar(&src, "e", "", "chain source (path or url)")
	flag.BoolVar(&verbose, "v", false, "be verbose")
}

func client() *seth.Client {
	var client *seth.Client
	if src == "" {
		client = seth.NewClientTransport(seth.InfuraTransport{})
	} else if strings.HasPrefix(src, "http") {
		client = seth.NewClientTransport(&seth.HTTPTransport{URL: src})
	} else {
		if _, err := os.Stat(src); err != nil {
			log.Fatal(err)
		}
		client = seth.NewClient(seth.IPCPath(src))
	}
	return client
}

func latest(c *seth.Client) int64 {
	blk, err := c.Latest(false)
	if err != nil {
		log.Fatal(err)
	}
	return int64(*blk.Number)
}

func main() {
	flag.Parse()

	var c *tevm.Chain
	args := flag.Args()
	if len(args) > 0 && args[0] == "fork" {
		network := client()
		var err error
		var bn int64
		switch len(args) {
		case 2:
			bn, err = strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				log.Fatalln("bad block number:", err)
			}
		case 1:
			bn = latest(network)
		default:
			log.Fatalln("expected args 'tevmd fork <optional: blocknum>'")
		}
		log.Printf("forking main chain at block %d", bn)
		c = tevm.NewFork(network, bn)
	} else {
		c = tevm.NewChain()
	}

	if verbose {
		c.Debugf = log.Printf
	}

	acct := c.NewAccount(10)
	log.Println("default account:", acct.String())
	log.Printf("binding to %s...", addr)
	log.Fatal(http.ListenAndServe(addr, c))
}
