package main

import (
	"os"

	"github.com/lavr/rctl/internal/cli"
)

// version is set at build time via ldflags.
var version = "dev"

func main() {
	os.Exit(cli.Execute(version))
}
