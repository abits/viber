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

type Data struct {
	Name        string
	Module      string
	Description string
	Remote      string
	Year        int
}

var ErrExists = errors.New("destination file already exists (use --force to overwrite)")

func Render(src fs.FS, dst string, data Data, force bool) error {
	return fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == "." {
			return os.MkdirAll(dst, 0o755)
		}
		out := filepath.Join(dst, filepath.FromSlash(strings.TrimSuffix(p, ".tmpl")))
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
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
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if strings.HasSuffix(in, ".tmpl") {
		tmpl, err := template.New(filepath.Base(in)).Option("missingkey=error").Parse(string(content))
		if err != nil {
			return fmt.Errorf("parse %s: %w", in, err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("execute %s: %w", in, err)
		}
		return os.WriteFile(out, buf.Bytes(), 0o644)
	}
	return os.WriteFile(out, content, 0o644)
}
