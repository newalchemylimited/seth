package seth

import (
	"encoding/json"
	"errors"
)

var PasswordDenied = errors.New("password denied")

// UnlockAccount unlocks the given address with
// the associated passphrase.
func (c *Client) UnlockAccount(addr *Address, password string) error {
	p0, _ := json.Marshal(addr)
	p1, _ := json.Marshal(password)
	var ret bool
	err := c.do("personal_unlockAccount", []json.RawMessage{p0, p1, json.RawMessage(`"0x0"`)}, &ret)
	if err != nil {
		return err
	}
	if !ret {
		return PasswordDenied
	}
	return nil
}
