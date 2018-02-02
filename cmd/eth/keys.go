package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/newalchemylimited/seth"
)

func home() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	if u, err := user.Current(); err == nil {
		return u.HomeDir
	}
	return ""
}

// keypaths returns the list of directories
// in which to look for ethereum key data
//
// the default paths may be overridden by
// specifiying a :-separated KEY_PATH environment variable
func keypaths() []string {
	if kp := os.Getenv("KEY_PATH"); kp != "" {
		return strings.Split(kp, ":")
	}
	h := home()
	switch runtime.GOOS {
	case "darwin":
		return []string{
			filepath.Join(h, "/Library/Application Support/io.parity.ethereum/keys/ethereum/"), // parity
			filepath.Join(h, "/Library/Ethereum/keystore/"),                                    // geth
		}
	case "linux":
		return []string{
			filepath.Join(h, ".parity/keys/ethereum"), // parity
			filepath.Join(h, ".ethereum/keystore"),    // geth
		}
	}
	return nil
}

type keydesc struct {
	path string        // the "path" of the key (which may or may not be a filesystem path)
	addr seth.Address  // the actual ethereum address
	kf   *seth.Keyfile // the keyfile, if there is one associated
}

func keys() []keydesc {
	var out []keydesc
	for _, p := range keypaths() {
		fis, err := ioutil.ReadDir(p)
		if err != nil {
			fatalf("getting key files: %s", err)
		}
		for _, fi := range fis {
			if fi.IsDir() {
				continue
			}
			fp := filepath.Join(p, fi.Name())
			kf := new(seth.Keyfile)
			buf, err := ioutil.ReadFile(fp)
			if err != nil {
				debugf("keyfile scan: can't read file %q %s", fp, err)
				continue
			}
			if len(buf) == 0 || buf[0] != '{' {
				debugf("keyfile scan: file %q doesn't look like json", fp)
				continue
			}
			if err := json.Unmarshal(buf, kf); err == nil &&
				kf.Address != "" &&
				kf.ID != "" &&
				strings.HasSuffix(fp, kf.ID) {
				debugf("using keyfile %q", fp)
				kd := keydesc{
					path: fp,
					kf:   kf,
				}
				kd.addr.FromString("0x" + kf.Address)
				out = append(out, kd)
			}
		}
	}
	return out
}
