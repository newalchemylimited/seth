package erc20

import (
	"testing"

	"github.com/newalchemylimited/seth"
	"github.com/newalchemylimited/seth/tevm"
)

func TestEthereumToken(t *testing.T) {
	bundle, err := tevm.CompileGlob("*.sol")
	if err != nil {
		t.Fatal(err)
	}
	code := bundle.Contract("TokenERC20")
	if code == nil {
		t.Fatal("no Token contract")
	}

	c := tevm.NewChain()
	owner := c.NewAccount(1)

	addr, err := c.Create(&owner, code.Code)
	if err != nil {
		t.Fatal(err)
	}

	TestCompliance(&Token{
		T:       t,
		C:       c,
		Addr:    &addr,
		Owner:   &owner,
		HasBurn: true,
		Mint: func(tok *Token, to *seth.Address, amt *seth.Int) error {
			_, err := c.Call(tok.Owner, tok.Addr, "mint(address,uint256)", to, amt)
			return err
		},
	})
}
