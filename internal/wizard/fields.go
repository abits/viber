package wizard

import (
	"fmt"
	"regexp"
	"strings"
)

// field describes one prompt in the wizard. A nil validate accepts any value.
type field struct {
	prompt     string
	hint       string
	required   bool
	validate   func(string) error
	getDefault func(Answers) string
	set        func(*Answers, string)
}

var nameRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*$`)

// fields is the prompt sequence, in the order the wizard asks them.
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
		prompt:     "One-line description",
		hint:       "Shown in README.md and CLAUDE.md.",
		getDefault: func(a Answers) string { return a.Description },
		set:        func(a *Answers, s string) { a.Description = s },
	},
	{
		prompt:     "Git remote (optional)",
		hint:       "e.g. git@github.com:<you>/<name>.git. Leave empty to skip.",
		getDefault: func(a Answers) string { return a.Remote },
		set:        func(a *Answers, s string) { a.Remote = s },
	},
}
