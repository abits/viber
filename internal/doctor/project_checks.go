package doctor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Project check names are exported so the renderer, the issue-sync probe, and
// tests can reference them without string drift.
const (
	CheckSettingsJSON    = ".claude/settings.json valid"
	CheckIntendMD        = "intend.md present"
	CheckHookScripts     = "hook scripts present"
	CheckOriginGitHub    = "origin on github.com"
	CheckIssueSyncActive = "issue sync active"
	CheckEnvNotTracked   = ".env not tracked"
	CheckEnvIgnored      = ".env ignored"
)

// Tool check names referenced by the issue-sync aggregator. Kept here so a
// rename in checks.go surfaces as a build break rather than a silent miss.
const (
	toolJQ     = "jq"
	toolGHAuth = "gh auth"
)

// hookScriptRe matches script references inside hook command strings. The
// regex is deliberately scoped to the scripts/ prefix viber scaffolds use;
// it treats JSON escape sequences as plain text, which is fine because every
// legal script path survives JSON escaping unchanged.
var hookScriptRe = regexp.MustCompile(`scripts/[A-Za-z0-9._/-]+`)

// ProjectChecks returns the ordered project-level checks that run inside a
// scaffold, with root resolved to the project root. inGit controls whether
// git-dependent checks (origin, .env tracked, .env ignored) actually probe or
// report StatusSkipped.
func ProjectChecks(root string, inGit bool) []Check {
	return []Check{
		{
			Name:        CheckSettingsJSON,
			Group:       GroupProject,
			Required:    true,
			Probe:       settingsJSONProbe(root),
			Remediation: "repair .claude/settings.json so it parses as JSON",
		},
		{
			Name:        CheckIntendMD,
			Group:       GroupProject,
			Probe:       intendMDProbe(root),
			Remediation: "create intend.md and describe what you want to build",
		},
		{
			Name:        CheckHookScripts,
			Group:       GroupProject,
			DependsOn:   CheckSettingsJSON,
			Probe:       hookScriptsProbe(root),
			Remediation: "restore the missing script, or remove its hook from .claude/settings.json",
		},
		{
			Name:        CheckOriginGitHub,
			Group:       GroupProject,
			Probe:       originGitHubProbe(root, inGit),
			Remediation: "make repo-init",
		},
		{
			Name:  CheckIssueSyncActive,
			Group: GroupProject,
			Probe: issueSyncProbe(),
			// Remediation inherits the first failed prerequisite's fix at probe
			// time; this fallback is only used when every prerequisite passed
			// yet the probe still reported a non-ok status.
			Remediation: "inspect the issue-sync hook in .claude/settings.json",
		},
		{
			Name:        CheckEnvNotTracked,
			Group:       GroupProject,
			Required:    true,
			Probe:       envNotTrackedProbe(root, inGit),
			Remediation: "git rm --cached .env",
		},
		{
			Name:        CheckEnvIgnored,
			Group:       GroupProject,
			DependsOn:   CheckEnvNotTracked,
			Probe:       envIgnoredProbe(root, inGit),
			Remediation: "add .env to .gitignore",
		},
	}
}

// settingsJSONProbe reads root/.claude/settings.json and asserts it parses
// as JSON. A missing file is an error because hooks cannot run without it.
func settingsJSONProbe(root string) func(context.Context, Results) Result {
	return func(_ context.Context, _ Results) Result {
		path := filepath.Join(root, ".claude", "settings.json")
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return Result{Status: StatusError, Problem: ".claude/settings.json is missing"}
			}
			return Result{Status: StatusError, Problem: fmt.Sprintf("read .claude/settings.json: %v", err)}
		}
		var raw any
		if err := json.Unmarshal(data, &raw); err != nil {
			var syntaxErr *json.SyntaxError
			if errors.As(err, &syntaxErr) {
				return Result{Status: StatusError, Problem: fmt.Sprintf("invalid JSON at byte %d: %v", syntaxErr.Offset, err)}
			}
			return Result{Status: StatusError, Problem: fmt.Sprintf("invalid JSON: %v", err)}
		}
		return Result{Status: StatusOK}
	}
}

// intendMDProbe only checks that intend.md exists; its contents are the
// user's to decide and are explicitly out of scope.
func intendMDProbe(root string) func(context.Context, Results) Result {
	return func(_ context.Context, _ Results) Result {
		if _, err := os.Stat(filepath.Join(root, "intend.md")); err != nil {
			return Result{Status: StatusWarning, Problem: "intend.md is missing"}
		}
		return Result{Status: StatusOK}
	}
}

