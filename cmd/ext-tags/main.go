package main

import (
	"os"

	"external-tags/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
