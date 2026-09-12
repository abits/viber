package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

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