// hookScriptsProbe re-reads settings.json and scans every string in it for
// `scripts/<path>` references, then checks that each referenced file exists.
// A regex over the raw bytes is intentional: it covers both current and
// future hook schemas without parsing shell.
func hookScriptsProbe(root string) func(context.Context, Results) Result {
	return func(_ context.Context, _ Results) Result {
		path := filepath.Join(root, ".claude", "settings.json")
		data, err := os.ReadFile(path)
		if err != nil {
			return Result{Status: StatusWarning, Problem: fmt.Sprintf("read .claude/settings.json: %v", err)}
		}
		refs := hookScriptRe.FindAllString(string(data), -1)
		// Deduplicate without reordering so the first missing match is stable.
		seen := make(map[string]bool, len(refs))
		for _, ref := range refs {
			if seen[ref] {
				continue
			}
			seen[ref] = true
			// Reject references that could escape the project root. The regex
			// matches dots, which can combine into "..", so paths like
			// scripts/../../../etc/passwd are syntactically allowed; refuse
			// them before passing to os.Stat.
			if strings.Contains(ref, "..") {
				return Result{Status: StatusWarning, Problem: "suspicious hook script path (contains ..): " + ref}
			}
			// gosec G703 flags the user-controlled path, but the preceding
			// guard rejects ".." traversal and the regex restricts ref to
			// a `scripts/` prefix with alphanumerics/dots/dashes/underscores
			// only. os.Stat is read-only, so this cannot touch anything.
			if _, err := os.Stat(filepath.Join(root, ref)); err != nil { //nolint:gosec // G703: ref sanitized above.
				return Result{Status: StatusWarning, Problem: "missing hook script: " + ref}
			}
		}
		return Result{Status: StatusOK}
	}
}

// originGitHubProbe shells out to `git remote get-url origin` and reports
// whether the result points at github.com. The three accepted URL forms
// cover HTTPS, SSH shorthand, and the full SSH URL.
func originGitHubProbe(root string, inGit bool) func(context.Context, Results) Result {
	return func(ctx context.Context, _ Results) Result {
		if !inGit {
			return Result{Status: StatusSkipped}
		}
		cmd := execCommand(ctx, "git", "-C", root, "remote", "get-url", "origin")
		cmd.WaitDelay = probeWaitDelay
		out, err := cmd.Output()
		if err != nil {
			return Result{Status: StatusWarning, Problem: "no origin remote"}
		}
		url := strings.TrimSpace(string(out))
		if isGitHubURL(url) {
			return Result{Status: StatusOK}
		}
		return Result{Status: StatusWarning, Problem: "origin is not on github.com: " + url}
	}
}

func isGitHubURL(url string) bool {
	return strings.HasPrefix(url, "https://github.com/") ||
		strings.HasPrefix(url, "git@github.com:") ||
		strings.HasPrefix(url, "ssh://git@github.com/")
}

// envNotTrackedProbe runs `git ls-files --error-unmatch .env`: exit 0 means
// the file is tracked (which leaks the secret into history), any error means
// it is not. We treat tracked as a required error.
func envNotTrackedProbe(root string, inGit bool) func(context.Context, Results) Result {
	return func(ctx context.Context, _ Results) Result {
		if !inGit {
			return Result{Status: StatusSkipped}
		}
		cmd := execCommand(ctx, "git", "-C", root, "ls-files", "--error-unmatch", ".env")
		cmd.WaitDelay = probeWaitDelay
		if err := cmd.Run(); err == nil {
			return Result{Status: StatusError, Problem: ".env is tracked by git"}
		}
		return Result{Status: StatusOK}
	}
}

// envIgnoredProbe runs `git check-ignore -q .env`: exit 0 means .env is
// matched by a gitignore rule, non-zero means it is not. DependsOn=.env not
// tracked skips this check when .env is tracked, so the warning never
// repeats what the preceding error already said.
func envIgnoredProbe(root string, inGit bool) func(context.Context, Results) Result {
	return func(ctx context.Context, _ Results) Result {
		if !inGit {
			return Result{Status: StatusSkipped}
		}
		cmd := execCommand(ctx, "git", "-C", root, "check-ignore", "-q", ".env")
		cmd.WaitDelay = probeWaitDelay
		if err := cmd.Run(); err != nil {
			return Result{Status: StatusWarning, Problem: ".env is not matched by any gitignore rule"}
		}
		return Result{Status: StatusOK}
	}
}

// issueSyncProbe aggregates the outcome of its four prerequisites. On the
// first non-ok result it reports a warning that names the failed prerequisite
// and borrows its remediation (via the Check.Remediation override wired in
// probe()). A prerequisite that was skipped or never ran counts as non-ok.
func issueSyncProbe() func(context.Context, Results) Result {
	deps := []string{toolJQ, toolGHAuth, CheckHookScripts, CheckOriginGitHub}
	return func(_ context.Context, prior Results) Result {
		for _, name := range deps {
			r, ok := prior[name]
			if !ok || r.Status != StatusOK {
				problem := fmt.Sprintf("inactive: %s did not pass", name)
				if ok && r.Problem != "" {
					problem = fmt.Sprintf("inactive: %s - %s", name, r.Problem)
				}
				override := ""
				if ok {
					override = r.Check.Remediation
				}
				return Result{
					Status:  StatusWarning,
					Problem: problem,
					Check:   Check{Remediation: override},
				}
			}
		}
		return Result{Status: StatusOK}
	}
}
