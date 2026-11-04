// Command builddate prints the last commit date in ISO-8601 strict format.
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	run(os.Stdout, gitLog)
}

func run(out io.Writer, log func() ([]byte, error)) {
	data, err := log()
	if err != nil {
		return // let the Makefile fall back to "unknown"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "Date:") {
			fmt.Fprint(out, strings.TrimSpace(strings.TrimPrefix(line, "Date:")))
			return
		}
	}
}

func gitLog() ([]byte, error) {
	cmd := exec.Command("git", "-c", "safe.directory=*", "log", "-1",
		"--date=iso-strict", "--pretty=medium")
	return cmd.Output()
}
