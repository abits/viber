package templates

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"testing/fstest"
)

var httpClient = http.DefaultClient

// maxUncompressedBytes is the ceiling for a --from tarball after gzip.
// A few MiB is generous for a template set; this also keeps gosec G110 quiet.
const defaultMaxUncompressedBytes int64 = 8 << 20

var maxUncompressedBytes int64 = defaultMaxUncompressedBytes

func Fetch(ctx context.Context, owner, repo, ref string) (fs.FS, error) {
	url := fmt.Sprintf("https://codeload.github.com/%s/%s/tar.gz/%s", owner, repo, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: %s", url, resp.Status)
	}
	return unpackTarGz(resp.Body)
}

func unpackTarGz(r io.Reader) (fs.FS, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	m := fstest.MapFS{}
	var total int64
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		parts := strings.SplitN(hdr.Name, "/", 2)
		if len(parts) < 2 || parts[1] == "" {
			continue
		}
		name := parts[1]
		if strings.HasPrefix(hdr.Name, "/") || !fs.ValidPath(name) {
			return nil, fmt.Errorf("template set contains an unsafe path %q", hdr.Name)
		}
		remain := maxUncompressedBytes - total
		if remain < 0 {
			remain = 0
		}
		buf, err := io.ReadAll(io.LimitReader(tr, remain+1))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", hdr.Name, err)
		}
		total += int64(len(buf))
		if total > maxUncompressedBytes {
			return nil, fmt.Errorf("template set exceeds %d byte uncompressed limit", maxUncompressedBytes)
		}
		// raw tar mode is truncated to permission bits (0o777) so any overflow is intentional.
		m[name] = &fstest.MapFile{Data: buf, Mode: fs.FileMode(hdr.Mode) & 0o777}
	}
	return m, nil
}
