package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/newalchemylimited/seth"
)

func client() *seth.Client {
	url := os.Getenv("SETH_URL")
	if url == "" {
		return seth.NewClient(seth.IPCDial)
	}
	if strings.HasPrefix(url, "http") {
		var t seth.Transport
		if strings.Contains(url, "api.infura.io") {
			t = seth.InfuraTransport{}
		} else {
			t = &seth.HTTPTransport{URL: url}
		}
		return seth.NewClientTransport(t)
	}
	if _, err := os.Stat(url); err == nil {
		return seth.NewClient(seth.IPCPath(url))
	}
	fmt.Fprintln(os.Stderr, "cannot derive client from SETH_URL=%q", url)
	os.Exit(1)
	return nil
}

func latest(c *seth.Client) int64 {
	b, err := c.GetBlock(seth.Latest, false)
	if err != nil {
		fatalf("getting block: %s\n", err)
	}
	return int64(*b.Number)
}

func pending(c *seth.Client) int64 {
	b, err := c.GetBlock(seth.Pending, false)
	if err != nil {
		fatalf("getting block: %s\n", err)
	}
	return int64(*b.Number)

}
