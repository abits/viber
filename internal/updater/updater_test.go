package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/abits/viber/internal/ghfetch"
)

// useTestServer points both the release-API endpoint and the HTTP client at
// srv for the duration of the test, restoring both on cleanup. This is the
// seam issue #2 asked for: Latest and Install previously had none, which is
// why internal/updater sat at 0% coverage.
func useTestServer(t *testing.T, srv *httptest.Server) {
	t.Helper()
	origAPI, origClient := releaseAPI, client
	releaseAPI = srv.URL + "/repos/%s/%s/releases/latest"
	client = ghfetch.Client{HTTP: srv.Client()}
	t.Cleanup(func() {
		releaseAPI = origAPI
		client = origClient
	})
}

func tarGz(t *testing.T, files map[string]string) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		hdr := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf
}

func TestExtractBinary(t *testing.T) {
	t.Run("finds the binary next to other files", func(t *testing.T) {
		archive := tarGz(t, map[string]string{
			"README.md": "docs",
			"LICENSE":   "mit",
			"viber":     "ELF-ish",
		})
		got, err := ExtractBinary(archive, "viber")
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "ELF-ish" {
			t.Errorf("got %q, want %q", got, "ELF-ish")
		}
	})

	t.Run("finds the binary in a nested directory", func(t *testing.T) {
		archive := tarGz(t, map[string]string{"viber_1.0.0_linux_amd64/viber": "nested"})
		got, err := ExtractBinary(archive, "viber")
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "nested" {
			t.Errorf("got %q, want %q", got, "nested")
		}
	})

	t.Run("reports a missing binary", func(t *testing.T) {
		archive := tarGz(t, map[string]string{"README.md": "docs"})
		_, err := ExtractBinary(archive, "viber")
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Fatalf("err = %v, want a not-found error", err)
		}
	})

	t.Run("rejects a non-gzip stream", func(t *testing.T) {
		if _, err := ExtractBinary(strings.NewReader("plain text"), "viber"); err == nil {
			t.Fatal("want an error for a non-gzip stream")
		}
	})

	t.Run("does not match an entry whose name traverses out of the archive", func(t *testing.T) {
		// A tar entry named "../viber" has base name "viber" too, but
		// fstest.MapFS (which UntarGz builds on) treats ".." as a directory
		// element rather than a literal name and never exposes such an
		// entry's content through the fs.FS interface. ExtractBinary must
		// keep looking rather than error out, and must not find this one.
		archive := tarGz(t, map[string]string{
			"../viber":                      "evil",
			"viber_1.0.0_linux_amd64/viber": "good",
		})
		got, err := ExtractBinary(archive, "viber")
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "good" {
			t.Errorf("got %q, want %q (the traversal entry must not win)", got, "good")
		}
	})
}

func TestWriteAtomic(t *testing.T) {
	t.Run("creates parent directories", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "nested", "deeper", "viber")
		if err := writeAtomic(dest, []byte("payload"), 0o755); err != nil {
			t.Fatal(err)
		}
		assertFile(t, dest, "payload", 0o755)
	})

	t.Run("replaces an existing binary", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "viber")
		if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := writeAtomic(dest, []byte("new"), 0o755); err != nil {
			t.Fatal(err)
		}
		assertFile(t, dest, "new", 0o755)
	})

	t.Run("leaves no temp files behind", func(t *testing.T) {
		dir := t.TempDir()
		if err := writeAtomic(filepath.Join(dir, "viber"), []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Errorf("dir holds %d entries, want just the destination", len(entries))
		}
	})
}

func assertFile(t *testing.T, path, wantBody string, wantMode os.FileMode) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != wantBody {
		t.Errorf("body = %q, want %q", got, wantBody)
	}
	if runtime.GOOS == "windows" {
		// Windows only models the read-only bit, and `viber update` refuses to
		// run there anyway, so the exact permission bits are not meaningful.
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != wantMode {
		t.Errorf("mode = %v, want %v", info.Mode().Perm(), wantMode)
	}
}

