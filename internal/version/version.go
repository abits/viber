// Package version formats viber's build-time identity.
package version

import "fmt"

// Info holds the version, commit, and build date injected via ldflags.
type Info struct {
	Version string
	Commit  string
	Date    string
}

// String returns a human-readable identity of the form
// "viber <version> (commit <commit>, built <date>)". Empty fields fall back
// to "dev" (version) or "unknown" (commit and date).
func (i Info) String() string {
	v := i.Version
	if v == "" {
		v = "dev"
	}
	c := i.Commit
	if c == "" {
		c = "unknown"
	}
	d := i.Date
	if d == "" {
		d = "unknown"
	}
	return fmt.Sprintf("viber %s (commit %s, built %s)", v, c, d)
}
