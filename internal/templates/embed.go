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

func Default() (fs.FS, error) {
	return fs.Sub(defaultFS, "default")
}

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

func parseFrom(s string) (owner, repo, ref string, err error) {
	ref = "main"
	if i := strings.Index(s, "@"); i >= 0 {
		ref = s[i+1:]
		s = s[:i]
	}
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", fmt.Errorf("invalid --from value %q; want owner/repo[@ref]", s)
	}
	return parts[0], parts[1], ref, nil
}
