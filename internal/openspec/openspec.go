// Package openspec invokes the OpenSpec CLI used by `viber init`.
package openspec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// ErrNotInstalled is returned by Verify when the openspec binary is not on PATH.
var ErrNotInstalled = errors.New("openspec not installed; install with: npm install -g @fission-ai/openspec")

var execCommand = exec.CommandContext

// Verify runs `openspec --version`. It returns ErrNotInstalled when the
// binary is missing from PATH.
func Verify(ctx context.Context) error {
	cmd := execCommand(ctx, "openspec", "--version")
	if err := cmd.Run(); err != nil {
		var pathErr *exec.Error
		if errors.As(err, &pathErr) {
			return ErrNotInstalled
		}
		return fmt.Errorf("openspec --version: %w", err)
	}
	return nil
}

// Init runs `openspec init` in dir, inheriting stdin, stdout, and stderr.
func Init(ctx context.Context, dir string) error {
	cmd := execCommand(ctx, "openspec", "init")
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("openspec init: %w", err)
	}
	return nil
}
