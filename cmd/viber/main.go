// Command viber scaffolds a vibe-coding project wired for Claude Code + OpenSpec.
package main

import (
	"os"

	"github.com/abits/viber/internal/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(cmd.Execute(version, commit, date))
}
