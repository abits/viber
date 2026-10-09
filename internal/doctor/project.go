package doctor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// DetectProject resolves the project root for cwd and reports whether cwd
// lives in a git work tree and whether the root looks like a viber scaffold.
//
// root is the git top-level directory when inGit is true, otherwise cwd.
// Any git failure (git missing, cwd outside a work tree, timeout) falls back
// to cwd with inGit = false; it is never an error that the caller must handle.
//
// ok reports whether root contains the three sentinels viber scaffolds ship:
// CLAUDE.md, openspec/, and .claude/. Project checks run only when ok is true.
// Detection itself produces no Result and no output line.
func DetectProject(ctx context.Context, cwd string) (root string, inGit bool, ok bool) {
	root = cwd

	pctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	cmd := execCommand(pctx, "git", "-C", cwd, "rev-parse", "--show-toplevel")
	cmd.WaitDelay = probeWaitDelay
	out, err := cmd.Output()
	if err == nil {
		if top := strings.TrimSpace(string(out)); top != "" {
			root = top
			inGit = true
		}
	}

	for _, name := range []string{"CLAUDE.md", "openspec", ".claude"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			return root, inGit, false
		}
	}
	return root, inGit, true
}
