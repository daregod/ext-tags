// Command devtool provides small cross-platform filesystem helpers for the
// Makefile, so build recipes do not depend on cmd.exe/PowerShell or sh.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	case "tag":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "usage: devtool tag <name>")
			os.Exit(2)
		}
		tag := os.Args[2]
		clean, err := gitOutput("status", "--porcelain")
		if err != nil {
			fail(err)
		}
		if strings.TrimSpace(clean) != "" {
			fmt.Fprintln(os.Stderr, "devtool: working tree is dirty, commit the release first")
			os.Exit(1)
		}
		cmd := exec.Command("git", "-c", "safe.directory=*", "tag", "-a", tag,
			"-m", "Release "+tag)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fail(err)
		}
		fmt.Println("Created tag", tag)
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

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-c", "safe.directory=*"}, args...)...)
	out, err := cmd.Output()
	return string(out), err
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "devtool:", err)
	os.Exit(1)
}
