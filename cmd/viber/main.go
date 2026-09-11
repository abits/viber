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
