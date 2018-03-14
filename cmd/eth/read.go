package main

import (
	"flag"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/newalchemylimited/seth"
)

var cmdread = &cmd{
	desc:  "read data from a contract",
	usage: "eth read <addr> <fn> <args>... <rettype>",
	do:    read,
}

func init() {
	cmdread.fs.Init("read", flag.ExitOnError)
	// re-use a flag from call()
	cmdread.fs.BoolVar(&forcecall, "f", false, "ignore jump table inconsistency")
}

func empty(typ string) interface{} {
	switch typ {
	case "string":
		return new(string)
	case "bool":
		return new(bool)
	case "bytes":
		return new(seth.Bytes)
	case "int8", "uint8", "int16", "uint16", "int32",
		"uint32", "int64", "uint64",
		"int128", "uint128", "int256", "uint256",
		"int", "uint":
		return big.NewInt(0)
	case "address":
		return new(seth.Address)
	}
	if strings.HasPrefix(typ, "bytes") {
		for i := 1; i <= 32; i++ {
			if typ == "bytes"+strconv.Itoa(i) {
				return new(seth.Data)
			}
		}
	}
	fatalf("unsupported return type %q", typ)
	return nil
}

func decoder(args []string) *seth.ABIDecoder {
	var d seth.ABIDecoder
	for _, a := range args {
		d = append(d, empty(a))
	}
	return &d
}

func read(fs *flag.FlagSet) {
	args := fs.Args()
	if len(args) < 2 {
		fs.Usage()
	}

	addr, err := seth.ParseAddress(args[0])
	if err != nil {
		fatalf("can't parse address: %q: %s\n", args[0], err)
	}

	c := client()
	callargs := parsefn(c, addr, args[1], args[2:])

	rettypes := args[2+len(callargs):]
	d := decoder(rettypes)

	var zeroaddr seth.Address
	s := seth.NewSender(c, &zeroaddr)
	err = s.ConstCall(addr, args[1], d, callargs...)
	if err != nil {
		fatalf("error making call: %s\n", err)
	}

	for _, v := range *d {
		switch v := v.(type) {
		case *string:
			fmt.Println(*v)
		case *bool:
			fmt.Println(*v)
		case *big.Int:
			fmt.Println(v.String())
		case *seth.Address:
			fmt.Println(v.String())
		case *seth.Data:
			fmt.Printf("%x\n", *v)
		case *seth.Bytes:
			fmt.Printf("%x\n", *v)
		default:
			fmt.Printf("%#v\n", v)
		}
	}
}
