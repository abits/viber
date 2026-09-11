package tty

import (
	"os"
	"testing"
)

func TestIsInteractiveOnPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()
	defer func() { _ = w.Close() }()
	if IsInteractive(r) {
		t.Fatal("pipe reader should not be interactive")
	}
}

func TestIsInteractiveNil(t *testing.T) {
	if IsInteractive(nil) {
		t.Fatal("nil file should not be interactive")
	}
}
