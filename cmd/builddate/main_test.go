package main

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"
)

func TestRun(t *testing.T) {
	var out bytes.Buffer
	run(&out, func() ([]byte, error) {
		return []byte("commit abc\nDate:   2026-11-04T10:00:00+03:00\n"), nil
	})
	if got := out.String(); got != "2026-11-04T10:00:00+03:00" {
		t.Fatalf("out = %q", got)
	}
}

func TestRunErrors(t *testing.T) {
	var out bytes.Buffer
	run(&out, func() ([]byte, error) { return nil, errors.New("git") })
	run(&out, func() ([]byte, error) { return []byte("no date"), nil })
	if out.Len() != 0 {
		t.Fatalf("out = %q", out.String())
	}
}
func TestGitLog(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	data, err := gitLog()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("Date:")) {
		t.Fatalf("git log = %q", data)
	}
}
