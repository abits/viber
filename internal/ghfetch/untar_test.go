package ghfetch

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io/fs"
	"strconv"
	"strings"
	"testing"
)

type tarEntry struct {
	name     string
	body     string
	typeflag byte
	linkname string
}

func tarGzArchive(t *testing.T, entries []tarEntry) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		flag := e.typeflag
		if flag == 0 {
			flag = tar.TypeReg
		}
		hdr := &tar.Header{
			Name:     e.name,
			Mode:     0o644,
			Size:     int64(len(e.body)),
			Typeflag: flag,
			Linkname: e.linkname,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if flag == tar.TypeReg {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
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

func TestUntarGzExtractsRegularFilesOnly(t *testing.T) {
	archive := tarGzArchive(t, []tarEntry{
		{name: "dir/", typeflag: tar.TypeDir},
		{name: "dir/file.txt", body: "hello"},
		{name: "link", typeflag: tar.TypeSymlink, linkname: "dir/file.txt"},
	})

	fsys, err := UntarGz(archive, Options{})
	if err != nil {
		t.Fatal(err)
	}

	b, err := fs.ReadFile(fsys, "dir/file.txt")
	if err != nil || string(b) != "hello" {
		t.Fatalf("dir/file.txt = %q, %v; want %q, nil", b, err, "hello")
	}
	if _, err := fs.Stat(fsys, "link"); err == nil {
		t.Error("symlink entry should not have been extracted")
	}
}

func TestUntarGzRejectsNonGzipInput(t *testing.T) {
	if _, err := UntarGz(strings.NewReader("plain text"), Options{}); err == nil {
		t.Fatal("want an error for a non-gzip stream")
	}
}

func TestUntarGzZeroOptionsDoesNotRejectUnusualPaths(t *testing.T) {
	// Matches internal/updater's use: it never writes an entry's tar path to
	// disk (only the content of the one entry it's looking for, to a
	// destination the caller already fixed), so an unusual name in some
	// other entry is not a reason to fail the whole archive.
	archive := tarGzArchive(t, []tarEntry{
		{name: "repo-main/README.md", body: "hello"},
		{name: "../oddly-named.txt", body: "still extracted without error"},
	})
	if _, err := UntarGz(archive, Options{}); err != nil {
		t.Fatalf("zero Options rejected an archive it should have allowed: %v", err)
	}
}

func TestUntarGzRejectUnsafePaths(t *testing.T) {
	for _, name := range []string{
		"repo-main/../escape.txt",
		"/absolute.txt",
		`repo-main/..\escape.txt`,
	} {
		t.Run(name, func(t *testing.T) {
			archive := tarGzArchive(t, []tarEntry{{name: name, body: "unsafe"}})
			_, err := UntarGz(archive, Options{RejectUnsafePaths: true})
			if err == nil {
				t.Fatal("expected an unsafe path error")
			}
			if !strings.Contains(err.Error(), strconv.Quote(name)) {
				t.Fatalf("error %q does not name unsafe entry %q", err, name)
			}
		})
	}
}

func TestUntarGzRejectUnsafePathsAllowsOrdinaryNames(t *testing.T) {
	archive := tarGzArchive(t, []tarEntry{
		{name: "repo-main/README.md", body: "hello"},
		{name: "repo-main/dir/file.txt", body: "world"},
	})
	fsys, err := UntarGz(archive, Options{RejectUnsafePaths: true})
	if err != nil {
		t.Fatal(err)
	}
	if b, err := fs.ReadFile(fsys, "repo-main/README.md"); err != nil || string(b) != "hello" {
		t.Fatalf("repo-main/README.md = %q, %v", b, err)
	}
}

func TestUntarGzMaxUncompressedBytes(t *testing.T) {
	t.Run("rejects a single entry over the limit", func(t *testing.T) {
		body := strings.Repeat("x", 65)
		archive := tarGzArchive(t, []tarEntry{{name: "big.txt", body: body}})
		_, err := UntarGz(archive, Options{MaxUncompressedBytes: 64})
		if err == nil || !strings.Contains(err.Error(), "maximum uncompressed size") {
			t.Fatalf("err = %v, want a maximum-uncompressed-size error", err)
		}
	})

	t.Run("rejects entries that exceed the limit only when summed", func(t *testing.T) {
		archive := tarGzArchive(t, []tarEntry{
			{name: "a.txt", body: strings.Repeat("a", 40)},
			{name: "b.txt", body: strings.Repeat("b", 40)},
		})
		_, err := UntarGz(archive, Options{MaxUncompressedBytes: 64})
		if err == nil || !strings.Contains(err.Error(), "maximum uncompressed size") {
			t.Fatalf("err = %v, want a maximum-uncompressed-size error", err)
		}
	})

	t.Run("allows entries at or under the limit", func(t *testing.T) {
		archive := tarGzArchive(t, []tarEntry{
			{name: "a.txt", body: strings.Repeat("a", 32)},
			{name: "b.txt", body: strings.Repeat("b", 32)},
		})
		fsys, err := UntarGz(archive, Options{MaxUncompressedBytes: 64})
		if err != nil {
			t.Fatal(err)
		}
		if b, err := fs.ReadFile(fsys, "b.txt"); err != nil || len(b) != 32 {
			t.Fatalf("b.txt = %d bytes, %v; want 32 bytes", len(b), err)
		}
	})

	t.Run("zero means unlimited", func(t *testing.T) {
		archive := tarGzArchive(t, []tarEntry{{name: "big.txt", body: strings.Repeat("x", 1<<20)}})
		if _, err := UntarGz(archive, Options{}); err != nil {
			t.Fatalf("unexpected error with no limit set: %v", err)
		}
	})
}
