package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunCommands(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer

	if code := run([]string{"mkdir", filepath.Join(dir, "a", "b")}, &out, &errOut); code != 0 {
		t.Fatalf("mkdir code = %d, err = %q", code, errOut.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "a", "b")); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"rm", file}, &out, &errOut); code != 0 {
		t.Fatalf("rm code = %d", code)
	}
	if code := run([]string{"rm", file}, &out, &errOut); code != 0 {
		t.Fatalf("rm missing code = %d", code)
	}
	globDir := filepath.Join(dir, "glob")
	if err := os.MkdirAll(globDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a", "b"} {
		if err := os.WriteFile(filepath.Join(globDir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if code := run([]string{"rmglob", filepath.Join(globDir, "*")}, &out, &errOut); code != 0 {
		t.Fatalf("rmglob code = %d, err = %q", code, errOut.String())
	}
	if code := run([]string{"rmdir", globDir}, &out, &errOut); code != 0 {
		t.Fatalf("rmdir code = %d", code)
	}
}

func TestRunErrors(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run(nil, &out, &errOut); code != 2 {
		t.Fatalf("empty code = %d", code)
	}
	if code := run([]string{"unknown", "x"}, &out, &errOut); code != 2 {
		t.Fatalf("unknown code = %d", code)
	}
	if code := run([]string{"tag"}, &out, &errOut); code != 2 {
		t.Fatalf("tag code = %d", code)
	}
	if code := run([]string{"rmglob", "["}, &out, &errOut); code != 1 {
		t.Fatalf("bad glob code = %d", code)
	}
}
func TestRunFilesystemErrors(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := run([]string{"mkdir", file}, &out, &errOut); code != 1 {
		t.Fatalf("mkdir file code = %d", code)
	}
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "child"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"rm", sub}, &out, &errOut); code != 1 {
		t.Fatalf("rm non-empty code = %d", code)
	}
}
