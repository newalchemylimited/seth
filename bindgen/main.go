package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/newalchemylimited/seth"
)

var destructors []func()

func atexit(f func()) {
	destructors = append(destructors, f)
}

func destroy() {
	for i := range destructors {
		destructors[i]()
	}
}

func fatal(j interface{}) {
	fmt.Fprintln(os.Stderr, j)
	destroy()
	os.Exit(1)
}

func usage() {
	fatal("usage: bindgen -c<contract> <files ...>")
}

var cstr string
var ofile string
var opkg string

func init() {
	flag.StringVar(&cstr, "c", "", "contracts for which to generate code")
	flag.StringVar(&ofile, "o", "", "output file")
	flag.StringVar(&opkg, "p", os.Getenv("GOPACKAGE"), "output package name")
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
	defer destroy()
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

	o := outfile()
	atexit(func() { o.Close() })
	w := bufio.NewWriter(o)

	fmt.Fprintf(w, "package %s\n\n", opkg)
	fmt.Fprintf(w, "import (\n\t\"github.com/newalchemylimited/seth\"\n)\n\n")

	for i := range contracts {
		c := bundle.Contract(contracts[i])
		if c == nil {
			fatal("no such contract " + contracts[i])
		}
		generate(w, c)
	}
	if err := w.Flush(); err != nil {
		fatal(err)
	}
}

func typeconv(a string) string {
	switch a {
	case "bool", "int8", "int16", "int32", "int64",
		"uint8", "uint16", "uint32", "uint64":
		// TODO: something better here;
		// we can just use native go types
		// if we do a little more work here
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

func deref(v string) string {
	if len(v) > 0 && v[0] == '*' {
		return v[1:]
	}
	return v
}

func generate(w io.Writer, c *seth.CompiledContract) {
	// type decl
	fmt.Fprintf(w, "type %s struct {\n", c.Name)
	fmt.Fprintln(w, "\taddr  *seth.Address")
	fmt.Fprintln(w, "\ts     *seth.Sender\n}\n")

	// constructor
	fmt.Fprintf(w, "func New%s(addr *seth.Address, sender *seth.Sender) *%[1]s {\n", c.Name)
	fmt.Fprintf(w, "\treturn &%s{addr: addr, s: sender}\n", c.Name)
	fmt.Fprintln(w, "}\n")

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

		fmt.Fprintf(w, "func (z *%s) %s(", c.Name, strings.Title(d.Name))
		// input arguments
		var argstrs []string
		for i := range d.Inputs {
			argstrs = append(argstrs, fmt.Sprintf("arg%d %s", i, typeconv(d.Inputs[i].Type)))
		}
		fmt.Fprint(w, strings.Join(argstrs, ", ")+") ")

		if d.Constant {
			// output arguments: for now, just punt
			// on multiple returns, but try to do
			// something sane for single-return values
			var rettype string
			if len(d.Outputs) == 1 {
				// sane
				rettype = deref(typeconv(d.Outputs[0].Type))
			} else {
				rettype = "json.RawMessage"
			}
			fmt.Fprintf(w, "(ret %s, err error) {\n", rettype)

			// body
			fmt.Fprintf(w, "\terr = z.s.ConstCall(z.addr, %q, &ret", d.Signature())
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
