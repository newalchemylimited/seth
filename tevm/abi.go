package tevm

import (
	"strings"
)

// Find finds a function or event based on the hash
// of the descriptor. Keep in mind that the descriptor
// must use canonical type names and no superfluous whitespace.
//
// For example:
//
//   c.Find("transfer(address,uint256)")
//
// find the ERC20 "transfer" function. Similarly,
//
//   c.Find("Transfer(address,address,uint256)")
//
// finds the ERC20 "Transfer" event.
func (c *CompiledContract) Find(desc string) *ABIDescriptor {
	for i := range c.ABI {
		d := &c.ABI[i]
		if d.Signature() == desc {
			return d
		}
	}
	return nil
}

// Signature returns the canonical function/event signature.
// For functions, the first 4 bytes of the hash of the
// signature is the function selector, and for events,
// the hash of the signature is the first topic in the log.
func (d *ABIDescriptor) Signature() string {
	var args []string
	for i := range d.Inputs {
		args = append(args, d.Inputs[i].Type)
	}
	return d.Name + "(" + strings.Join(args, ",") + ")"
}
