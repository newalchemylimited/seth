package seth

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
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

// IPCPath returns a closure that dials a unix socket.
//
// It can be used in NewClient like
//
//  NewClient(IPCPath("/path/to/geth.ipc"))
//
func IPCPath(s string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return net.Dial("unix", s)
	}
}

// IPCDial dials geth over local IPC
//
// It can be used in NewClient like
//
//  NewClient(IPCDial)
//
func IPCDial() (io.ReadWriteCloser, error) {
	var sockpath string
	switch runtime.GOOS {
	case "darwin":
		sockpath = filepath.Join(home(), "Library", "Ethereum")
	case "linux":
		sockpath = filepath.Join(home(), ".ethereum")
	default:
		return nil, fmt.Errorf("unsupported GOOS %q", runtime.GOOS)
	}
	sockpath = filepath.Join(sockpath, "geth.ipc")
	return net.Dial("unix", sockpath)
}
