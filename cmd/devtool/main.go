// Command devtool provides small cross-platform filesystem helpers for the
// Makefile, so build recipes do not depend on cmd.exe/PowerShell or sh.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: devtool mkdir|rmdir|rm|rmglob <path...>")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "mkdir":
		for _, p := range os.Args[2:] {
			if err := os.MkdirAll(p, 0o755); err != nil {
				fail(err)
			}
		}
	case "rmdir":
		for _, p := range os.Args[2:] {
			if err := os.RemoveAll(p); err != nil {
				fail(err)
			}
		}
	case "rm":
		for _, p := range os.Args[2:] {
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				fail(err)
			}
		}
	case "rmglob":
		for _, pattern := range os.Args[2:] {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				fail(err)
			}
			for _, p := range matches {
				if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
					fail(err)
				}
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "devtool: unknown command %q\n", os.Args[1])
		os.Exit(2)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "devtool:", err)
	os.Exit(1)
}
