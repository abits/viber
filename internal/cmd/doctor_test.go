package cmd

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/abits/viber/internal/doctor"
)

// chdirForTest switches the working directory to dir, restoring the previous
// one when the test returns. Doctor tests use it to point os.Getwd() at a
// scaffold the test owns, so DetectProject can see the sentinels.
func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

// mkScaffoldForDoctor creates a temp directory with the three scaffold
// sentinels viber doctor detects (CLAUDE.md, openspec/, .claude/) plus a
// valid settings.json so the project checks actually run instead of only
// skipping on dependency failures.
func mkScaffoldForDoctor(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("# scaffold\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{"openspec", ".claude"} {
		if err := os.Mkdir(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".claude", "settings.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// scaffoldExec returns a fake executor that routes every git subprocess the
// doctor pipeline issues against scaffoldDir, so DetectProject sees the dir
// and every project probe gets a plausible response. All tool probes return
// a stub version string.
func scaffoldExec(scaffoldDir string) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if name == "git" {
			switch {
			case slices.Contains(args, "rev-parse"):
				return exec.CommandContext(ctx, "sh", "-c", "echo "+scaffoldDir)
			case slices.Contains(args, "remote"):
				return exec.CommandContext(ctx, "sh", "-c", "echo https://github.com/me/x.git")
			case slices.Contains(args, "ls-files"):
				// Non-zero exit = .env is not tracked.
				return exec.CommandContext(ctx, "false")
			case slices.Contains(args, "check-ignore"):
				// Exit 0 = .env is ignored.
				return exec.CommandContext(ctx, "true")
			case slices.Contains(args, "--version"):
				return exec.CommandContext(ctx, "sh", "-c", "echo 2.43.0")
			}
		}
		return exec.CommandContext(ctx, "sh", "-c", "echo 1.0.0")
	}
}

// fakeExecAll returns a fake executor that simulates all tools as installed
// and reporting a fixed version string.
func fakeExecAll(version string) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "echo "+version)
	}
}

// fakeExecRequiredMissing returns a fake executor where required tools (git,
// openspec) are missing, and all others succeed.
func fakeExecRequiredMissing() func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if name == "git" || name == "openspec" {
			return exec.CommandContext(ctx, "nonexistent-viber-tool-xyz", args...)
		}
		return exec.CommandContext(ctx, "sh", "-c", "echo 1.0.0")
	}
}

// fakeExecOptionalsMissing returns a fake executor where optional tools are
// missing, but required ones succeed.
func fakeExecOptionalsMissing() func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if name == "git" || name == "openspec" {
			return exec.CommandContext(ctx, "sh", "-c", "echo 2.0.0")
		}
		return exec.CommandContext(ctx, "nonexistent-viber-tool-xyz", args...)
	}
}

func TestDoctorExtraArgExits2(t *testing.T) {
	root, _, _ := newTestRoot(t, "doctor", "extra-arg")
	if got := execute(context.Background(), root); got != exitUsage {
		t.Errorf("exit = %d, want %d (usage error)", got, exitUsage)
	}
}

func TestDoctorWarningsOnlyExits0(t *testing.T) {
	restore := doctor.ExecCommandForTest(fakeExecOptionalsMissing())
	defer restore()

	root, out, _ := newTestRoot(t, "doctor")
	if got := execute(context.Background(), root); got != exitOK {
		t.Errorf("exit = %d, want %d (warnings only)", got, exitOK)
	}

	outStr := out.String()
	if !strings.Contains(outStr, "0 errors") {
		t.Errorf("summary missing '0 errors':\n%s", outStr)
	}
	// At least some warnings from missing optional tools.
	if !strings.Contains(outStr, "warning") {
		t.Errorf("expected some warnings in output:\n%s", outStr)
	}
}

func TestDoctorRequiredFailureExits1(t *testing.T) {
	restore := doctor.ExecCommandForTest(fakeExecRequiredMissing())
	defer restore()

	root, _, _ := newTestRoot(t, "doctor")
	if got := execute(context.Background(), root); got != exitFailure {
		t.Errorf("exit = %d, want %d (required failure)", got, exitFailure)
	}
}

