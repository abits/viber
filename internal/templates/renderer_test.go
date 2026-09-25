package templates

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestRender(t *testing.T) {
	src := fstest.MapFS{
		"README.md.tmpl":     &fstest.MapFile{Data: []byte("# {{.Name}}\n{{.Description}}\n")},
		"hello.txt":          &fstest.MapFile{Data: []byte("plain\n")},
		"sub/nested.md.tmpl": &fstest.MapFile{Data: []byte("nested: {{.Description}}\n")},
	}
	dst := t.TempDir()
	data := Data{Name: "myproj", Description: "hello"}

	if err := Render(src, dst, data, false); err != nil {
		t.Fatal(err)
	}

	cases := map[string]string{
		"README.md":     "# myproj\nhello\n",
		"hello.txt":     "plain\n",
		"sub/nested.md": "nested: hello\n",
	}
	for name, want := range cases {
		got, err := os.ReadFile(filepath.Join(dst, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if string(got) != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestRenderRefusesOverwrite(t *testing.T) {
	src := fstest.MapFS{"a.txt": &fstest.MapFile{Data: []byte("x")}}
	dst := t.TempDir()
	if err := os.WriteFile(filepath.Join(dst, "a.txt"), []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := Render(src, dst, Data{}, false)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected already-exists error, got %v", err)
	}
}

func TestRenderForceOverwrites(t *testing.T) {
	src := fstest.MapFS{"a.txt": &fstest.MapFile{Data: []byte("new")}}
	dst := t.TempDir()
	if err := os.WriteFile(filepath.Join(dst, "a.txt"), []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Render(src, dst, Data{}, true); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("want %q, got %q", "new", got)
	}
}

// TestRenderCreatesFreshDestinationAtomically pins the common case: when dst
// does not yet exist, Render creates it with a single rename rather than
// writing files into it one at a time, and leaves no temporary directory
// behind next to it.
func TestRenderCreatesFreshDestinationAtomically(t *testing.T) {
	src := fstest.MapFS{
		"a.txt":          &fstest.MapFile{Data: []byte("x")},
		"sub/b.txt.tmpl": &fstest.MapFile{Data: []byte("{{.Name}}")},
	}
	parent := t.TempDir()
	dst := filepath.Join(parent, "proj")

	if err := Render(src, dst, Data{Name: "hi"}, false); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil || string(got) != "x" {
		t.Fatalf("a.txt = %q, %v", got, err)
	}
	got, err = os.ReadFile(filepath.Join(dst, "sub/b.txt"))
	if err != nil || string(got) != "hi" {
		t.Fatalf("sub/b.txt = %q, %v", got, err)
	}

	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "proj" {
		t.Fatalf("parent dir = %v, want exactly [proj] (no leftover temp dir)", entries)
	}
}

// TestRenderLeavesNoTraceOnFailure covers issue #3's acceptance criterion: a
// render that fails partway through (here, a template that fails to parse)
// must leave dst untouched - it must not exist at all - and must not leave
// a stray temporary directory next to it either.
func TestRenderLeavesNoTraceOnFailure(t *testing.T) {
	src := fstest.MapFS{
		"good.txt":        &fstest.MapFile{Data: []byte("fine")},
		"broken.txt.tmpl": &fstest.MapFile{Data: []byte("{{.Unclosed")},
	}
	parent := t.TempDir()
	dst := filepath.Join(parent, "proj")

	err := Render(src, dst, Data{Name: "p"}, false)
	if err == nil {
		t.Fatal("want an error from the unparseable template")
	}

	if _, statErr := os.Stat(dst); !os.IsNotExist(statErr) {
		t.Fatalf("dst = %v after a failed render, want it to not exist", statErr)
	}
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("parent dir has %d leftover entries after a failed render, want 0: %v", len(entries), entries)
	}
}

// TestRenderMergeStopsAtFirstConflict documents the accepted, narrower
// guarantee for the case Render cannot make atomic: when dst already
// exists, a conflicting file still stops the merge with ErrExists, but
// files already copied in before the conflict stay on disk - the same
// partial-write behavior Render always had for this case, now confined to
// files actually merged into an existing dst rather than every fresh
// scaffold.
func TestRenderMergeStopsAtFirstConflict(t *testing.T) {
	src := fstest.MapFS{
		"a.txt": &fstest.MapFile{Data: []byte("a")},
		"b.txt": &fstest.MapFile{Data: []byte("b")},
	}
	dst := t.TempDir()
	if err := os.WriteFile(filepath.Join(dst, "b.txt"), []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Render(src, dst, Data{}, false)
	if err == nil || !errors.Is(err, ErrExists) {
		t.Fatalf("err = %v, want ErrExists", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "a.txt")); err != nil {
		t.Errorf("a.txt should have been merged in before the conflict on b.txt: %v", err)
	}
}

func TestRenderRejectsTwoSourcesForOnePath(t *testing.T) {
	src := fstest.MapFS{
		"cfg/settings.json":      &fstest.MapFile{Data: []byte("plain")},
		"cfg/settings.json.tmpl": &fstest.MapFile{Data: []byte("{{.Name}}")},
	}
	dst := filepath.Join(t.TempDir(), "proj")

	err := Render(src, dst, Data{Name: "p"}, false)

	if !errors.Is(err, ErrDuplicateTarget) {
		t.Fatalf("err = %v, want ErrDuplicateTarget", err)
	}
	for _, name := range []string{"cfg/settings.json", "cfg/settings.json.tmpl"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not name source %q", err, name)
		}
	}
	if _, statErr := os.Stat(dst); !os.IsNotExist(statErr) {
		t.Errorf("dst = %v after a duplicate-target render, want it to not exist", statErr)
	}
}
