package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/alecthomas/template"
	"github.com/davecgh/go-spew/spew"
	"github.com/iancoleman/strcase"
	"github.com/newalchemylimited/seth"
	"golang.org/x/tools/imports"
)

//go:generate ./embedtmpl.sh

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

	var sources []seth.Source
	for i := range args {
		sources = append(sources, readfile(args[i]))
	}

	bundle, err := seth.Compile(sources)
	if err != nil {
		fatal(err)
	}

	w := bytes.NewBuffer(nil)

	err = Generate(w, opkg, bundle)
	if err != nil {
		fatal(err)
	}

	//log.Println(string(w.Bytes()))

	formattedSource, err := format.Source(w.Bytes())
	if err != nil {
		fatal(err)
	}

	buf, err := imports.Process(ofile, formattedSource, nil)
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

var funcMap = template.FuncMap{
	"CodeVar": func(code []byte) string {
		return fmt.Sprintf("%#v", code)
	},
	"FuncName": func(name string) string {
		name = strcase.ToSnake(name)
		return strcase.ToCamel(name)
	},
	"ArgName": func(name string) string {
		name = strcase.ToSnake(name)
		return strcase.ToLowerCamel(name)
	},

	"ArgNameUpper": func(name string) string {
		name = strcase.ToSnake(name)
		return strcase.ToCamel(name)
	},

	"ArgType": func(a string) string {
		if strings.HasPrefix(a, "bytes") {
			return fmt.Sprintf("[%s]byte", strings.TrimPrefix(a, "bytes"))
		}

		switch a {
		case "bool", "int8", "int16", "int32", "int64",
			"uint8", "uint16", "uint32", "uint64", "string":
			return a
		case "int", "uint", "uint128", "uint256":
			return "*big.Int"
		case "address":
			return "seth.Address"
		case "address[]":
			return "*seth.AddrSlice"
		case "uint256[]":
			return "*seth.IntSlice"
		default:
			return "seth.Data"
		}
	},
	"RetType": func(a string) string {
		if strings.HasPrefix(a, "bytes") {
			return fmt.Sprintf("[%s]byte", strings.TrimPrefix(a, "bytes"))
		}

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
		case "bytes32":
			return "[32]byte"
		default:
			return "seth.Data"
		}
	},
}

func Generate(buf io.Writer, packageName string, bundle *seth.CompiledBundle) error {

	spew.Dump(bundle.Contracts)

	for iContract := range bundle.Contracts {
		contract := &bundle.Contracts[iContract]
		for iDescriptor := range contract.ABI {
			descriptor := &contract.ABI[iDescriptor]

			if descriptor.Type != "function" {
				continue
			}

			for iInput := range descriptor.Inputs {
				input := &descriptor.Inputs[iInput]
				if input.Name == "" {
					input.Name = fmt.Sprintf("arg%d", iInput)
				}
			}

			for iOutput := range descriptor.Outputs {
				output := &descriptor.Outputs[iOutput]
				if output.Name == "" {
					output.Name = fmt.Sprintf("ret%d", iOutput)
				}
			}

		}

	}

	tmpl, err := template.New(".").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(buf, map[string]interface{}{
		"package":   packageName,
		"contracts": bundle.Contracts,
	}); err != nil {
		return err
	}

	return nil
}