func TestDoctorSummaryCountsMatchResults(t *testing.T) {
	restore := doctor.ExecCommandForTest(fakeExecRequiredMissing())
	defer restore()

	root, out, _ := newTestRoot(t, "doctor")
	execute(context.Background(), root)

	outStr := out.String()
	// git and openspec are required and missing => 2 errors.
	if !strings.Contains(outStr, "2 errors") {
		t.Errorf("expected '2 errors' in summary:\n%s", outStr)
	}
}

func TestDoctorJSONValidOutput(t *testing.T) {
	restore := doctor.ExecCommandForTest(fakeExecAll("1.2.3"))
	defer restore()

	root, out, _ := newTestRoot(t, "doctor", "--json")
	execute(context.Background(), root)

	var results []jsonResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &results); err != nil {
		t.Fatalf("--json output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(results) == 0 {
		t.Error("--json output is an empty array")
	}
}

func TestDoctorJSONMissingJqHasWarning(t *testing.T) {
	restore := doctor.ExecCommandForTest(fakeExecOptionalsMissing())
	defer restore()

	root, out, _ := newTestRoot(t, "doctor", "--json")
	execute(context.Background(), root)

	var results []jsonResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &results); err != nil {
		t.Fatalf("--json output is not valid JSON: %v", err)
	}

	var jqEntry *jsonResult
	for i := range results {
		if results[i].Name == "jq" {
			jqEntry = &results[i]
			break
		}
	}
	if jqEntry == nil {
		t.Fatal("no jq entry in JSON output")
	}
	if jqEntry.Status != "warning" {
		t.Errorf("jq status = %q, want %q", jqEntry.Status, "warning")
	}
	if jqEntry.Remediation == "" {
		t.Error("jq remediation should be non-empty")
	}
}

func TestDoctorJSONExitCodesMatchHuman(t *testing.T) {
	cases := []struct {
		name    string
		exec    func(context.Context, string, ...string) *exec.Cmd
		wantExt int
	}{
		{
			name:    "all ok",
			exec:    fakeExecAll("1.0.0"),
			wantExt: exitOK,
		},
		{
			name:    "required missing",
			exec:    fakeExecRequiredMissing(),
			wantExt: exitFailure,
		},
		{
			name:    "optional missing only",
			exec:    fakeExecOptionalsMissing(),
			wantExt: exitOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			restoreH := doctor.ExecCommandForTest(tc.exec)
			rootH, _, _ := newTestRoot(t, "doctor")
			humanExit := execute(context.Background(), rootH)
			restoreH()

			restoreJ := doctor.ExecCommandForTest(tc.exec)
			rootJ, _, _ := newTestRoot(t, "doctor", "--json")
			jsonExit := execute(context.Background(), rootJ)
			restoreJ()

			if humanExit != tc.wantExt {
				t.Errorf("human exit = %d, want %d", humanExit, tc.wantExt)
			}
			if jsonExit != tc.wantExt {
				t.Errorf("json exit = %d, want %d", jsonExit, tc.wantExt)
			}
			if humanExit != jsonExit {
				t.Errorf("human exit %d != json exit %d", humanExit, jsonExit)
			}
		})
	}
}

func TestDoctorScaffoldAppendsProjectChecks(t *testing.T) {
	scaffold := mkScaffoldForDoctor(t)
	chdirForTest(t, scaffold)

	restore := doctor.ExecCommandForTest(scaffoldExec(scaffold))
	defer restore()

	root, out, _ := newTestRoot(t, "doctor")
	if exit := execute(context.Background(), root); exit != exitOK {
		t.Errorf("exit = %d, want %d (healthy scaffold)", exit, exitOK)
	}

	outStr := out.String()
	for _, name := range []string{
		doctor.CheckSettingsJSON,
		doctor.CheckIntendMD,
		doctor.CheckHookScripts,
		doctor.CheckOriginGitHub,
		doctor.CheckIssueSyncActive,
		doctor.CheckEnvNotTracked,
		doctor.CheckEnvIgnored,
	} {
		if !strings.Contains(outStr, name) {
			t.Errorf("output missing %q:\n%s", name, outStr)
		}
	}
	if !strings.Contains(outStr, "Project "+scaffold) {
		t.Errorf("output missing 'Project <root>' heading:\n%s", outStr)
	}
}

