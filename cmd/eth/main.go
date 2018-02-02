package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
)

type cmd struct {
	desc string              // command description
	do   func(args []string) // do it
	fs   flag.FlagSet
}

func fatalf(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, f, args...)
	os.Exit(1)
}

var subcommands = map[string]*cmd{
	"balance":  cmdbal,
	"block":    cmdblock,
	"accounts": cmdaccounts,
}

func usage() {
	fmt.Println("usage: eth <cmd> <args...>")
	fmt.Println("subcommands:")
	var out [][2]string
	for name, c := range subcommands {
		out = append(out, [2]string{name, c.desc})
	}
	sort.Slice(out, func(i, j int) bool { return out[i][0] < out[j][0] })
	for i := range out {
		fmt.Printf("%16s    %s\n", out[i][0], out[i][1])
	}
	os.Exit(1)
}

func main() {
	args := os.Args
	if len(args) == 1 || args[1] == "help" {
		usage()
	}
	cmd, ok := subcommands[args[1]]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", args[1])
		usage()
	}
	cmd.fs.Parse(args[2:])
	cmd.do(cmd.fs.Args())
}
