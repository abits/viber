// Package updater downloads GitHub release assets and installs them atomically.
package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/abits/viber/internal/ghfetch"
)

// releaseAPI is the GitHub API endpoint for a repository's latest release.
// A var, not a const, so tests can point it at an httptest.Server.
var releaseAPI = "https://api.github.com/repos/%s/%s/releases/latest"

// client issues the HTTP requests Latest and Install make. A package var so
// tests can inject an httptest.Server's client.
var client ghfetch.Client

// Release is a GitHub release as returned by the latest-release API.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset is a downloadable file attached to a Release.
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// Latest fetches the latest GitHub release for owner/repo.
func Latest(ctx context.Context, owner, repo string) (*Release, error) {
	url := fmt.Sprintf(releaseAPI, owner, repo)
	req, err := client.NewRequest(ctx, url)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	var r Release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Install downloads the release asset matching the current GOOS/GOARCH,
// extracts `binary` from it, and replaces `dest` atomically (0755). It
// returns the installed tag name. Windows archives are .zip and are not
// supported.
func Install(ctx context.Context, rel *Release, binary, dest string) (string, error) {
	if runtime.GOOS == "windows" {
		return "", errors.New("updater.Install: windows archives (.zip) are not supported")
	}
	ver := strings.TrimPrefix(rel.TagName, "v")
	wantPrefix := fmt.Sprintf("%s_%s_%s_%s", binary, ver, runtime.GOOS, runtime.GOARCH)
	var chosen *Asset
	for i := range rel.Assets {
		if strings.HasPrefix(rel.Assets[i].Name, wantPrefix) {
			chosen = &rel.Assets[i]
			break
		}
	}
	if chosen == nil {
		return "", fmt.Errorf("no release asset matches %s in %s", wantPrefix, rel.TagName)
	}

	resp, err := client.Get(ctx, chosen.URL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", chosen.URL, resp.Status)
	}

	body, err := ExtractBinary(resp.Body, binary)
	if err != nil {
		return "", err
	}
	if err := writeAtomic(dest, body, 0o755); err != nil {
		return "", err
	}
	return rel.TagName, nil
}

// ExtractBinary reads a gzip-compressed tar archive from r and returns the
// contents of the regular file whose base name equals name.
func ExtractBinary(r io.Reader, name string) ([]byte, error) {
	fsys, err := ghfetch.UntarGz(r, ghfetch.Options{})
	if err != nil {
		return nil, err
	}
	var found []byte
	walkErr := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			// Only relevant when RejectUnsafePaths is off (as it is here):
			// skip whatever this fs.FS implementation couldn't stat and
			// keep looking, rather than letting one odd entry abort the
			// search for the binary we actually want.
			return nil
		}
		if d.IsDir() || filepath.Base(p) != name {
			return nil
		}
		b, readErr := fs.ReadFile(fsys, p)
		if readErr != nil {
			return readErr
		}
		found = b
		return fs.SkipAll
	})
	if walkErr != nil {
		return nil, walkErr
	}
	if found == nil {
		return nil, fmt.Errorf("binary %q not found in archive", name)
	}
	return found, nil
}

func writeAtomic(dest string, body []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dest), ".viber-update-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, dest)
}

// DefaultDest returns $HOME/bin/<binary>, the install path used by `make install`.
func DefaultDest(binary string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "bin", binary), nil
}