func TestDoctorNonScaffoldHasOnlyToolChecks(t *testing.T) {
	nonScaffold := t.TempDir() // No sentinels.
	chdirForTest(t, nonScaffold)

	restore := doctor.ExecCommandForTest(fakeExecAll("1.0.0"))
	defer restore()

	root, out, _ := newTestRoot(t, "doctor")
	if exit := execute(context.Background(), root); exit != exitOK {
		t.Errorf("exit = %d, want %d", exit, exitOK)
	}

	outStr := out.String()
	for _, name := range []string{
		doctor.CheckSettingsJSON,
		doctor.CheckIntendMD,
		doctor.CheckHookScripts,
		doctor.CheckOriginGitHub,
		doctor.CheckIssueSyncActive,
		doctor.CheckEnvNotTracked,
		doctor.CheckEnvIgnored,
	} {
		if strings.Contains(outStr, name) {
			t.Errorf("non-scaffold output should not include %q:\n%s", name, outStr)
		}
	}
	if strings.Contains(outStr, "Project ") {
		t.Errorf("non-scaffold output should not include Project heading:\n%s", outStr)
	}
}

// writeIntend augments the scaffold with intend.md so the scaffold-golden
// test captures the all-ok path, not an incidental warning.
func writeIntend(t *testing.T, root string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "intend.md"), []byte("# intend\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// assertGolden compares got to the file at path. Set UPDATE_GOLDEN=1 to
// (re)write the file instead of comparing. The scaffold placeholder lets a
// test substitute per-run temp paths out of the stored file.
func assertGolden(t *testing.T, path, got, scaffold string) {
	t.Helper()
	if os.Getenv("UPDATE_GOLDEN") != "" {
		body := got
		if scaffold != "" {
			body = strings.ReplaceAll(body, scaffold, "{{SCAFFOLD}}")
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (hint: run with UPDATE_GOLDEN=1 to create it)", path, err)
	}
	want := string(raw)
	if scaffold != "" {
		want = strings.ReplaceAll(want, "{{SCAFFOLD}}", scaffold)
	}
	if got != want {
		t.Errorf("doctor output does not match %s:\n--- got ---\n%s--- want ---\n%s", path, got, want)
	}
}

// goldenPath resolves name against the package's testdata directory using an
// absolute path, so tests that chdir elsewhere still find the golden file.
func goldenPath(t *testing.T, name string) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func TestDoctorToolOnlyGolden(t *testing.T) {
	// Non-scaffold cwd keeps project checks out of the output; this is the
	// pre-change baseline the spec requires to stay byte-identical.
	gp := goldenPath(t, "doctor-tools-only.txt")
	chdirForTest(t, t.TempDir())
	restore := doctor.ExecCommandForTest(fakeExecAll("2.43.0"))
	defer restore()

	root, out, _ := newTestRoot(t, "doctor")
	execute(context.Background(), root)

	assertGolden(t, gp, out.String(), "")
}

func TestDoctorScaffoldGolden(t *testing.T) {
	gp := goldenPath(t, "doctor-scaffold.txt")
	scaffold := mkScaffoldForDoctor(t)
	writeIntend(t, scaffold)
	chdirForTest(t, scaffold)

	restore := doctor.ExecCommandForTest(scaffoldExec(scaffold))
	defer restore()

	root, out, _ := newTestRoot(t, "doctor")
	execute(context.Background(), root)

	assertGolden(t, gp, out.String(), scaffold)
}

func TestDoctorJSONGroupField(t *testing.T) {
	scaffold := mkScaffoldForDoctor(t)
	writeIntend(t, scaffold)
	chdirForTest(t, scaffold)

	restore := doctor.ExecCommandForTest(scaffoldExec(scaffold))
	defer restore()

	root, out, _ := newTestRoot(t, "doctor", "--json")
	execute(context.Background(), root)

	var results []jsonResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &results); err != nil {
		t.Fatalf("--json output is not a flat JSON array: %v\noutput: %s", err, out.String())
	}

	var toolSeen, projectSeen bool
	for _, r := range results {
		switch r.Group {
		case doctor.GroupTools:
			toolSeen = true
		case doctor.GroupProject:
			projectSeen = true
		default:
			t.Errorf("entry %q has unexpected group %q", r.Name, r.Group)
		}
	}
	if !toolSeen || !projectSeen {
		t.Errorf("expected both groups in output, got tool=%v project=%v", toolSeen, projectSeen)
	}
}

func TestDoctorScaffoldExitCodes(t *testing.T) {
	// Each case tailors a scaffold + exec stub for one of the four scenarios
	// the spec contract exercises, then runs doctor through Execute.
	cases := []struct {
		name     string
		mutate   func(t *testing.T, scaffold string)
		execFn   func(scaffold string) func(context.Context, string, ...string) *exec.Cmd
		wantExit int
	}{
		{
			name:     "healthy scaffold exits 0",
			mutate:   func(t *testing.T, s string) { writeIntend(t, s) },
			execFn:   scaffoldExec,
			wantExit: exitOK,
		},
		{
			name: "malformed settings.json exits 1",
			mutate: func(t *testing.T, s string) {
				writeIntend(t, s)
				// Overwrite the valid "{}" with a parse error.
				if err := os.WriteFile(filepath.Join(s, ".claude", "settings.json"), []byte(`{"hooks": }`), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			execFn:   scaffoldExec,
			wantExit: exitFailure,
		},
		{
			name:   "tracked .env exits 1",
			mutate: func(t *testing.T, s string) { writeIntend(t, s) },
			execFn: func(scaffold string) func(context.Context, string, ...string) *exec.Cmd {
				base := scaffoldExec(scaffold)
				return func(ctx context.Context, name string, args ...string) *exec.Cmd {
					if name == "git" && slices.Contains(args, "ls-files") {
						// Exit 0 = .env is tracked.
						return exec.CommandContext(ctx, "true")
					}
					return base(ctx, name, args...)
				}
			},
			wantExit: exitFailure,
		},
		{
			name:   ".env not ignored is warning only, exits 0",
			mutate: func(t *testing.T, s string) { writeIntend(t, s) },
			execFn: func(scaffold string) func(context.Context, string, ...string) *exec.Cmd {
				base := scaffoldExec(scaffold)
				return func(ctx context.Context, name string, args ...string) *exec.Cmd {
					if name == "git" && slices.Contains(args, "check-ignore") {
						return exec.CommandContext(ctx, "false") // not ignored
					}
					return base(ctx, name, args...)
				}
			},
			wantExit: exitOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scaffold := mkScaffoldForDoctor(t)
			tc.mutate(t, scaffold)
			chdirForTest(t, scaffold)
			restore := doctor.ExecCommandForTest(tc.execFn(scaffold))
			defer restore()

			root, _, _ := newTestRoot(t, "doctor")
			if got := execute(context.Background(), root); got != tc.wantExit {
				t.Errorf("exit = %d, want %d", got, tc.wantExit)
			}
		})
	}
}

func TestDoctorInterruptedExitsNonZero(t *testing.T) {
	restore := doctor.ExecCommandForTest(fakeExecAll("1.0.0"))
	defer restore()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	root, _, errb := newTestRoot(t, "doctor")
	got := execute(ctx, root)

	if got != exitFailure {
		t.Errorf("exit = %d after interrupt, want %d", got, exitFailure)
	}
	if !strings.Contains(errb.String(), "interrupted") {
		t.Errorf("stderr = %q, want it to mention the interruption", errb.String())
	}
}
