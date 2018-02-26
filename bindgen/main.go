package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/newalchemylimited/seth"
	"golang.org/x/tools/imports"
)

func fatal(j interface{}) {
	fmt.Fprintln(os.Stderr, j)
	os.Exit(1)
}

func usage() {
	fatal("usage: bindgen -c<contract> <files ...>")
}

var cstr string
var ofile string
var opkg string
var bin bool

func init() {
	flag.StringVar(&cstr, "c", "", "contracts for which to generate code")
	flag.StringVar(&ofile, "o", "", "output file")
	flag.StringVar(&opkg, "p", os.Getenv("GOPACKAGE"), "output package name")
	flag.BoolVar(&bin, "b", false, "output binary as a go variable")
}

func readfile(v string) seth.Source {
	buf, err := ioutil.ReadFile(v)
	if err != nil {
		fatal(err)
	}
	return seth.Source{
		Filename: v,
		Body:     string(buf),
	}
}

func outfile() io.WriteCloser {
	if ofile != "" {
		f, err := os.Create(ofile)
		if err != nil {
			fatal(err)
		}
		return f
	}
	return os.Stdout
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		usage()
	}

	contracts := strings.Split(cstr, ",")

	var sources []seth.Source
	for i := range args {
		sources = append(sources, readfile(args[i]))
	}

	bundle, err := seth.Compile(sources)
	if err != nil {
		fatal(err)
	}

	w := bytes.NewBuffer(nil)

	fmt.Fprintf(w, "package %s\n\n", opkg)
	fmt.Fprintf(w, "import (\n\t\"github.com/newalchemylimited/seth\"\n)\n\n")

	for i := range contracts {
		c := bundle.Contract(contracts[i])
		if c == nil {
			fatal("no such contract " + contracts[i])
		}
		generate(w, c)
	}

	buf, err := imports.Process(ofile, w.Bytes(), nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, w.String())
		fatal("goimports: " + err.Error())
	}

	f := outfile()
	_, err = f.Write(buf)
	if err != nil {
		fatal(err)
	}
	f.Close()
}

func typeconv(a string) string {
	switch a {
	case "bool", "int8", "int16", "int32", "int64",
		"uint8", "uint16", "uint32", "uint64":
		fallthrough
	case "int", "uint", "uint128", "uint256":
		return "*seth.Int"
	case "address":
		return "*seth.Address"
	case "address[]":
		return "*seth.AddrSlice"
	case "uint256[]":
		return "*seth.IntSlice"
	default:
		return "seth.Data"
	}
}

func rettype(a string) string {
	switch a {
	case "bool", "int8", "int16", "int32", "int64",
		"uint8", "uint16", "uint32", "uint64":
		return a
	case "int", "uint", "uint128", "uint256":
		return "seth.Int"
	case "address":
		return "seth.Address"
	case "address[]":
		return "seth.AddrSlice"
	case "uint256[]":
		return "seth.IntSlice"
	case "string":
		return "string"
	case "bytes":
		return "[]byte"
	default:
		return "seth.Data"
	}
}

func deref(v string) string {
	if len(v) > 0 && v[0] == '*' {
		return v[1:]
	}
	return v
}

func generate(w io.Writer, c *seth.CompiledContract) {
	if bin {
		fmt.Fprintf(w, "var %sCode = %#v\n", c.Name, c.Code)
	}

	// type decl
	fmt.Fprintf(w, "type %s struct {\n", c.Name)
	fmt.Fprintln(w, "\taddr  *seth.Address")
	fmt.Fprintln(w, "\ts     *seth.Sender\n}")
	fmt.Fprintln(w)

	// constructor
	fmt.Fprintf(w, "func New%s(addr *seth.Address, sender *seth.Sender) *%[1]s {\n", c.Name)
	fmt.Fprintf(w, "\treturn &%s{addr: addr, s: sender}\n", c.Name)
	fmt.Fprintln(w, "}")

	// methods
	for i := range c.ABI {
		d := &c.ABI[i]
		if d.Type != "function" {
			continue
		}
		// don't do anything for pointless functions
		if d.Constant && len(d.Outputs) == 0 {
			continue
		}

		fmt.Fprintln(w)

		fmt.Fprintf(w, "func (z *%s) %s(", c.Name, strings.Title(d.Name))
		// input arguments
		var argstrs []string
		for i := range d.Inputs {
			argstrs = append(argstrs, fmt.Sprintf("arg%d %s", i, typeconv(d.Inputs[i].Type)))
		}
		fmt.Fprint(w, strings.Join(argstrs, ", ")+") ")

		if d.Constant {
			var retargs []string
			for i := range d.Outputs {
				// TODO: handle composite types (tuples, structs)
				retargs = append(retargs, fmt.Sprintf("ret%d %s", i, rettype(d.Outputs[i].Type)))
			}
			retargs = append(retargs, "err error")
			fmt.Fprintln(w, "("+strings.Join(retargs, ", ")+") {")

			// body

			retargs = retargs[:0]
			for i := 0; i < len(d.Outputs); i++ {
				retargs = append(retargs, fmt.Sprintf("&ret%d", i))
			}

			fmt.Fprintln(w, "\td := seth.NewABIDecoder("+strings.Join(retargs, ", ")+")")

			fmt.Fprintf(w, "\terr = z.s.ConstCall(z.addr, %q, d", d.Signature())
			for i := 0; i < len(d.Inputs); i++ {
				fmt.Fprintf(w, ", arg%d", i)
			}
			fmt.Fprintln(w, ")")
			fmt.Fprintln(w, "\treturn")
			fmt.Fprintln(w, "}")
		} else {
			fmt.Fprintln(w, "(seth.Hash, error) {")
			fmt.Fprintf(w, "\treturn z.s.Send(z.addr, %q", d.Signature())
			for i := 0; i < len(d.Inputs); i++ {
				fmt.Fprintf(w, ", arg%d", i)
			}
			fmt.Fprintln(w, ")")
			fmt.Fprintln(w, "}")
		}
	}
}
