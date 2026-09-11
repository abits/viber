package updater

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const releaseAPI = "https://api.github.com/repos/%s/%s/releases/latest"

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func Latest(ctx context.Context, owner, repo string) (*Release, error) {
	url := fmt.Sprintf(releaseAPI, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := http.DefaultClient.Do(req)
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

func Install(ctx context.Context, rel *Release, binary, dest string) (string, error) {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, chosen.URL, nil)
	if err != nil {
		return "", err
	}
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := http.DefaultClient.Do(req)
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

func ExtractBinary(r io.Reader, name string) ([]byte, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(hdr.Name) == name {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("binary %q not found in archive", name)
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

func DefaultDest(binary string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "bin", binary), nil
}
