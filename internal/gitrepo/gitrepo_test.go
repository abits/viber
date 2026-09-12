package gitrepo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
}

func TestIsRepo(t *testing.T) {
	t.Run("plain directory is not a repo", func(t *testing.T) {
		if IsRepo(t.TempDir()) {
			t.Error("empty dir reported as a repo")
		}
	})

	t.Run("missing directory is not a repo", func(t *testing.T) {
		if IsRepo(filepath.Join(t.TempDir(), "nope")) {
			t.Error("missing dir reported as a repo")
		}
	})

	t.Run("worktree with a .git file is a repo", func(t *testing.T) {
		dir := t.TempDir()
		// Linked worktrees and submodules carry a .git *file*, not a dir.
		if err := os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: ../x\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if !IsRepo(dir) {
			t.Error("worktree with a .git file not reported as a repo")
		}
	})
}

func TestInitAndAddRemote(t *testing.T) {
	requireGit(t)
	ctx := context.Background()
	dir := t.TempDir()

	if IsRepo(dir) {
		t.Fatal("fresh temp dir already looks like a repo")
	}
	if err := Init(ctx, dir); err != nil {
		t.Fatal(err)
	}
	if !IsRepo(dir) {
		t.Error("directory is not a repo after Init")
	}

	if err := AddRemote(ctx, dir, "origin", "https://example.invalid/x.git"); err != nil {
		t.Fatal(err)
	}
	out, err := exec.CommandContext(ctx, "git", "-C", dir, "remote", "get-url", "origin").Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := string(out); got != "https://example.invalid/x.git\n" {
		t.Errorf("origin = %q", got)
	}

	t.Run("adding origin twice is an error", func(t *testing.T) {
		if err := AddRemote(ctx, dir, "origin", "https://example.invalid/y.git"); err == nil {
			t.Error("want an error when the remote already exists")
		}
	})
}

func TestInitRejectsCancelledContext(t *testing.T) {
	requireGit(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Init(ctx, t.TempDir()); err == nil {
		t.Error("want an error from a cancelled context")
	}
}
