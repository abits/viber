package doctor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// mkScaffold creates the three sentinels viber scaffolds carry under dir.
func mkScaffold(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# CLAUDE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{"openspec", ".claude"} {
		if err := os.Mkdir(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDetectProject(t *testing.T) {
	// stubGit returns a factory that mimics `git -C ... rev-parse --show-toplevel`
	// by writing top to stdout and exiting 0; empty top means git failed.
	stubGit := func(top string) func(context.Context, string, ...string) *exec.Cmd {
		return func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
			if top == "" {
				return exec.CommandContext(ctx, "false")
			}
			return exec.CommandContext(ctx, "sh", "-c", "echo "+top)
		}
	}

	t.Run("scaffold at git root", func(t *testing.T) {
		root := t.TempDir()
		mkScaffold(t, root)
		defer ExecCommandForTest(stubGit(root))()

		gotRoot, inGit, ok := DetectProject(context.Background(), root)
		if gotRoot != root || !inGit || !ok {
			t.Errorf("DetectProject = (%q, %v, %v), want (%q, true, true)", gotRoot, inGit, ok, root)
		}
	})

	t.Run("scaffold reached from subdirectory", func(t *testing.T) {
		root := t.TempDir()
		mkScaffold(t, root)
		sub := filepath.Join(root, "internal", "foo")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		defer ExecCommandForTest(stubGit(root))()

		gotRoot, inGit, ok := DetectProject(context.Background(), sub)
		if gotRoot != root || !inGit || !ok {
			t.Errorf("DetectProject = (%q, %v, %v), want (%q, true, true)", gotRoot, inGit, ok, root)
		}
	})

	t.Run("scaffold outside git falls back to cwd", func(t *testing.T) {
		root := t.TempDir()
		mkScaffold(t, root)
		defer ExecCommandForTest(stubGit(""))()

		gotRoot, inGit, ok := DetectProject(context.Background(), root)
		if gotRoot != root || inGit || !ok {
			t.Errorf("DetectProject = (%q, %v, %v), want (%q, false, true)", gotRoot, inGit, ok, root)
		}
	})

	t.Run("missing a sentinel is not a scaffold", func(t *testing.T) {
		root := t.TempDir()
		// Only CLAUDE.md + openspec, no .claude.
		if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(root, "openspec"), 0o755); err != nil {
			t.Fatal(err)
		}
		defer ExecCommandForTest(stubGit(root))()

		gotRoot, inGit, ok := DetectProject(context.Background(), root)
		if gotRoot != root || !inGit || ok {
			t.Errorf("DetectProject = (%q, %v, %v), want (%q, true, false)", gotRoot, inGit, ok, root)
		}
	})
}

