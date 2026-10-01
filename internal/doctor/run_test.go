package doctor

import (
	"context"
	"os/exec"
	"testing"
)

func TestRunDependentSkip(t *testing.T) {
	orig := execCommand
	defer func() { execCommand = orig }()

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		switch name {
		case "gh":
			// gh is missing.
			return exec.CommandContext(ctx, "nonexistent-viber-tool-xyz", args...)
		default:
			return exec.CommandContext(ctx, "true")
		}
	}

	checks := []Check{
		{Name: "gh", Args: []string{"gh", "--version"}, Remediation: "install gh"},
		{Name: "gh-auth", Args: []string{"gh", "auth", "status"}, DependsOn: "gh", Remediation: "gh auth login"},
	}

	results := Run(context.Background(), checks)
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Status != StatusWarning {
		t.Errorf("gh status = %q, want warning", results[0].Status)
	}
	if results[1].Status != StatusSkipped {
		t.Errorf("gh-auth status = %q, want skipped (dependency failed)", results[1].Status)
	}
}

func TestRunCancelledContext(t *testing.T) {
	orig := execCommand
	defer func() { execCommand = orig }()

	execCommand = func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "true")
	}

	ctx, cancel := context.WithCancel(context.Background())

	checks := []Check{
		{Name: "git", Args: []string{"git", "--version"}, Required: true, Remediation: "install git"},
		{Name: "openspec", Args: []string{"openspec", "--version"}, Required: true, Remediation: "install openspec"},
		{Name: "jq", Args: []string{"jq", "--version"}, Remediation: "install jq"},
	}

	// Cancel before running; no checks should complete.
	cancel()
	results := Run(ctx, checks)
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0 after immediate cancellation", len(results))
	}
}
