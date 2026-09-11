package version

import "fmt"

type Info struct {
	Version string
	Commit  string
	Date    string
}

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
