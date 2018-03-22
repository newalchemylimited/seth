package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/newalchemylimited/seth"
)

var cmdcall = &cmd{
	desc:  "call a function",
	usage: "eth call <addr> <fn> <args ...>",
	do:    call,
}

var forcecall bool
var noncecall int
var gweicall int

func init() {
	cmdcall.fs.Init("call", flag.ExitOnError)
	cmdcall.fs.BoolVar(&forcecall, "f", false, "force call (avoid checking jump-table)")
	cmdcall.fs.IntVar(&noncecall, "n", -1, "call nonce")
	cmdcall.fs.IntVar(&gweicall, "g", 4, "gas price (gwei)")
}

func etherstring(s string) seth.EtherType {
	return etherbytes(s)
}

func unhex(s string) []byte {
	if len(s) >= 2 && s[:2] == "0x" {
		s = s[2:]
	}
	buf, err := hex.DecodeString(s)
	if err != nil {
		fatalf("decoding %q: %s\n", s, err)
	}
	return buf
}

func etherbytes(s string) seth.EtherType {
	b := seth.Bytes(unhex(s))
	return &b
}

func etheraddr(s string) seth.EtherType {
	addr, err := seth.ParseAddress(s)
	if err != nil {
		fatalf("parsing %q as address: %s\n", s, err)
	}
	return addr
}

func etherdata(typ, s string) seth.EtherType {
	width, err := strconv.Atoi(strings.TrimPrefix(typ, "bytes"))
	if err != nil {
		fatalf("internal error: bad bytes type %q\n", typ)
	}
	buf := unhex(s)
	if len(buf) != width {
		fatalf("%q is length %d (inappropriate for bytes%d)", s, len(buf), width)
	}
	d := seth.Data(buf)
	return &d
}

func etherint(s string, bitwidth int, signed bool) seth.EtherType {
	var val big.Int
	v, ok := val.SetString(s, 0)
	if !ok {
		fatalf("can't parse %q as integer\n", s)
	}
	if v.Sign() < 0 && !signed {
		fatalf("value %q not allowed to be signed\n", s)
	}
	if v.BitLen() >= bitwidth {
		fatalf("value %q overflows bit-width %d", s, bitwidth)
	}
	return (*seth.Int)(v)
}

func etherarr(typ string, val string) seth.EtherType {
	if val[0] != '[' || val[len(val)-1] != ']' {
		fatalf("array values must be enclosed with square braces\n")
	}
	parts := strings.Split(val, ",")
	if len(parts) == 0 {
		// doesn't really matter what type this is...
		s := make(seth.IntSlice, 0)
		return &s
	}

	// Warning: this is super gross.
	// This should be replaced with something better.
	args := make([]seth.EtherType, len(parts))
	for i := range parts {
		args[i] = parsearg(typ, parts[i])
	}
	switch t := args[0].(type) {
	case *seth.Int:
		out := make(seth.IntSlice, len(args))
		for i := range args {
			out[i] = *(args[i].(*seth.Int))
		}
		return &out
	case *seth.Address:
		out := make(seth.AddrSlice, len(args))
		for i := range args {
			out[i] = *(args[i].(*seth.Address))
		}
		return &out
	default:
		fatalf("slice of %T unimplemented\n", t)
	}
	return nil
}

func parsearg(typ, val string) seth.EtherType {
	switch typ {
	case "string":
		return etherstring(val)
	case "bool":
		switch val {
		case "true":
			return seth.NewInt(1)
		case "false":
			return seth.NewInt(0)
		default:
			fatalf("%q not a boolean\n", val)
			return nil
		}
	case "bytes":
		return etherbytes(val)
	case "int8", "uint8":
		return etherint(val, 8, val == "int8")
	case "int16", "uint16":
		return etherint(val, 16, val == "int16")
	case "int32", "uint32":
		return etherint(val, 32, val == "int32")
	case "int64", "uint64":
		return etherint(val, 64, val == "int64")
	case "int128", "uint128":
		return etherint(val, 128, val == "int128")
	case "int256", "uint256":
		return etherint(val, 256, val == "int256")
	case "address":
		return etheraddr(val)
	}
	if strings.HasPrefix(typ, "bytes") {
		for i := 1; i <= 32; i++ {
			if typ == "bytes"+strconv.Itoa(i) {
				return etherdata(typ, val)
			}
		}
	}
	if strings.HasSuffix(typ, "[]") {
		return etherarr(strings.TrimSuffix(typ, "[]"), val)
	}
	fatalf("don't know how to parse %q as %q", val, typ)
	return nil
}

func parsefn(c *seth.Client, addr *seth.Address, fn string, args []string) []seth.EtherType {
	lparen, rparen := strings.IndexByte(fn, '('), strings.IndexByte(fn, ')')
	if lparen == -1 || rparen == -1 ||
		rparen-lparen < 1 ||
		rparen != len(fn)-1 {
		fatalf("bad function signature %q (bad parens)\n", args[1])
	}

	fname := fn[:lparen]
	if fname == "" && lparen != rparen-1 {
		fatalf("fallback function can't have arguments\n")
	}

	// check that the function we're calling is
	// actually present in the jump table
	if c != nil && !forcecall && fname != "" {
		h := seth.HashString(fn)
		entries := jumpentries(getcode(c, addr))
		found := false
		for i := range entries {
			if bytes.Equal(entries[i].prefix[:], h[:4]) {
				found = true
				break
			}
		}
		if !found {
			fatalf("signature %q (jump table %x) not found in code\n", args[1], h[:4])
		}
	}

	arglist := fn[lparen+1 : rparen]
	if arglist == "" {
		return make([]seth.EtherType, 0)
	}
	argtypes := strings.Split(arglist, ",")
	if len(args) < len(argtypes) {
		fatalf("signature wants %d arguments, but %d were provided\n", len(argtypes), len(args))
	}

	callargs := make([]seth.EtherType, len(argtypes))
	for i := range argtypes {
		callargs[i] = parsearg(argtypes[i], args[i])
	}
	return callargs
}

func call(fs *flag.FlagSet) {
	args := fs.Args()
	if len(args) < 2 {
		fs.Usage()
		fatalf("usage: eth call <address> <sig> <args...>\n")
	}

	addr, err := seth.ParseAddress(args[0])
	if err != nil {
		fatalf("can't parse address %q: %s\n", args[0], err)
	}

	c := client()
	callargs := parsefn(c, addr, args[1], args[2:])

	sign, from := signer()
	opts := seth.CallOpts{
		From:     &from,
		To:       addr,
		GasPrice: seth.NewInt(int64(gweicall) * 1e9),
	}
	if noncecall >= 0 {
		u := seth.Uint64(noncecall)
		opts.Nonce = &u
	}
	opts.EncodeCall(args[1], callargs...)

	s := seth.NewSender(c, &from)
	s.Signer = sign

	h, err := s.Call(&opts)
	if err != nil {
		fatalf("failed to send transaction: %s\n", err)
	}
	fmt.Println(h.String())
}
