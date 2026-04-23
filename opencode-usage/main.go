package main

import (
	"os"

	"github.com/sigco3111/opencode-usage-cli/opencode-usage/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
