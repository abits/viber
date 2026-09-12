package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/abits/viber/internal/version"
)

// newTestRoot builds the real command tree with its output captured.
func newTestRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	root := newRootCmd(version.Info{Version: "0.0.0-test", Commit: "abc", Date: "today"})
	var out, errb bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errb)
	root.SetArgs(args)
	return root, &out, &errb
}

func TestExitCode(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil is success", nil, exitOK},
		{"plain error is runtime failure", errors.New("boom"), exitFailure},
		{"usage error", UsageError(nil, errors.New("bad flag")), exitUsage},
		{"wrapped usage error", errWrap(UsageError(nil, errors.New("bad flag"))), exitUsage},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := exitCode(tc.err); got != tc.want {
				t.Errorf("exitCode(%v) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}
}

func errWrap(err error) error { return errors.Join(errors.New("context"), err) }

// TestUsageErrorsAreReported guards the contract that no failure exits
// silently: every non-zero exit must tell the user something.
func TestUsageErrorsAreReported(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"init without name in non-interactive mode", []string{"init", "--no-tui"}},
		{"unknown flag", []string{"init", "--bogus"}},
		{"unknown subcommand flag on update", []string{"update", "--repo", "not-a-repo-spec"}},
		{"too many positional args", []string{"init", "a", "b", "c"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, _, errb := newTestRoot(t, tc.args...)
			if got := execute(context.Background(), root); got != exitUsage {
				t.Errorf("exit = %d, want %d", got, exitUsage)
			}
			if errb.Len() == 0 {
				t.Error("usage error produced no output at all")
			}
			if !strings.Contains(errb.String(), "Error:") {
				t.Errorf("stderr missing error line:\n%s", errb)
			}
			if !strings.Contains(errb.String(), "Usage:") {
				t.Errorf("usage error did not print usage:\n%s", errb)
			}
		})
	}
}

// TestErrorPrintedOnce pins A3: the runners report status, execute prints text.
func TestErrorPrintedOnce(t *testing.T) {
	root, _, errb := newTestRoot(t, "init", "--no-tui")
	execute(context.Background(), root)
	if n := strings.Count(errb.String(), "name and destination directory are required"); n != 1 {
		t.Errorf("error text appears %d times, want exactly 1:\n%s", n, errb)
	}
}

func TestVersionCommand(t *testing.T) {
	root, out, _ := newTestRoot(t, "version")
	if got := execute(context.Background(), root); got != exitOK {
		t.Fatalf("exit = %d, want %d", got, exitOK)
	}
	if !strings.Contains(out.String(), "0.0.0-test") {
		t.Errorf("version output = %q, want it to contain the version", out)
	}
}

func TestSplitRepoSpec(t *testing.T) {
	cases := []struct {
		in          string
		owner, repo string
		wantErr     bool
	}{
		{"abits/viber", "abits", "viber", false},
		{"o/r/extra", "o", "r/extra", false},
		{"", "", "", true},
		{"noslash", "", "", true},
		{"/viber", "", "", true},
		{"abits/", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			owner, repo, err := splitRepoSpec(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("splitRepoSpec(%q) = %q,%q; want error", tc.in, owner, repo)
				}
				return
			}
			if err != nil {
				t.Fatalf("splitRepoSpec(%q): %v", tc.in, err)
			}
			if owner != tc.owner || repo != tc.repo {
				t.Errorf("splitRepoSpec(%q) = %q,%q; want %q,%q", tc.in, owner, repo, tc.owner, tc.repo)
			}
		})
	}
}

func TestCheckDest(t *testing.T) {
	t.Run("missing dir is fine", func(t *testing.T) {
		if err := checkDest(t.TempDir()+"/nope", false); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("empty dir is fine", func(t *testing.T) {
		if err := checkDest(t.TempDir(), false); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("non-empty dir is rejected", func(t *testing.T) {
		dir := t.TempDir()
		mustWrite(t, dir+"/file", "x")
		if err := checkDest(dir, false); err == nil {
			t.Error("want error for non-empty destination")
		}
	})
	t.Run("force overrides", func(t *testing.T) {
		dir := t.TempDir()
		mustWrite(t, dir+"/file", "x")
		if err := checkDest(dir, true); err != nil {
			t.Errorf("unexpected error with force: %v", err)
		}
	})
}
