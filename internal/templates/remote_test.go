package templates

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io/fs"
	"reflect"
	"sort"
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