func TestLatest(t *testing.T) {
	t.Run("returns the decoded release", func(t *testing.T) {
		var gotAccept, gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAccept = r.Header.Get("Accept")
			gotPath = r.URL.Path
			_ = json.NewEncoder(w).Encode(Release{
				TagName: "v1.2.3",
				Assets:  []Asset{{Name: "viber_1.2.3_linux_amd64.tar.gz", URL: "http://example.invalid/asset"}},
			})
		}))
		defer srv.Close()
		useTestServer(t, srv)

		rel, err := Latest(context.Background(), "abits", "viber")
		if err != nil {
			t.Fatal(err)
		}
		if rel.TagName != "v1.2.3" {
			t.Errorf("TagName = %q, want %q", rel.TagName, "v1.2.3")
		}
		if len(rel.Assets) != 1 || rel.Assets[0].Name != "viber_1.2.3_linux_amd64.tar.gz" {
			t.Errorf("Assets = %+v", rel.Assets)
		}
		if want := "application/vnd.github+json"; gotAccept != want {
			t.Errorf("Accept header = %q, want %q", gotAccept, want)
		}
		if want := "/repos/abits/viber/releases/latest"; gotPath != want {
			t.Errorf("request path = %q, want %q", gotPath, want)
		}
	})

	t.Run("sends the configured bearer token", func(t *testing.T) {
		var gotAuth string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			_ = json.NewEncoder(w).Encode(Release{TagName: "v1.0.0"})
		}))
		defer srv.Close()
		useTestServer(t, srv)
		orig := client
		client.Token = "test-token"
		defer func() { client = orig }()

		if _, err := Latest(context.Background(), "abits", "viber"); err != nil {
			t.Fatal(err)
		}
		if want := "Bearer test-token"; gotAuth != want {
			t.Errorf("Authorization = %q, want %q", gotAuth, want)
		}
	})

	t.Run("wraps a non-200 status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "nope", http.StatusNotFound)
		}))
		defer srv.Close()
		useTestServer(t, srv)

		_, err := Latest(context.Background(), "abits", "viber")
		if err == nil || !strings.Contains(err.Error(), "404") {
			t.Fatalf("err = %v, want it to mention 404", err)
		}
	})
}

func TestInstall(t *testing.T) {
	t.Run("downloads, extracts, and atomically installs the binary", func(t *testing.T) {
		assetName := fmt.Sprintf("viber_1.2.3_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		archive := tarGz(t, map[string]string{
			"viber_1.2.3_" + runtime.GOOS + "_" + runtime.GOARCH + "/viber": "compiled-binary-bytes",
		})
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(archive.Bytes())
		}))
		defer srv.Close()
		useTestServer(t, srv)

		rel := &Release{
			TagName: "v1.2.3",
			Assets:  []Asset{{Name: assetName, URL: srv.URL}},
		}
		dest := filepath.Join(t.TempDir(), "viber")

		gotVer, err := Install(context.Background(), rel, "viber", dest)
		if err != nil {
			t.Fatal(err)
		}
		if gotVer != "v1.2.3" {
			t.Errorf("version = %q, want %q", gotVer, "v1.2.3")
		}
		got, err := os.ReadFile(dest)
		if err != nil || string(got) != "compiled-binary-bytes" {
			t.Fatalf("installed content = %q, %v", got, err)
		}
	})

	t.Run("errors when no asset matches the current platform", func(t *testing.T) {
		rel := &Release{
			TagName: "v1.2.3",
			Assets:  []Asset{{Name: "viber_1.2.3_plan9_amd64.tar.gz"}},
		}
		_, err := Install(context.Background(), rel, "viber", filepath.Join(t.TempDir(), "viber"))
		if err == nil || !strings.Contains(err.Error(), "no release asset matches") {
			t.Fatalf("err = %v, want a no-matching-asset error", err)
		}
	})

	t.Run("wraps a non-200 asset download", func(t *testing.T) {
		assetName := fmt.Sprintf("viber_1.2.3_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "gone", http.StatusGone)
		}))
		defer srv.Close()
		useTestServer(t, srv)

		rel := &Release{TagName: "v1.2.3", Assets: []Asset{{Name: assetName, URL: srv.URL}}}
		_, err := Install(context.Background(), rel, "viber", filepath.Join(t.TempDir(), "viber"))
		if err == nil || !strings.Contains(err.Error(), "410") {
			t.Fatalf("err = %v, want it to mention 410", err)
		}
	})
}

func TestInstallRefusesWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("this guard only triggers on windows")
	}
	_, err := Install(context.Background(), &Release{TagName: "v1.0.0"}, "viber", filepath.Join(t.TempDir(), "viber"))
	if err == nil {
		t.Fatal("want an error on windows")
	}
}

func TestDefaultDest(t *testing.T) {
	home := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
	} else {
		t.Setenv("HOME", home)
	}
	got, err := DefaultDest("viber")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, "bin", "viber"); got != want {
		t.Errorf("DefaultDest() = %q, want %q", got, want)
	}
}
