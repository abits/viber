package templates

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"testing/fstest"

	"github.com/abits/viber/internal/ghfetch"
)

var client ghfetch.Client

const maxTemplateSetSize int64 = 8 << 20

// Fetch downloads a GitHub repository tarball and returns it as an fs.FS
// with the top-level directory stripped.
func Fetch(ctx context.Context, owner, repo, ref string) (fs.FS, error) {
	url := fmt.Sprintf("https://codeload.github.com/%s/%s/tar.gz/%s", owner, repo, ref)
	resp, err := client.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: %s", url, resp.Status)
	}
	return unpackTarGz(resp.Body)
}

func unpackTarGz(r io.Reader) (fs.FS, error) {
	fsys, err := ghfetch.UntarGz(r, ghfetch.Options{
		MaxUncompressedBytes: maxTemplateSetSize,
		RejectUnsafePaths:    true,
	})
	if err != nil {
		return nil, err
	}
	return stripTopDir(fsys)
}

// stripTopDir drops each entry's first path element — the "repo-ref/"
// directory GitHub's tarballs always wrap everything in — and drops any
// entry that has no such element to strip (matching the prior behavior of
// simply skipping tar entries without a leading directory component).
func stripTopDir(fsys fs.FS) (fs.FS, error) {
	out := fstest.MapFS{}
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		parts := strings.SplitN(p, "/", 2)
		if len(parts) < 2 || parts[1] == "" {
			return nil
		}
		b, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		out[parts[1]] = &fstest.MapFile{Data: b, Mode: info.Mode()}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
