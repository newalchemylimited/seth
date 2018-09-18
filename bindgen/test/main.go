package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/newalchemylimited/seth"
)

//go:generate bindgen -b -c=Test -o generated.go compiletest.sol

func fatal(j ...interface{}) {
	fmt.Fprintln(os.Stderr, j...)
	os.Exit(1)
}

func check(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fatal(file, line, err)
	}
}

func assert(cond bool) {
	if !cond {
		_, file, line, _ := runtime.Caller(1)
		fatal(file, line, "assertion failed")
	}
}

func main() {

	// note: this program gets run from the parent directory
	bundle, err := seth.CompileGlob("test/*.sol")
	if err != nil {
		bundle, err = seth.CompileGlob("*.sol")
		if err != nil {
			fatal(err)
		}
	}

	// c := tevm.NewChain()
	// acct := c.NewAccount(1)

	url := "http://localhost:7545"
	//url := "http://localhost:8545"

	c := seth.NewHTTPClient(url)
	fundingAddress, err := seth.ParseAddress("0x84ede7C61cBFf3056D6dEb24FF774b79c1d2c4E4") // ganache
	//fundingAddress, err := seth.ParseAddress("0x5231a93db3ce6cbb709af94a267dd0e747d30f82") // parity

	sender := seth.NewSender(c, fundingAddress)
	contract := bundle.Contract("Test")
	ccode := contract.Code

	if !bytes.Equal(seth.StripBytecode(ccode), seth.StripBytecode(TestCode)) {
		//fatal("compiled and precompiled code not identical")
	}

	//addr, _ := seth.ParseAddress("0xc42286d90be0bc5ebe8c141de13d0451e62ca897")

	//*
	cc, addr, err := DeployTest(sender, nil, uint16(123), "hi how are you")
	//addr, err := sender.Create(TestCode, nil, "(uint16,string)", uint16(123), "hi how are you")
	//addr, err := c.Create(&acct, TestCode)
	if err != nil {
		fatal("deploying the contract:", err)
	} //*/

	log.Printf("Installed contract to: %s", addr.String())

	//sender := c.Sender(&acct)
	//sender.Pending = true

	//cc := NewTest(&addr, sender)

	//s := "hello"
	// b := []byte(s)

	// var b32 [32]byte
	// copy(b32[:], b[:])

	// spew.Dump(cc.SetBytes32Val(b32))

	// spew.Dump(cc.Bytes32Val())

	// spew.Dump(cc.SetBytesVal(b))

	// spew.Dump(cc.BytesVal())

	//spew.Dump(cc.SetStringVal(s))
	//spew.Dump(cc.SetStringVal("elliot"))

	//spew.Dump(cc.StringVal())
	//time.Sleep(time.Second * 5)

	spew.Dump("constructor")
	spew.Dump(cc.Cstring())
	spew.Dump(cc.Cuint16())

	spew.Dump(cc.SendTestEvent(123, "test", []byte("hihi")))
	spew.Dump(cc.SendTestEvent(321, "something else", []byte("goodbye")))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)

	it, err := cc.FilterSomethingHappened(ctx, 0, -1)
	if err != nil {
		panic(err)
	}

	for it.Next() {
		spew.Dump("Event:", it.Event)
	}

	if it.Error == context.DeadlineExceeded {
		log.Printf("context timed out")
	} else if it.Error != nil {
		panic(it.Error)
	}

	cancel()

	//spew.Dump(cc.AddElliot())

	//time.Sleep(time.Second * 10)

	//spew.Dump(cc.AllPeople())

	// log.Printf("installed contract to: %s", addr.String())

	// client := seth.NewHTTPClient("http://localhost:7545")
	// fundingAddress, err := seth.ParseAddress("0x84ede7C61cBFf3056D6dEb24FF774b79c1d2c4E4")
	// if err != nil {
	// 	fatal("bad funding address")
	// }

	// sender := seth.NewSender(client, fundingAddress)

	// // contract := bundle.Contract("Test")
	// // ccode := contract.Code

	// // if !bytes.Equal(seth.StripBytecode(ccode), seth.StripBytecode(TestCode)) {
	// // 	fatal("compiled and precompiled code not identical")
	// // }

	// // addr, err := sender.Create(TestCode, nil)
	// // //addr, err := c.Create(&acct, TestCode)
	// // if err != nil {
	// // 	fatal("deploying the contract:", err)
	// // }

	// // spew.Dump("Deployed contract to", addr.String())

	// addr, _ := seth.ParseAddress("0xcd655ab80b149302831855cf91b7794705f1e564")

	// //cc := NewTest(&addr, c.Sender(&acct))
	// cc := NewTest(addr, sender)

	//spew.Dump(cc.Name())

	/*
		BAD:
		0x4df9dcd362656e0000000000000000000000000000000000000000000000000000000000

		GOOD:
		0x4df9dcd3000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000026869000000000000000000000000000000000000000000000000000000000000
	*/

	// initial, err := cc.Value()
	// check(err)

	// initial2, err := cc.Counter()
	// check(err)
	// assert(initial == initial2)

	// _, err = cc.MustThrow()
	// assert(err != nil)

	// _, err = cc.Inc()
	// check(err)

	// afterIncrement, err := cc.Value()
	// check(err)
	// assert(afterIncrement == initial+1)

	// afterIncrement2, err := cc.Counter()
	// check(err)
	// assert(afterIncrement2 == afterIncrement)

	// v2, err := cc.DoubleThis(100)
	// check(err)
	// assert(v2 == 200)

	// _, err = cc.SetName("ben")
	// check(err)
	// name, err := cc.Name()
	// check(err)
	// assert(name == "ben")

	// defaultRoundTripper := http.DefaultTransport
	// defaultTransportPointer, ok := defaultRoundTripper.(*http.Transport)
	// if !ok {
	// 	panic(fmt.Sprintf("defaultRoundTripper not an *http.Transport"))
	// }

	// defaultTransportPointer.MaxIdleConns = 1000
	// defaultTransportPointer.MaxIdleConnsPerHost = 1000

	// var wg sync.WaitGroup

	// for i := 0; i < 100; i++ {
	// 	log.Printf("adding %d\n", i)
	// 	wg.Add(100)
	// 	go func(i int) {
	// 		log.Printf("starting %d\n", i)

	// 		// client := seth.NewHTTPClient(url)

	// 		// p := &seth.ParityClient{*client}
	// 		// peers, _ := p.NetPeers()

	// 		// j, _ := json.MarshalIndent(peers, " ", "  ")
	// 		// log.Println(string(j))

	// 		// spew.Dump(p.Chain())
	// 		// spew.Dump(p.Mode())
	// 		// fundingAddress, err := seth.ParseAddress("0x5231a93db3ce6cbb709af94a267dd0e747d30f82")
	// 		// if err != nil {
	// 		// 	fatal("bad funding address")
	// 		// }

	// 		// sender := seth.NewSender(client, fundingAddress)

	// 		//addr, _ := seth.ParseAddress("0xcd655ab80b149302831855cf91b7794705f1e564")

	// 		// cc := NewTest(addr, sender)

	// 		for j := 0; j < 100; j++ {

	// 			_, err = cc.Inc()
	// 			if err != nil {
	// 				panic(err)
	// 			}

	// 			if j%33 == 0 {
	// 				spew.Dump(cc.Counter())
	// 			}
	// 			wg.Done()
	// 		}
	// 		log.Printf("ending %d\n", i)

	// 	}(i)
	// }

	// wg.Wait()

}

// hexstring returns a hex string of the given data
func hexstring(b []byte, trunc bool) []byte {
	buf := make([]byte, 2+2*len(b))
	hex.Encode(buf[2:], b)
	if trunc && len(buf) > 2 && buf[2] == '0' {
		buf = buf[1:]
	}
	copy(buf, "0x")
	return buf
}
