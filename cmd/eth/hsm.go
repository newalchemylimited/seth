package main

import (
	"os"
	"strings"
	"sync"

	"github.com/newalchemylimited/seth"
)

var hsms []seth.HSM
var hsmonce sync.Once

func hsmprobe() []seth.HSM {
	hsmonce.Do(func() {
		for _, p := range []string{
			"yubihsm",
		} {
			var hints []string
			if hstr := os.Getenv(strings.ToUpper(p) + "_HINT"); hstr != "" {
				hints = strings.Split(hstr, ",")
			}
			h := seth.FindHSM(p, hints...)
			if h == nil {
				continue
			}
			var pass []byte
			if pstr := os.Getenv(strings.ToUpper(p) + "_PASS"); pstr != "" {
				pass = []byte(pstr)
			} else {
				pass = passpromptf("enter HSM password for %s:\n", p)
			}
			if err := h.Unlock(pass); err != nil {
				fatalf("unlock %s: %s", p, err)
			}
			debugf("probe for hsm %q succeeded", p)
			hsms = append(hsms, h)
		}
	})
	return hsms
}
