package main

import (
	"fmt"
	"os"

	"rewind/internal/cli"
)

// Version can be set at build time with -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	if err := cli.Execute(Version); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
