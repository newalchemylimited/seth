package tevm

import (
	"bytes"
	"testing"
	"time"

	"github.com/newalchemylimited/seth"
)

func TestFilter(t *testing.T) {
	bundle, err := seth.CompileGlob("./erc20/*.sol")
	if err != nil {
		t.Fatal(err)
	}

	chain := NewChain()
	acct := chain.NewAccount(1)

	client := seth.NewClientTransport(chain)
	sender := seth.NewSender(client, &acct)

	token, err := sender.Create(bundle.Contract("TokenERC20").Code)
	if err != nil {
		t.Fatal(err)
	}

	// This filter should capture all ERC20 Transfer() events
	filter, err := client.FilterTopics([]*seth.Hash{&seth.ERC20Transfer}, &token, -1, -1)
	if err != nil {
		t.Fatal(err)
	}

	checkTransfer := func() {
		if err := filter.Err(); err != nil {
			t.Fatal(err)
		}
		select {
		case l := <-filter.Out():
			if !bytes.Equal(l.Topics[0], seth.ERC20Transfer[:]) {
				t.Fatal("not an ERC20 Transfer...")
			}
		case <-time.Tick(10 * time.Second):
			t.Helper()
			t.Fatal("no ERC20 event")
		}
	}

	h, err := sender.Send(&token, "mint(address,uint256)", &acct, seth.NewInt(10000))
	if err != nil {
		t.Fatal(err)
	}
	if err := sender.Wait(&h); err != nil {
		t.Fatal(err)
	}

	checkTransfer()

	acct2 := chain.NewAccount(0)
	h, err = sender.Send(&token, "transfer(address,uint256)", &acct2, seth.NewInt(9000))
	if err != nil {
		t.Fatal(err)
	}
	if err := sender.Wait(&h); err != nil {
		t.Fatal(err)
	}

	checkTransfer()
	filter.Close()
}
