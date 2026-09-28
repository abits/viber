package doctor

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func TestProbe(t *testing.T) {
	check := Check{
		Name:        "tool",
		Args:        []string{"tool", "--version"},
		Remediation: "install tool",
	}

	t.Run("present and reporting version", func(t *testing.T) {
		orig := execCommand
		defer func() { execCommand = orig }()
		execCommand = func(ctx context.Context, _ string, args ...string) *exec.Cmd {
			// `echo 1.2.3` exits 0 and prints a version string.
			return exec.CommandContext(ctx, "sh", append([]string{"-c", "echo 1.2.3"}, args...)...)
		}
		r := probe(context.Background(), check)
		if r.Status != StatusOK {
			t.Errorf("status = %q, want ok", r.Status)
		}
		if r.Version != "1.2.3" {
			t.Errorf("version = %q, want 1.2.3", r.Version)
		}
	})

	t.Run("missing binary", func(t *testing.T) {
		orig := execCommand
		defer func() { execCommand = orig }()
		execCommand = func(ctx context.Context, _ string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "nonexistent-viber-tool-xyz", args...)
		}
		r := probe(context.Background(), check)
		if r.Status != StatusWarning {
			t.Errorf("status = %q, want warning", r.Status)
		}
		if r.Problem != "not installed" {
			t.Errorf("problem = %q, want 'not installed'", r.Problem)
		}
	})

	t.Run("missing binary required", func(t *testing.T) {
		orig := execCommand
		defer func() { execCommand = orig }()
		execCommand = func(ctx context.Context, _ string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "nonexistent-viber-tool-xyz", args...)
		}
		req := check
		req.Required = true
		r := probe(context.Background(), req)
		if r.Status != StatusError {
			t.Errorf("status = %q, want error", r.Status)
		}
	})

	t.Run("non-zero exit", func(t *testing.T) {
		orig := execCommand
		defer func() { execCommand = orig }()
		execCommand = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "sh", "-c", "echo 'auth failed' >&2; exit 1")
		}
		r := probe(context.Background(), check)
		if r.Status != StatusWarning {
			t.Errorf("status = %q, want warning", r.Status)
		}
		if r.Problem == "" {
			t.Error("problem should be non-empty for non-zero exit")
		}
	})

	t.Run("timeout", func(t *testing.T) {
		orig := execCommand
		defer func() { execCommand = orig }()
		// Replace the package timeout so the test doesn't take 5 s.
		origTimeout := probeTimeout
		probeTimeout = 50 * time.Millisecond
		defer func() { probeTimeout = origTimeout }()

		execCommand = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "sleep", "10")
		}
		r := probe(context.Background(), check)
		if r.Status != StatusWarning {
			t.Errorf("status = %q, want warning", r.Status)
		}
		if r.Problem != "timed out" {
			t.Errorf("problem = %q, want 'timed out'", r.Problem)
		}
	})

	t.Run("unparsable version falls back to first line", func(t *testing.T) {
		orig := execCommand
		defer func() { execCommand = orig }()
		execCommand = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
			return exec.CommandContext(ctx, "sh", "-c", "echo 'nightly-build-xyz'; echo second line")
		}
		r := probe(context.Background(), check)
		if r.Status != StatusOK {
			t.Errorf("status = %q, want ok", r.Status)
		}
		if r.Version != "nightly-build-xyz" {
			t.Errorf("version = %q, want 'nightly-build-xyz'", r.Version)
		}
	})
}

func TestExtractVersion(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"git version 2.43.0", "2.43.0"},
		{"gh version 2.40.1 (2024-01-01)", "2.40.1"},
		{"1.13.2\n", "1.13.2"},
		{"nightly-build", "nightly-build"},
		{"", ""},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := extractVersion(tc.input)
			if got != tc.want {
				t.Errorf("extractVersion(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
