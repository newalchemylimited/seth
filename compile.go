package seth

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/newalchemylimited/seth/keccak"
	"github.com/tinylib/msgp/msgp"
)

type sourcefile struct {
	Content string `json:"content"`
}

type sourceblob struct {
	Language string                `json:"language"`
	Sources  map[string]sourcefile `json:"sources"`
	Settings struct {
		Optimizer struct {
			Enabled bool `json:"enabled"`
			Runs    int  `json:"runs"`
		} `json:"optimizer"`
		OutputSelection map[string]map[string][]string `json:"outputSelection"`
	} `json:"settings"`
	OutputSelection map[string]map[string][]string `json:"outputSelection"`
}

var defaultOutput = map[string]map[string][]string{
	"*": map[string][]string{
		"*": []string{"evm.bytecode", "abi"},
	},
}

type solcerror struct {
	SourceLocation struct {
		File  string `json:"file"`
		Start int    `json:"start"`
		End   int    `json:"end"`
	} `json:"sourceLocation"`
	Type             string `json:"type"`
	Component        string `json:"component"`
	Message          string `json:"message"`
	FormattedMessage string `json:"formattedMessage"`
}

func (s *solcerror) Error() string { return s.FormattedMessage }

type sourcedesc struct {
	ID int `json:"id"`
}

type contractout struct {
	ABI []ABIDescriptor
	EVM struct {
		Bytecode struct {
			Object    string `json:"object"`    // hex string of opcodes
			Sourcemap string `json:"sourceMap"` // source map string
		} `json:"bytecode"`
	} `json:"evm"`
}

type solcout struct {
	Errors    []solcerror                       `json:"errors"`
	Sources   map[string]sourcedesc             `json:"sources"`
	Contracts map[string]map[string]contractout `json:"contracts"` // map[sourcefile]map[contractname]
}

// represents a "s:l:f:j" group in a sourcemap
type srcinfo struct {
	s int   // starting position
	l int   // byte length
	f int   // file id
	j uint8 // one of 'i', 'o', or '-'
}

// ABIParam describes an input or output parameter
// to an ABIDescriptor
type ABIParam struct {
	Name       string     `json:"name" msg:"name"`
	Type       string     `json:"type" msg:"type"`
	Components []ABIParam `json:"components,omitempty" msg:"componenets"` // for tuples, nested parameters
	Indexed    bool       `json:"indexed" msg:"indexed"`                  // for events, whether or not the parameter is indexed
}

// ABIDescriptor describes a function, constructor, or event
type ABIDescriptor struct {
	// one of "function" "constructor" "fallback" "event"
	Type       string     `json:"type" msg:"type"`
	Name       string     `json:"name" msg:"name"`
	Inputs     []ABIParam `json:"inputs" msg:"inputs"`
	Outputs    []ABIParam `json:"outputs,omitempty" msg:"outputs"`
	Payable    bool       `json:"payable" msg:"payable"`
	Mutability string     `json:"stateMutability" msg:"stateMutability"` // "pure", "view", "nonpayable", "payable"
	Constant   bool       `json:"constant" msg:"constant"`               // either "pure" or "view"
	Anonymous  bool       `json:"anonymous" msg:"anonymous"`
}

// CompiledContract represents a single solidity contract.
type CompiledContract struct {
	Name      string          `msg:"name"`      // Contract name
	Code      []byte          `msg:"code"`      // Code is the EVM bytecode for a contract
	Sourcemap string          `msg:"sourcemap"` // Sourcemap is the stringified source map for the contract
	ABI       []ABIDescriptor `msg:"abi"`       // Raw JSON ABI

	srcmap []srcinfo
	pos    []int // pos[pc] = opcode number
}

var metadataPrefix = []byte{0xa1, 0x65, 'b', 'z', 'z', 'r', '0', 0x58, 0x20}
var metadataSuffix = []byte{0x00, 0x29}

// MetadataHash tries to return the portion of a
// compiled solidity contract that represents the
// hash of the metadata. If the metadata can't be
// found, a nil slice is returned.
func MetadataHash(b []byte) []byte {
	// see: http://solidity.readthedocs.io/en/develop/metadata.html#encoding-of-the-metadata-hash-in-the-bytecode
	mp := len(metadataPrefix)
	ms := len(metadataSuffix)
	mlen := 32 + mp + ms
	if len(b) < mlen {
		return nil
	}
	suffix := b[len(b)-mlen:]
	if !bytes.Equal(suffix[:mp], metadataPrefix) ||
		!bytes.Equal(suffix[mp+32:], metadataSuffix) {
		return nil
	}
	return suffix[mp : mp+32]
}

