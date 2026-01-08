package main

import (
	"os"

	"github.com/jaknapp/applink/internal/cli"
)

// Set by goreleaser via ldflags
var (
	version = "dev"
	commit  = "none"
)

func main() {
	cli.SetVersion(version, commit)
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
