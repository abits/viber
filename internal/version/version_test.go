package version

import "testing"

func TestInfoString(t *testing.T) {
	cases := []struct {
		name string
		info Info
		want string
	}{
		{
			name: "all set",
			info: Info{Version: "1.2.3", Commit: "abc1234", Date: "2026-01-01T00:00:00Z"},
			want: "viber 1.2.3 (commit abc1234, built 2026-01-01T00:00:00Z)",
		},
		{
			name: "zero value falls back on every field",
			info: Info{},
			want: "viber dev (commit unknown, built unknown)",
		},
		{"missing version", Info{Commit: "c", Date: "d"}, "viber dev (commit c, built d)"},
		{"missing commit", Info{Version: "v", Date: "d"}, "viber v (commit unknown, built d)"},
		{"missing date", Info{Version: "v", Commit: "c"}, "viber v (commit c, built unknown)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.info.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}
