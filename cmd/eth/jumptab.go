package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/newalchemylimited/seth"
)

var cmdjumptab = &cmd{
	desc: "print the jump table of a contract",
	do:   jumptab,
}

func init() {
	cmdjumptab.fs.Init("jumptab", flag.ExitOnError)
}

// just the opcodes we need to parse the jump table
const (
	opdup1  = 0x80
	oppush1 = 0x60
	opeq    = 0x14
	opjumpi = 0x57
)

type jmpentry struct {
	prefix  [4]byte // jump table prefix
	jmpdest int     // PC of actual code
}

// preimage is a list of common function selectors
var preimage = []string{
	"collect()",
	"collect(address)",
	"acceptOwnership()",
	"allowance(address,address)",
	"approve(address,uint256)",
	"approveAndCall(address,uint256,bytes)",
	"balanceOf(address)",
	"burn(address,uint256)",
	"changeOwner(address)",
	"decimals()",
	"deposit()",
	"finalize()",
	"finalized()",
	"halt()",
	"locked()",
	"makeWallet()",
	"mint(address,uint256)",
	"name()",
	"owner()",
	"pause()",
	"setPrice(uint256)",
	"start()",
	"sweep(address,uint256)",
	"symbol()",
	"tokenFallback(address,uint256,bytes)",
	"totalSupply()",
	"transfer(address,uint256)",
	"transferFrom(address,address,uint256)",
	"version()",
}

// code sequences that are equivalent to
//   (calldata[0] >> 224)
var prefixes = []string{
	// PUSH1 0x0 CALLDATALOAD PUSH29
	// 0x100000000000000000000000000000000000000000000000000000000
	// SWAP1 DIV PUSH4 0xFFFFFFFF AND
	"6000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16",

	// same as above, with AND produced before instead of after calldata load
	"63ffffffff6000357c0100000000000000000000000000000000000000000000000000000000900416",

	// ... and another permutation of the above
	// PUSH4 0xffffffff PUSH29 (1<<224) PUSH1 0x0 CALLDATALOAD DIV AND
	"63ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416",

	// like the first one, but omitting the superfluous AND
	"6000357c01000000000000000000000000000000000000000000000000000000009004",

	// PUSH4 0xffffffff
	// PUSH1 0xe0 PUSH1 0x02 EXP
	// PUSH1 0x00 CALLDATALOAD
	// DIV AND
	"63ffffffff60e060020a6000350416",

	// same as above, without superflous AND:
	// PUSH1 0xe0 PUSH1 0x02 EXP
	// PUSH1 0x00 CALLDATALOAD DIV
	"60e060020a60003504",
}

func jumptab(fs *flag.FlagSet) {
	args := fs.Args()
	if len(args) != 1 {
		fatalf("usage: eth jumptab <address|->\n")
	}
	var code []byte
	if args[0] == "-" {
		buf, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fatalf("couldn't read stdin: %s\n", err)
		}
		if len(buf) > 1 && buf[len(buf)-1] == '\n' {
			buf = buf[:len(buf)-1]
		}
		code = make([]byte, hex.DecodedLen(len(buf)))
		_, err = hex.Decode(code, buf)
		if err != nil {
			fatalf("couldn't decode input: %s\n", err)
		}
	} else {
		var addr seth.Address
		err := addr.FromString(args[0])
		if err != nil {
			fatalf("jumptab: bad address: %s\n", err)
		}
		code = getcode(client(), &addr)
	}
	if len(code) == 0 {
		fatalf("address has no code\n")
	}

	entries := jumpentries(code)
	if len(entries) == 0 {
		return
	}

	dict := make(map[uint32]string)
	for _, sig := range preimage {
		h := seth.HashString(sig)
		dict[binary.LittleEndian.Uint32(h[:4])] = sig
	}
	for i := range entries {
		sigword := binary.LittleEndian.Uint32(entries[i].prefix[:])
		sig := dict[sigword]
		if sig == "" {
			fmt.Printf("%x pc:%5d\n", entries[i].prefix[:], entries[i].jmpdest)
		} else {
			fmt.Printf("%x pc:%5d %s\n", entries[i].prefix[:], entries[i].jmpdest, sig)
		}
	}
}

func jumpentries(code []byte) []jmpentry {
	// for each of the possible jump table preambles,
	// try to find an appropriate match in the code
	preamble := -1
	for _, p := range prefixes {
		buf, err := hex.DecodeString(p)
		if err != nil {
			panic(err)
		}
		preamble = bytes.Index(code, buf)
		if preamble != -1 {
			preamble = preamble + len(buf)
			break
		}
	}
	if preamble == -1 {
		fatalf("couldn't find a jump table preamble\n")
	}

	// supported jump table formats:
	//
	//   DUP1 PUSH4 0x06fdde03 EQ PUSH2 0x0145 JUMPI
	//
	//   PUSH4 0x06fdde03 DUP2 EQ PUSH2 0x0145 JUMPI
	//
	// TODO: is the PUSH after EQ always PUSH2?
	// That would make the code a bit simpler.
	var entries []jmpentry
	base := code[preamble:]
	for len(base) > 12 {
		var pushbytes, prefixbytes [4]byte

		if base[0] == oppush1+3 &&
			base[5] == opdup1+1 &&
			base[6] == opeq {
			// first case: PUSH4 <prefix> DUP2 EQ
			copy(prefixbytes[:], base[1:5])
		} else if base[0] == opdup1 &&
			base[1] == oppush1+3 &&
			base[6] == opeq {
			// second case: DUP1 PUSH4 <prefix> EQ
			copy(prefixbytes[:], base[2:6])
		} else {
			break
		}

		// width of PUSH used to identify PC
		pwidth := 1 + int(base[7]-oppush1)
		if pwidth > 4 {
			break // ???
		}
		copy(pushbytes[4-pwidth:], base[8:8+pwidth])
		if base[8+pwidth] != opjumpi {
			break // ???
		}
		entries = append(entries, jmpentry{
			prefix:  prefixbytes,
			jmpdest: int(binary.BigEndian.Uint32(pushbytes[:])),
		})
		base = base[8+pwidth+1:]
	}

	return entries
}
