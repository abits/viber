package openspec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

var ErrNotInstalled = errors.New("openspec not installed; install with: npm install -g @fission-ai/openspec")

var execCommand = exec.CommandContext

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
