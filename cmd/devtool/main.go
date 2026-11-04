// Command devtool provides small cross-platform filesystem helpers for the
// Makefile, so build recipes do not depend on cmd.exe/PowerShell or sh.
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, "usage: devtool mkdir|rmdir|rm|rmglob <path...>")
		return 2
	}
	switch args[0] {
	case "mkdir":
		for _, p := range args[1:] {
			if err := os.MkdirAll(p, 0o755); err != nil {
				fmt.Fprintln(stderr, "devtool:", err)
				return 1
			}
		}
	case "rmdir":
		for _, p := range args[1:] {
			if err := os.RemoveAll(p); err != nil {
				fmt.Fprintln(stderr, "devtool:", err)
				return 1
			}
		}
	case "rm":
		for _, p := range args[1:] {
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				fmt.Fprintln(stderr, "devtool:", err)
				return 1
			}
		}
	case "tag":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: devtool tag <name>")
			return 2
		}
		tag := args[1]
		clean, err := gitOutput("status", "--porcelain")
		if err != nil {
			fmt.Fprintln(stderr, "devtool:", err)
			return 1
		}
		if strings.TrimSpace(clean) != "" {
			fmt.Fprintln(stderr, "devtool: working tree is dirty, commit the release first")
			return 1
		}
		cmd := exec.Command("git", "-c", "safe.directory=*", "tag", "-a", tag,
			"-m", "Release "+tag)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(stderr, "devtool:", err)
			return 1
		}
		fmt.Fprintln(stdout, "Created tag", tag)
	case "rmglob":
		for _, pattern := range args[1:] {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				fmt.Fprintln(stderr, "devtool:", err)
				return 1
			}
			for _, p := range matches {
				if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
					fmt.Fprintln(stderr, "devtool:", err)
					return 1
				}
			}
		}
	default:
		fmt.Fprintf(stderr, "devtool: unknown command %q\n", args[0])
		return 2
	}
	return 0
}

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-c", "safe.directory=*"}, args...)...)
	out, err := cmd.Output()
	return string(out), err
}
