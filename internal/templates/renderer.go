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

// File modes used for rendered output.
const (
	dirMode  = 0o755
	fileMode = 0o644
)

// Render expands every file in src into dst, stripping the ".tmpl" suffix and
// executing those files as text/template against data. Plain files are copied
// verbatim. Unless force is set, an existing destination file is an error.
//
// Render is not atomic: a failure partway through leaves the files written so
// far in place.
func Render(src fs.FS, dst string, data Data, force bool) error {
	return fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == "." {
			return os.MkdirAll(dst, dirMode)
		}
		out := filepath.Join(dst, filepath.FromSlash(strings.TrimSuffix(p, ".tmpl")))
		if d.IsDir() {
			return os.MkdirAll(out, dirMode)
		}
		if !force {
			if _, err := os.Stat(out); err == nil {
				return fmt.Errorf("%w: %s", ErrExists, out)
			}
		}
		return writeFile(src, p, out, data)
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
