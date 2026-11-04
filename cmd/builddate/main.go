// Command builddate prints the last commit date in ISO-8601 strict format.
package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	cmd := exec.Command("git", "-c", "safe.directory=*", "log", "-1",
		"--date=iso-strict", "--pretty=medium")
	out, err := cmd.Output()
	if err != nil {
		return // let the Makefile fall back to "unknown"
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "Date:") {
			fmt.Print(strings.TrimSpace(strings.TrimPrefix(line, "Date:")))
			return
		}
	}
}
