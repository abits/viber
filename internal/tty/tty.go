package tty

import (
	"os"

	"golang.org/x/term"
)

func IsInteractive(f *os.File) bool {
	if f == nil {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
