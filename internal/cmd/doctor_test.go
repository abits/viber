package cmd

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/abits/viber/internal/doctor"
)

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
