package templates

import (
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
