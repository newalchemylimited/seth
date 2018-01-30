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
	fs   *flag.FlagSet
}

func fatalf(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, f, args...)
	os.Exit(1)
}

var subcommands = map[string]cmd{
	"balance": cmdbal,
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: eth <cmd> <args...>")
	fmt.Fprintln(os.Stderr, "subcommands:")
	var out []string
	for name, c := range subcommands {
		out = append(out, fmt.Sprintf("%s %s", name, c.desc))
	}
	sort.Strings(out)
	for i := range out {
		fmt.Fprintln(os.Stderr, out[i])
	}
	os.Exit(1)
}

func main() {
	args := os.Args
	if len(args) == 1 {
		usage()
	}
	cmd, ok := subcommands[args[1]]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", args[0])
		usage()
	}
	cmd.fs.Parse(args[2:])
	cmd.do(cmd.fs.Args())
}
