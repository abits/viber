// Package tty reports whether a file descriptor is an interactive terminal.
package tty

import (
	"os"

	"golang.org/x/term"
)

// IsInteractive reports whether f is a terminal. A nil file is not interactive.
func IsInteractive(f *os.File) bool {
	if f == nil {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
