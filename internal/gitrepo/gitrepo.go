package gitrepo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Init(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "git", "init", "--quiet", dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init %s: %w: %s", dir, err, out)
	}
	return nil
}

func AddRemote(ctx context.Context, dir, name, url string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "remote", "add", name, url)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git remote add %s %s: %w: %s", name, url, err, out)
	}
	return nil
}

func IsRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		return false
	}
	return info.IsDir() || info.Mode().IsRegular()
}
