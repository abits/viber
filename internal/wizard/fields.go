package wizard

import (
	"fmt"
	"regexp"
	"strings"
)

type field struct {
	prompt     string
	hint       string
	required   bool
	validate   func(string) error
	getDefault func(Answers) string
	set        func(*Answers, string)
}

var (
	nameRe   = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*$`)
	moduleRe = regexp.MustCompile(`^[a-zA-Z0-9._~-]+(?:/[a-zA-Z0-9._~-]+)+$`)
)

var fields = []field{
	{
		prompt:   "Project name",
		hint:     "letters/digits/dot/dash/underscore. Used as the title.",
		required: true,
		validate: func(s string) error {
			if !nameRe.MatchString(s) {
				return fmt.Errorf("invalid name; use [A-Za-z][A-Za-z0-9._-]*")
			}
			return nil
		},
		getDefault: func(a Answers) string { return a.Name },
		set:        func(a *Answers, s string) { a.Name = s },
	},
	{
		prompt:   "Directory",
		hint:     "Where to create the project. Defaults to ./<name>.",
		required: true,
		validate: func(s string) error {
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf("directory required")
			}
			return nil
		},
		getDefault: func(a Answers) string {
			if a.Dir != "" {
				return a.Dir
			}
			if a.Name != "" {
				return "./" + a.Name
			}
			return ""
		},
		set: func(a *Answers, s string) { a.Dir = s },
	},
	{
		prompt:   "Module path",
		hint:     "e.g. github.com/<you>/<name>. Leave empty if not a Go project.",
		required: false,
		validate: func(s string) error {
			if s == "" {
				return nil
			}
			if !moduleRe.MatchString(s) {
				return fmt.Errorf("module path should look like host/org/name")
			}
			return nil
		},
		getDefault: func(a Answers) string { return a.Module },
		set:        func(a *Answers, s string) { a.Module = s },
	},
	{
		prompt:     "One-line description",
		hint:       "Shown in README.md and CLAUDE.md.",
		validate:   func(string) error { return nil },
		getDefault: func(a Answers) string { return a.Description },
		set:        func(a *Answers, s string) { a.Description = s },
	},
	{
		prompt:     "Git remote (optional)",
		hint:       "e.g. git@github.com:<you>/<name>.git. Leave empty to skip.",
		validate:   func(string) error { return nil },
		getDefault: func(a Answers) string { return a.Remote },
		set:        func(a *Answers, s string) { a.Remote = s },
	},
}
