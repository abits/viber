package templates

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io/fs"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func tarGzWith(t *testing.T, entries map[string]string) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range entries {
		hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}
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

func TestUnpackTarGzStripsTopDir(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	entries := map[string]string{
		"repo-main/README.md":    "hello",
		"repo-main/dir/file.txt": "world",
	}
	for name, body := range entries {
		hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	_ = tw.Close()
	_ = gz.Close()

	got, err := unpackTarGz(&buf)
	if err != nil {
		t.Fatal(err)
	}

	var paths []string
	_ = fs.WalkDir(got, ".", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			paths = append(paths, p)
		}
		return nil
	})
	sort.Strings(paths)
	want := []string{"README.md", "dir/file.txt"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths=%v want %v", paths, want)
	}

	b, err := fs.ReadFile(got, "README.md")
	if err != nil || string(b) != "hello" {
		t.Fatalf("README.md contents wrong: %q err %v", b, err)
	}
}

func TestUnpackTarGzRejectsDotDot(t *testing.T) {
	buf := tarGzWith(t, map[string]string{
		"repo-main/README.md":    "ok",
		"repo-main/../pwned.txt": "nope",
	})
	_, err := unpackTarGz(buf)
	if err == nil {
		t.Fatal("expected error for ../ entry")
	}
	if !strings.Contains(err.Error(), "repo-main/../pwned.txt") {
		t.Fatalf("error should name the entry, got %v", err)
	}
}

func TestUnpackTarGzRejectsAbsolutePath(t *testing.T) {
	buf := tarGzWith(t, map[string]string{
		"/abs/secret.txt": "nope",
	})
	_, err := unpackTarGz(buf)
	if err == nil {
		t.Fatal("expected error for absolute path")
	}
	if !strings.Contains(err.Error(), "/abs/secret.txt") {
		t.Fatalf("error should name the entry, got %v", err)
	}
}

func TestUnpackTarGzRejectsOverLimit(t *testing.T) {
	old := maxUncompressedBytes
	maxUncompressedBytes = 32
	t.Cleanup(func() { maxUncompressedBytes = old })

	buf := tarGzWith(t, map[string]string{
		"repo-main/big.txt": strings.Repeat("x", 64),
	})
	_, err := unpackTarGz(buf)
	if err == nil {
		t.Fatal("expected error for over-limit archive")
	}
	if !strings.Contains(err.Error(), "uncompressed limit") {
		t.Fatalf("error should mention the limit, got %v", err)
	}
}
