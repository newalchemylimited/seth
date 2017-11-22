package main

import (
	"github.com/newalchemylimited/seth"
)

type Test struct {
	addr *seth.Address
	s    *seth.Sender
}

func NewTest(addr *seth.Address, sender *seth.Sender) *Test {
	return &Test{addr: addr, s: sender}
}

func (z *Test) MustThrow() (seth.Hash, error) {
	return z.s.Send(z.addr, "mustThrow()")
}
func (z *Test) Inc() (seth.Hash, error) {
	return z.s.Send(z.addr, "inc()")
}
func (z *Test) Value() (ret0 seth.Int, err error) {
	d := seth.NewABIDecoder(&ret0)
	err = z.s.ConstCall(z.addr, "value()", d)
	return
}
func (z *Test) Counter() (ret0 seth.Int, err error) {
	d := seth.NewABIDecoder(&ret0)
	err = z.s.ConstCall(z.addr, "counter()", d)
	return
}
