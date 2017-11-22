package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func do(t *testing.T, cmdline string) {
	t.Helper()
	f := strings.Fields(cmdline)
	cmd := exec.Command(f[0], f[1:]...)
	t.Log(cmdline)
	err := cmd.Run()
	if err != nil {
		t.Fatal(cmdline+":", err)
	}
}

func TestBindgen(t *testing.T) {
	gofiles, err := filepath.Glob("./test/*.go")
	if err != nil {
		t.Fatal(err)
	}
	do(t, "go generate ./test/")
	do(t, "go run "+strings.Join(gofiles, " "))
}
