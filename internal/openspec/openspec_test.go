package openspec

import (
	"context"
	"errors"
	"os/exec"
	"testing"
)

func TestVerifyMissingBinary(t *testing.T) {
	orig := execCommand
	defer func() { execCommand = orig }()
	execCommand = func(ctx context.Context, _ string, args ...string) *exec.Cmd {
		// No slash → goes through PATH lookup, returns *exec.Error like real "openspec" missing.
		return exec.CommandContext(ctx, "nonexistent-viber-openspec-binary-xyz", args...)
	}
	err := Verify(context.Background())
	if !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("want ErrNotInstalled, got %v", err)
	}
}

func TestVerifyOK(t *testing.T) {
	orig := execCommand
	defer func() { execCommand = orig }()
	execCommand = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "true")
	}
	if err := Verify(context.Background()); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