// StripBytecode returns the portion of the bytecode
// that isn't solidity metadata. If no solidity metadata
// is present, the argument is returned unchanged.
func StripBytecode(b []byte) []byte {
	if MetadataHash(b) != nil {
		return b[:len(b)-43]
	}
	return b
}

func mustint(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func (c *CompiledContract) pc2info(pc int) *srcinfo {
	if pc >= len(c.pos) {
		return nil
	}
	n := c.pos[pc]
	if n >= len(c.srcmap) {
		return nil
	}
	return &c.srcmap[n]
}

func (c *CompiledContract) compilePos() {
	opnum := 0
	var out []int
	for i := range c.Code {
		width := 0
		b := c.Code[i]
		out = append(out, opnum)
		if b >= 0x60 && b < 0x80 {
			width = int(b - 0x59)
		}
		// for multi-byte instructions,
		// any pc that points into the instruction
		// gets the same opcode number
		for j := 0; j < width; j++ {
			out = append(out, opnum)
		}
		opnum++
	}
	c.pos = out
}

func (c *CompiledContract) compileSourcemap() {
	ops := strings.Split(c.Sourcemap, ";")
	out := make([]srcinfo, len(ops))
	for i := range out {
		if i > 0 {
			out[i] = out[i-1]
		}
		sections := strings.Split(ops[i], ":")
		switch len(sections) {
		case 4:
			if len(sections[3]) > 0 {
				out[i].j = sections[3][0]
			}
			fallthrough
		case 3:
			if len(sections[2]) > 0 {
				out[i].f = mustint(sections[2])
			}
			fallthrough
		case 2:
			if len(sections[1]) > 0 {
				out[i].l = mustint(sections[1])
			}
			fallthrough
		case 1:
			if len(sections[0]) > 0 {
				out[i].s = mustint(sections[0])
			}
		}
	}
	c.srcmap = out
}

//go:generate msgp

// CompiledBundle represents the output of a single invocation of solc.
// A CompiledBundle contains zero or more contracts.
type CompiledBundle struct {
	Filenames []string           `msg:"filenames"` // Input filenames in sourcemap ID order
	Sources   []string           `msg:"sources"`   // actual source code, in filename order
	Contracts []CompiledContract `msg:"contracts"` // Compiled code
	Warnings  []string           `msg:"warnings"`
}

func (c *CompiledBundle) Contract(name string) *CompiledContract {
	for i := range c.Contracts {
		if c.Contracts[i].Name == name {
			return &c.Contracts[i]
		}
	}
	return nil
}

func h2b(s string) []byte {
	bits, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bits
}

func (s *solcout) toBundle() *CompiledBundle {
	nc := 0
	for _, cs := range s.Contracts {
		nc += len(cs)
	}
	b := &CompiledBundle{
		Filenames: make([]string, len(s.Contracts)),
		Sources:   make([]string, len(s.Contracts)),
	}
	// XXX: assume IDs are 0-indexed...
	for name, id := range s.Sources {
		b.Filenames[id.ID] = name
	}
	for _, contracts := range s.Contracts {
		for contract, out := range contracts {
			b.Contracts = append(b.Contracts, CompiledContract{
				Name:      contract,
				Code:      h2b(out.EVM.Bytecode.Object),
				Sourcemap: out.EVM.Bytecode.Sourcemap,
				ABI:       out.ABI,
			})
		}
	}
	return b
}

// SourceOf returns the source line (text) of the given pc.
//
// NOTE: right now solc creates terrible source maps. You may
// get the entire contract code back in the source string.
func (b *CompiledBundle) SourceOf(c *CompiledContract, pc int) string {
	if len(c.srcmap) == 0 {
		c.compileSourcemap()
		c.compilePos()
	}
	info := c.pc2info(pc)
	if info == nil {
		return ""
	}
	return b.Sources[info.f][info.s : info.s+info.l]
}

func compileError(errors []solcerror) error {
	for i := range errors {
		if errors[i].Type != "Warning" {
			return &errors[i]
		}
	}
	return nil
}

type Source struct {
	Filename string
	Body     string
}

func short(h hash.Hash) string {
	return hex.EncodeToString(h.Sum(nil))[:12]
}

func cachename(sources []Source) string {
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Filename < sources[j].Filename
	})
	h0 := keccak.New256()
	h1 := keccak.New256()
	for i := range sources {
		io.WriteString(h0, sources[i].Filename)
		io.WriteString(h1, sources[i].Body)
	}
	// filename is 'name hash'-'content hash'.bundle
	// so that we can to a quick glob search to remove
	// any stale bundles with the same name hash
	base := short(h0) + "-" + short(h1) + ".bundle"
	return filepath.Join(os.TempDir(), "tevm", base)
}