func TestSettingsJSONProbe(t *testing.T) {
	root := t.TempDir()
	claudeDir := filepath.Join(root, ".claude")
	if err := os.Mkdir(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run := func() Result {
		return settingsJSONProbe(root)(context.Background(), nil)
	}

	t.Run("valid JSON", func(t *testing.T) {
		write(`{"hooks": {}}`)
		r := run()
		if r.Status != StatusOK {
			t.Errorf("status = %q, want ok (%q)", r.Status, r.Problem)
		}
	})

	t.Run("malformed JSON names the byte offset", func(t *testing.T) {
		write(`{"hooks": }`)
		r := run()
		if r.Status != StatusError {
			t.Fatalf("status = %q, want error", r.Status)
		}
		if !strings.Contains(r.Problem, "byte") {
			t.Errorf("problem %q, want it to name the byte offset", r.Problem)
		}
	})

	t.Run("missing file is an error", func(t *testing.T) {
		if err := os.Remove(filepath.Join(claudeDir, "settings.json")); err != nil {
			t.Fatal(err)
		}
		r := run()
		if r.Status != StatusError {
			t.Fatalf("status = %q, want error", r.Status)
		}
		if !strings.Contains(r.Problem, "missing") {
			t.Errorf("problem %q, want it to mention missing", r.Problem)
		}
	})
}

func TestIntendMDProbe(t *testing.T) {
	root := t.TempDir()
	run := func() Result { return intendMDProbe(root)(context.Background(), nil) }

	t.Run("absent", func(t *testing.T) {
		if r := run(); r.Status != StatusWarning {
			t.Errorf("status = %q, want warning", r.Status)
		}
	})

	t.Run("present", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(root, "intend.md"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if r := run(); r.Status != StatusOK {
			t.Errorf("status = %q, want ok (%q)", r.Status, r.Problem)
		}
	})
}

func TestHookScriptsProbe(t *testing.T) {
	root := t.TempDir()
	claudeDir := filepath.Join(root, ".claude")
	scriptsDir := filepath.Join(root, "scripts")
	for _, d := range []string{claudeDir, scriptsDir} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeSettings := func(body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	settings := `{"hooks":{"PostToolUse":[{"hooks":[{"command":"bash scripts/sync-issues.sh"}]}]}}`

	t.Run("all referenced scripts exist", func(t *testing.T) {
		writeSettings(settings)
		if err := os.WriteFile(filepath.Join(scriptsDir, "sync-issues.sh"), []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}
		r := hookScriptsProbe(root)(context.Background(), nil)
		if r.Status != StatusOK {
			t.Errorf("status = %q, want ok (%q)", r.Status, r.Problem)
		}
	})

	t.Run("missing script named", func(t *testing.T) {
		writeSettings(settings)
		if err := os.Remove(filepath.Join(scriptsDir, "sync-issues.sh")); err != nil {
			t.Fatal(err)
		}
		r := hookScriptsProbe(root)(context.Background(), nil)
		if r.Status != StatusWarning {
			t.Fatalf("status = %q, want warning", r.Status)
		}
		if !strings.Contains(r.Problem, "scripts/sync-issues.sh") {
			t.Errorf("problem %q, want it to name the missing script", r.Problem)
		}
	})

	t.Run("DependsOn skip when settings parsing failed", func(t *testing.T) {
		// Simulate the real pipeline via Run so DependsOn skip actually triggers.
		writeSettings(`{"hooks": }`) // malformed
		checks := []Check{
			{Name: CheckSettingsJSON, Group: GroupProject, Required: true, Probe: settingsJSONProbe(root)},
			{Name: CheckHookScripts, Group: GroupProject, DependsOn: CheckSettingsJSON, Probe: hookScriptsProbe(root)},
		}
		results := Run(context.Background(), checks)
		if len(results) != 2 {
			t.Fatalf("len(results) = %d, want 2", len(results))
		}
		if results[1].Status != StatusSkipped {
			t.Errorf("hook-scripts status = %q, want skipped (DependsOn failed)", results[1].Status)
		}
	})
}

// stubExec runs the test's assertion over each exec call and chooses a shell
// command to simulate the git subprocess's response.
func stubExec(fn func(args []string) *exec.Cmd) func(context.Context, string, ...string) *exec.Cmd {
	return func(_ context.Context, name string, args ...string) *exec.Cmd {
		return fn(append([]string{name}, args...))
	}
}

func TestOriginGitHubProbe(t *testing.T) {
	cases := []struct {
		name   string
		url    string
		inGit  bool
		fails  bool
		want   Status
		substr string
	}{
		{"https remote ok", "https://github.com/me/x.git", true, false, StatusOK, ""},
		{"ssh shorthand ok", "git@github.com:me/x.git", true, false, StatusOK, ""},
		{"ssh full url ok", "ssh://git@github.com/me/x", true, false, StatusOK, ""},
		{"non-github warning", "https://gitlab.com/me/x.git", true, false, StatusWarning, "gitlab.com"},
		{"no remote warning", "", true, true, StatusWarning, "no origin"},
		{"no git skipped", "", false, false, StatusSkipped, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer ExecCommandForTest(stubExec(func(_ []string) *exec.Cmd {
				if tc.fails {
					return exec.Command("false")
				}
				return exec.Command("sh", "-c", "echo "+tc.url)
			}))()

			r := originGitHubProbe("/root", tc.inGit)(context.Background(), nil)
			if r.Status != tc.want {
				t.Fatalf("status = %q, want %q (%q)", r.Status, tc.want, r.Problem)
			}
			if tc.substr != "" && !strings.Contains(r.Problem, tc.substr) {
				t.Errorf("problem %q, want it to contain %q", r.Problem, tc.substr)
			}
		})
	}
}

func TestEnvNotTrackedProbe(t *testing.T) {
	t.Run("untracked ok", func(t *testing.T) {
		defer ExecCommandForTest(stubExec(func(_ []string) *exec.Cmd {
			return exec.Command("false") // ls-files --error-unmatch failed = untracked
		}))()
		if r := (envNotTrackedProbe("/root", true))(context.Background(), nil); r.Status != StatusOK {
			t.Errorf("status = %q, want ok (%q)", r.Status, r.Problem)
		}
	})
	t.Run("tracked is an error", func(t *testing.T) {
		defer ExecCommandForTest(stubExec(func(_ []string) *exec.Cmd {
			return exec.Command("true") // exit 0 = tracked
		}))()
		r := envNotTrackedProbe("/root", true)(context.Background(), nil)
		if r.Status != StatusError {
			t.Fatalf("status = %q, want error (%q)", r.Status, r.Problem)
		}
	})
	t.Run("no git skipped", func(t *testing.T) {
		if r := envNotTrackedProbe("/root", false)(context.Background(), nil); r.Status != StatusSkipped {
			t.Errorf("status = %q, want skipped", r.Status)
		}
	})
}

func TestEnvIgnoredProbe(t *testing.T) {
	t.Run("ignored ok", func(t *testing.T) {
		defer ExecCommandForTest(stubExec(func(_ []string) *exec.Cmd {
			return exec.Command("true") // check-ignore -q exit 0 = ignored
		}))()
		if r := envIgnoredProbe("/root", true)(context.Background(), nil); r.Status != StatusOK {
			t.Errorf("status = %q, want ok (%q)", r.Status, r.Problem)
		}
	})
	t.Run("not ignored warning", func(t *testing.T) {
		defer ExecCommandForTest(stubExec(func(_ []string) *exec.Cmd {
			return exec.Command("false")
		}))()
		r := envIgnoredProbe("/root", true)(context.Background(), nil)
		if r.Status != StatusWarning {
			t.Errorf("status = %q, want warning (%q)", r.Status, r.Problem)
		}
	})
	t.Run("no git skipped", func(t *testing.T) {
		if r := envIgnoredProbe("/root", false)(context.Background(), nil); r.Status != StatusSkipped {
			t.Errorf("status = %q, want skipped", r.Status)
		}
	})
	t.Run("DependsOn skip when .env is tracked", func(t *testing.T) {
		defer ExecCommandForTest(stubExec(func(_ []string) *exec.Cmd {
			// Both probes return "tracked" (exit 0) for ls-files; the second
			// never runs because DependsOn skips it.
			return exec.Command("true")
		}))()
		checks := []Check{
			{Name: CheckEnvNotTracked, Group: GroupProject, Required: true, Probe: envNotTrackedProbe("/root", true)},
			{Name: CheckEnvIgnored, Group: GroupProject, DependsOn: CheckEnvNotTracked, Probe: envIgnoredProbe("/root", true)},
		}
		results := Run(context.Background(), checks)
		if results[1].Status != StatusSkipped {
			t.Errorf("env-ignored status = %q, want skipped", results[1].Status)
		}
	})
}

func TestIssueSyncProbe(t *testing.T) {
	okCheck := func(name, rem string) Check { return Check{Name: name, Remediation: rem} }
	fullOk := Results{
		toolJQ:            {Status: StatusOK, Check: okCheck(toolJQ, "install jq")},
		toolGHAuth:        {Status: StatusOK, Check: okCheck(toolGHAuth, "gh auth login")},
		CheckHookScripts:  {Status: StatusOK, Check: okCheck(CheckHookScripts, "restore")},
		CheckOriginGitHub: {Status: StatusOK, Check: okCheck(CheckOriginGitHub, "make repo-init")},
	}

	t.Run("all prerequisites ok", func(t *testing.T) {
		r := issueSyncProbe()(context.Background(), fullOk)
		if r.Status != StatusOK {
			t.Errorf("status = %q, want ok (%q)", r.Status, r.Problem)
		}
	})

	t.Run("gh auth missing borrows its remediation", func(t *testing.T) {
		prior := Results{}
		for k, v := range fullOk {
			prior[k] = v
		}
		prior[toolGHAuth] = Result{Status: StatusWarning, Problem: "not authenticated", Check: okCheck(toolGHAuth, "gh auth login")}
		r := issueSyncProbe()(context.Background(), prior)
		if r.Status != StatusWarning {
			t.Fatalf("status = %q, want warning", r.Status)
		}
		if !strings.Contains(r.Problem, toolGHAuth) {
			t.Errorf("problem %q, want it to name %q", r.Problem, toolGHAuth)
		}
		if r.Check.Remediation != "gh auth login" {
			t.Errorf("remediation override = %q, want %q", r.Check.Remediation, "gh auth login")
		}
	})

	t.Run("names the earliest failure in order", func(t *testing.T) {
		prior := Results{}
		for k, v := range fullOk {
			prior[k] = v
		}
		prior[toolJQ] = Result{Status: StatusWarning, Problem: "missing", Check: okCheck(toolJQ, "install jq")}
		prior[CheckOriginGitHub] = Result{Status: StatusWarning, Problem: "no remote", Check: okCheck(CheckOriginGitHub, "make repo-init")}
		r := issueSyncProbe()(context.Background(), prior)
		if !strings.Contains(r.Problem, toolJQ) {
			t.Errorf("problem %q, want it to name jq first", r.Problem)
		}
	})

	t.Run("missing prerequisite entry is non-ok", func(t *testing.T) {
		// No prior results at all = every dep missing; expect warning on jq.
		r := issueSyncProbe()(context.Background(), Results{})
		if r.Status != StatusWarning {
			t.Fatalf("status = %q, want warning", r.Status)
		}
		if !strings.Contains(r.Problem, toolJQ) {
			t.Errorf("problem %q, want it to name jq", r.Problem)
		}
	})
}

func TestProjectChecks(t *testing.T) {
	checks := ProjectChecks("/root", true)
	wantNames := []string{
		CheckSettingsJSON,
		CheckIntendMD,
		CheckHookScripts,
		CheckOriginGitHub,
		CheckIssueSyncActive,
		CheckEnvNotTracked,
		CheckEnvIgnored,
	}
	if len(checks) != len(wantNames) {
		t.Fatalf("len(ProjectChecks) = %d, want %d", len(checks), len(wantNames))
	}
	for i, want := range wantNames {
		if checks[i].Name != want {
			t.Errorf("checks[%d].Name = %q, want %q", i, checks[i].Name, want)
		}
		if checks[i].Group != GroupProject {
			t.Errorf("checks[%d] (%s).Group = %q, want %q", i, checks[i].Name, checks[i].Group, GroupProject)
		}
	}

	// Required flags per spec.
	wantRequired := map[string]bool{
		CheckSettingsJSON:  true,
		CheckEnvNotTracked: true,
	}
	for _, c := range checks {
		if c.Required != wantRequired[c.Name] {
			t.Errorf("%s.Required = %v, want %v", c.Name, c.Required, wantRequired[c.Name])
		}
	}

	// DependsOn edges.
	wantDeps := map[string]string{
		CheckHookScripts: CheckSettingsJSON,
		CheckEnvIgnored:  CheckEnvNotTracked,
	}
	for _, c := range checks {
		if c.DependsOn != wantDeps[c.Name] {
			t.Errorf("%s.DependsOn = %q, want %q", c.Name, c.DependsOn, wantDeps[c.Name])
		}
	}
}
