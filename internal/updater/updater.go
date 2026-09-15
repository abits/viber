// Package updater downloads GitHub release assets and installs them atomically.
package updater

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// checksumsAssetName is the name GoReleaser publishes the release's combined
// checksum file under; see checksum.name_template in .goreleaser.yaml. It's
// a fixed literal there, not templated, so it's safe to hard-code here too.
const checksumsAssetName = "checksums.txt"

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
// verifies it against the release's checksums.txt, extracts `binary` from
// it, and replaces `dest` atomically (0755). It returns the installed tag
// name. Windows archives are .zip and are not supported.
//
// A release that doesn't publish checksums.txt is an error, not a skipped
// check: a self-updating binary that can't verify what it's about to run
// should refuse rather than proceed unverified.
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

	sums := findAsset(rel, checksumsAssetName)
	if sums == nil {
		return "", fmt.Errorf("%s %s does not publish %s; refusing to install an unverified download",
			binary, rel.TagName, checksumsAssetName)
	}
	wantSum, err := checksumFor(ctx, sums, chosen.Name)
	if err != nil {
		return "", err
	}

	resp, err := client.Get(ctx, chosen.URL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", chosen.URL, resp.Status)
	}

	// Tee the download through the hasher while extracting rather than
	// buffering the whole archive twice: ExtractBinary already reads the
	// stream to completion (UntarGz drains it to EOF, not just to the entry
	// it's looking for), so by the time it returns successfully, hasher has
	// seen every byte of the downloaded archive - exactly what
	// checksums.txt's digest was computed over.
	hasher := sha256.New()
	body, err := ExtractBinary(io.TeeReader(resp.Body, hasher), binary)
	if err != nil {
		return "", err
	}
	if gotSum := hex.EncodeToString(hasher.Sum(nil)); gotSum != wantSum {
		return "", fmt.Errorf("checksum mismatch for %s: got %s, want %s (from %s)",
			chosen.Name, gotSum, wantSum, checksumsAssetName)
	}

	if err := writeAtomic(dest, body, 0o755); err != nil {
		return "", err
	}
	return rel.TagName, nil
}

// findAsset returns the release asset named name, or nil if rel has none.
func findAsset(rel *Release, name string) *Asset {
	for i := range rel.Assets {
		if rel.Assets[i].Name == name {
			return &rel.Assets[i]
		}
	}
	return nil
}

// checksumFor downloads sums (the release's checksums.txt) and returns the
// hex-encoded SHA-256 digest it lists for assetName. GoReleaser's default
// format is sha256sum-compatible: "<hex digest>  <filename>" per line, one
// optional leading '*' on the filename for sha256sum's binary-mode marker.
func checksumFor(ctx context.Context, sums *Asset, assetName string) (string, error) {
	resp, err := client.Get(ctx, sums.URL)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", sums.Name, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", sums.URL, resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}
		digest, name := fields[0], strings.TrimPrefix(fields[1], "*")
		if name == assetName {
			return digest, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read %s: %w", sums.Name, err)
	}
	return "", fmt.Errorf("%s does not list a checksum for %s", sums.Name, assetName)
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
