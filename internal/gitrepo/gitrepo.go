// Package gitrepo wraps the git subprocesses used while scaffolding a project.
package gitrepo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Init runs `git init --quiet` in dir.
func Init(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "git", "init", "--quiet", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init %s: %w: %s", dir, err, out)
	}
	return nil
}

// AddRemote runs `git remote add <name> <url>` in dir.
func AddRemote(ctx context.Context, dir, name, url string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "remote", "add", name, url)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git remote add %s %s: %w: %s", name, url, err, out)
	}
	return nil
}

// IsRepo reports whether dir already contains a .git directory or file.
func IsRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		return false
	}
	return info.IsDir() || info.Mode().IsRegular()
}
