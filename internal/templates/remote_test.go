package templates

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io/fs"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

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

func TestUnpackTarGzRejectsUnsafePaths(t *testing.T) {
	for _, name := range []string{
		"repo-main/../escape.txt",
		"/absolute.txt",
		`repo-main/..\escape.txt`,
	} {
		t.Run(name, func(t *testing.T) {
			archive := tarGzWithFile(t, name, []byte("unsafe"))
			_, err := unpackTarGz(archive)
			if err == nil {
				t.Fatal("expected unsafe path error")
			}
			if !strings.Contains(err.Error(), strconv.Quote(name)) {
				t.Fatalf("error %q does not name unsafe entry %q", err, name)
			}
		})
	}
}

func TestUnpackTarGzRejectsOversizedTemplateSet(t *testing.T) {
	body := bytes.Repeat([]byte("x"), int(maxTemplateSetSize)+1)
	archive := tarGzWithFile(t, "repo-main/large.txt", body)

	_, err := unpackTarGz(archive)
	if err == nil {
		t.Fatal("expected template set size error")
	}
	if !strings.Contains(err.Error(), "maximum uncompressed size") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func tarGzWithFile(t *testing.T, name string, body []byte) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf
}
