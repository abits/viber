// Package templates resolves a template set — either the one embedded in the
// binary or a tarball fetched from GitHub — and renders it into a destination
// directory.
package templates

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed all:default
var defaultFS embed.FS

// defaultRef is the git ref used when --from names no explicit one.
const defaultRef = "main"

// Default returns the template set embedded in the binary.
func Default() (fs.FS, error) {
	return fs.Sub(defaultFS, "default")
}

// Resolve returns the template set selected by from, which is either empty (the
// embedded set) or a GitHub reference in the form owner/repo[@ref].
func Resolve(ctx context.Context, from string) (fs.FS, error) {
	if from == "" {
		return Default()
	}
	owner, repo, ref, err := parseFrom(from)
	if err != nil {
		return nil, err
	}
	return Fetch(ctx, owner, repo, ref)
}

// parseFrom splits a --from value of the form owner/repo[@ref]. The ref
// defaults to defaultRef; a trailing "@" with nothing after it is an error
// rather than an empty ref, which would build an unfetchable URL.
func parseFrom(s string) (owner, repo, ref string, err error) {
	const want = "want owner/repo[@ref]"
	ref = defaultRef
	if i := strings.Index(s, "@"); i >= 0 {
		ref = s[i+1:]
		s = s[:i]
		if ref == "" {
			return "", "", "", fmt.Errorf("invalid --from value %q: empty ref after '@'; %s", s+"@", want)
		}
	}
	owner, repo, found := strings.Cut(s, "/")
	if !found || owner == "" || repo == "" {
		return "", "", "", fmt.Errorf("invalid --from value %q; %s", s, want)
	}
	return owner, repo, ref, nil
}