func cachedBundle(name string) (*CompiledBundle, bool) {
	f, err := os.Open(name)
	if err != nil {
		// let's look for stale builds and clean them up
		i := strings.IndexByte(name, '-')
		if i != -1 && i != len(name)-1 {
			all, err := filepath.Glob(name[:i+1] + "*")
			if err == nil {
				for i := range all {
					os.Remove(all[i])
				}
			}
		}
		return nil, false
	}
	defer f.Close()
	b := new(CompiledBundle)
	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, false
	}
	d := msgp.NewReader(r)
	if err := b.DecodeMsg(d); err != nil {
		return nil, false
	}
	return b, true
}

func writeCache(name string, bundle *CompiledBundle) {
	dir := filepath.Dir(name)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return
	}
	f, err := os.Create(name)
	if err != nil {
		return
	}
	w := gzip.NewWriter(f)
	e := msgp.NewWriter(w)
	bundle.EncodeMsg(e)
	e.Flush()
	w.Close()
	f.Close()
}

// Compile sources into a bundle with caching.
func Compile(sources []Source) (*CompiledBundle, error) {
	cn := cachename(sources)
	if bundle, ok := cachedBundle(cn); ok {
		return bundle, nil
	}
	b, err := compile(sources)
	if err != nil {
		return nil, err
	}
	writeCache(cn, b)
	return b, nil
}

// compile sources without caching.
func compile(sources []Source) (*CompiledBundle, error) {
	cmd := exec.Command("solc", "--standard-json")
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	blob := sourceblob{
		Language: "Solidity",
	}
	blob.Sources = make(map[string]sourcefile, len(sources))
	for i := range sources {
		blob.Sources[sources[i].Filename] = sourcefile{Content: sources[i].Body}
	}
	blob.Settings.Optimizer.Enabled = true
	blob.Settings.Optimizer.Runs = 200
	blob.Settings.OutputSelection = defaultOutput
	blob.OutputSelection = defaultOutput
	err = json.NewEncoder(in).Encode(&blob)
	if err != nil {
		in.Close()
		cmd.Wait()
		return nil, err
	}
	in.Close()
	var jsout solcout
	err = json.NewDecoder(out).Decode(&jsout)
	if err != nil {
		cmd.Wait()
		return nil, fmt.Errorf("reading solc output: %s", err)
	}
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("solc: %s", err)
	}
	if err := compileError(jsout.Errors); err != nil {
		return nil, fmt.Errorf("solc compilation error: %s", err)
	}
	b := jsout.toBundle()
	for i := range jsout.Errors {
		if jsout.Errors[i].Type == "Warning" {
			b.Warnings = append(b.Warnings, jsout.Errors[i].FormattedMessage)
		}
	}
	for i := range b.Sources {
		b.Sources[i] = sources[i].Body
	}
	return b, nil
}

// CompileGlob compiles all the files that
// match the given filepath glob (e.g. "*.sol")
func CompileGlob(glob string) (*CompiledBundle, error) {
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("glob %q matches zero files", matches)
	}
	var body []byte
	sources := make([]Source, len(matches))
	for i := range matches {
		sources[i].Filename = matches[i]
		body, err = ioutil.ReadFile(matches[i])
		if err != nil {
			return nil, err
		}
		sources[i].Body = string(body)
	}
	return Compile(sources)
}

// CompileString compiles source code from a string.
func CompileString(code string) (*CompiledBundle, error) {
	return compile([]Source{{
		Filename: "<stdin>",
		Body:     code,
	}})
}
