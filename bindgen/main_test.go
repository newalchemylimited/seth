package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func do(t *testing.T, cmdline string) {
	f := strings.Fields(cmdline)
	buf, err := exec.Command(f[0], f[1:]...).CombinedOutput()
	if err != nil {
		t.Logf("%s: %s\n", cmdline, err)
		t.Fatalf("command output:\n%s\n", string(buf))
	}
}

func TestBindgen(t *testing.T) {
	gofiles, err := filepath.Glob("./test/*.go")
	if err != nil {
		t.Fatal(err)
	}
	do(t, "go generate .")
	do(t, "go generate -v ./test/")
	do(t, "go run "+strings.Join(gofiles, " "))
}
