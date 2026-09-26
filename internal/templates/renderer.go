package templates

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Data is the value every template in a template set is executed against.
//
// Every field here must be referenced by at least one template in the embedded
// set; TestEmbeddedTemplatesUseEveryDataField enforces that, so a field can
// never again be collected from the user and then silently dropped.
type Data struct {
	// Name is the project name, used as the title throughout the scaffold.
	Name string
	// Description is the one-line summary shown in README.md and CLAUDE.md.
	Description string
}

// ErrExists is returned when rendering would overwrite an existing file and
// force is not set. Callers can match it with errors.Is.
var ErrExists = errors.New("destination file already exists (use --force to overwrite)")

// ErrDuplicateTarget is returned when two entries of a template set render
// to the same path, such as "foo" and "foo.tmpl". Callers can match it with
// errors.Is.
var ErrDuplicateTarget = errors.New("template set renders two entries to the same path")

// File modes used for rendered output.
const (
	dirMode  = 0o755
	fileMode = 0o644
)

// Render expands every file in src into dst, stripping the ".tmpl" suffix and
// executing those files as text/template against data. Plain files are
// copied verbatim.
//
// Render is atomic when dst does not yet exist: it renders into a temporary
// directory next to dst (so the final rename can never cross a filesystem
// boundary) and moves it into place only once every file has been written,
// so a failure partway through - a broken template, a full disk - leaves no
// trace at dst at all. When dst already exists (a --force re-run onto a
// previous attempt, or any other caller-managed directory), Render cannot
// offer that guarantee without deleting the caller's existing content
// first, so it falls back to copying the rendered files into dst one by
// one, with the same per-file overwrite rule Render always had: an existing
// file is ErrExists unless force is set.
func Render(src fs.FS, dst string, data Data, force bool) error {
	parent := filepath.Dir(dst)
	if err := os.MkdirAll(parent, dirMode); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(parent, ".viber-render-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	if err := renderAll(src, tmp, data); err != nil {
		return err
	}
	return publish(tmp, dst, force)
}

// renderAll expands every entry of src into the already-created, empty
// directory tmp. Two files that render to the same path are an error rather
// than a silent overwrite, since whichever WalkDir visits last would win.
func renderAll(src fs.FS, tmp string, data Data) error {
	sources := map[string]string{}
	return fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == "." {
			return nil
		}
		target := strings.TrimSuffix(p, ".tmpl")
		out := filepath.Join(tmp, filepath.FromSlash(target))
		if d.IsDir() {
			return os.MkdirAll(out, dirMode)
		}
		if prev, ok := sources[target]; ok {
			return fmt.Errorf("%w: %s and %s both render to %s", ErrDuplicateTarget, prev, p, target)
		}
		sources[target] = p
		return writeFile(src, p, out, data)
	})
}

// publish moves a fully rendered tmp directory into dst: a single rename
// when dst does not exist, or a file-by-file merge (see Render's doc
// comment) when it does.
func publish(tmp, dst string, force bool) error {
	switch _, err := os.Lstat(dst); {
	case errors.Is(err, os.ErrNotExist):
		return os.Rename(tmp, dst)
	case err != nil:
		return err
	default:
		return mergeInto(tmp, dst, force)
	}
}

// mergeInto copies every file under tmp into the corresponding path under
// the already-existing dst, in the same layout Render always produced:
// existing files are left alone and reported as ErrExists unless force is
// set, in which case they're overwritten.
func mergeInto(tmp, dst string, force bool) error {
	tmpFS := os.DirFS(tmp)
	return fs.WalkDir(tmpFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == "." {
			return nil
		}
		out := filepath.Join(dst, filepath.FromSlash(p))
		if d.IsDir() {
			return os.MkdirAll(out, dirMode)
		}
		if !force {
			if _, err := os.Lstat(out); err == nil {
				return fmt.Errorf("%w: %s", ErrExists, out)
			}
		}
		content, err := fs.ReadFile(tmpFS, p)
		if err != nil {
			return err
		}
		return os.WriteFile(out, content, fileMode)
	})
}

func writeFile(src fs.FS, in, out string, data Data) error {
	content, err := fs.ReadFile(src, in)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(out), dirMode); err != nil {
		return err
	}
	if strings.HasSuffix(in, ".tmpl") {
		tmpl, err := template.New(filepath.Base(in)).Parse(string(content))
		if err != nil {
			return fmt.Errorf("parse %s: %w", in, err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("execute %s: %w", in, err)
		}
		return os.WriteFile(out, buf.Bytes(), fileMode)
	}
	return os.WriteFile(out, content, fileMode)
}
