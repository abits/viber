// Package doctor probes external tools that viber and its scaffolds depend on.
package doctor

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// ErrRequiredFailed signals that at least one required check failed. Callers
// rendering []Result return it so the command exits 1.
var ErrRequiredFailed = errors.New("required check failed")

// Status describes the outcome of a single check.
type Status string

// Status constants mirror the spec's four outcomes.
const (
	StatusOK      Status = "ok"
	StatusWarning Status = "warning"
	StatusError   Status = "error"
	StatusSkipped Status = "skipped"
)

// Group tags which rendering group a Check belongs to. Tool checks probe
// external binaries viber or its scaffolds depend on; project checks inspect
// the scaffold's own files and git state and run only inside a scaffolded
// project (see DetectProject).
const (
	GroupTools   = "tools"
	GroupProject = "project"
)

// Results is a lookup of recorded Result values keyed by Check.Name. It is
// passed to a Probe so a probe-based check can aggregate the outcome of
// earlier checks (for example, the issue-sync probe needs the status of jq,
// gh auth, the hook-script check, and the remote check). Probes must treat
// the map as read-only.
type Results map[string]Result

// Check describes one check viber doctor runs.
type Check struct {
	// Name is the human-readable label shown in output.
	Name string
	// Group is "tools" or "project" and controls which rendering group this
	// check belongs to. See GroupTools / GroupProject.
	Group string
	// Args is the command and arguments to run (e.g. ["git", "--version"]).
	// Ignored when Probe is non-nil.
	Args []string
	// Probe runs the check without shelling out to a subprocess. When Probe
	// is non-nil, probe() delegates to it and Args is unused. Probes receive
	// a per-check timeout context and a read-only snapshot of prior Results.
	Probe func(ctx context.Context, prior Results) Result
	// Required marks the check as mandatory; a failure becomes StatusError and
	// causes ErrRequiredFailed.
	Required bool
	// DependsOn is the Name of another Check that must have passed for this one
	// to run. When the dependency did not pass, this check is StatusSkipped.
	DependsOn string
	// Remediation is the concrete fix shown when the check does not pass.
	Remediation string
}

// Result is the outcome of running a Check.
type Result struct {
	Check   Check
	Status  Status
	Version string // populated when the tool reported a parseable version
	Problem string // populated when the check did not pass
}

var execCommand = exec.CommandContext

var versionRe = regexp.MustCompile(`\d+\.\d+(\.\d+)?`)

var probeTimeout = 5 * time.Second

// probeWaitDelay bounds how long probe waits for a tool's output pipes after
// the timeout kills it. Without it, a child process that inherited stdout
// (node shims, .cmd wrappers) keeps cmd.Run blocked past the timeout.
const probeWaitDelay = 500 * time.Millisecond

// probe runs c and returns its Result. The caller supplies the outer ctx and
// a read-only snapshot of earlier results; probe derives a per-check timeout
// from ctx and either calls c.Probe (when non-nil) or shells out to c.Args.
func probe(ctx context.Context, c Check, prior Results) Result {
	pctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	if c.Probe != nil {
		r := c.Probe(pctx, prior)
		// Record which Check produced the Result so callers (renderer, tests)
		// have a single source for Name/Group/Required. A probe may override
		// Remediation (the issue-sync probe borrows its failed prerequisite's
		// fix); any non-empty value set on the returned Check wins.
		override := r.Check.Remediation
		r.Check = c
		if override != "" {
			r.Check.Remediation = override
		}
		return r
	}

	var stdout, stderr bytes.Buffer
	cmd := execCommand(pctx, c.Args[0], c.Args[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.WaitDelay = probeWaitDelay

	err := cmd.Run()
	if err == nil {
		return Result{
			Check:   c,
			Status:  StatusOK,
			Version: extractVersion(stdout.String()),
		}
	}

	// Missing binary.
	var pathErr *exec.Error
	if errors.As(err, &pathErr) {
		st := StatusWarning
		if c.Required {
			st = StatusError
		}
		return Result{Check: c, Status: st, Problem: "not installed"}
	}

	// Timed out.
	if ctx.Err() != nil || pctx.Err() != nil {
		st := StatusWarning
		if c.Required {
			st = StatusError
		}
		return Result{Check: c, Status: st, Problem: "timed out"}
	}

	// Non-zero exit (e.g. auth check failed).
	firstLine := firstNonEmpty(stderr.String(), stdout.String())
	if len(firstLine) > 120 {
		firstLine = firstLine[:120]
	}
	st := StatusWarning
	if c.Required {
		st = StatusError
	}
	return Result{Check: c, Status: st, Problem: firstLine}
}

// Run probes each check in order, honoring DependsOn and ctx cancellation.
// Probe-based checks receive the Results recorded so far, keyed by Name.
func Run(ctx context.Context, checks []Check) []Result {
	results := make([]Result, 0, len(checks))
	// prior keyed by Name serves both DependsOn skip logic and Probe lookups.
	prior := make(Results, len(checks))

	for _, c := range checks {
		if ctx.Err() != nil {
			break
		}
		if c.DependsOn != "" {
			if dep, ok := prior[c.DependsOn]; ok && dep.Status != StatusOK {
				r := Result{Check: c, Status: StatusSkipped}
				results = append(results, r)
				prior[c.Name] = r
				continue
			}
		}
		r := probe(ctx, c, prior)
		results = append(results, r)
		prior[c.Name] = r
	}
	return results
}

func extractVersion(s string) string {
	if m := versionRe.FindString(s); m != "" {
		return m
	}
	line := strings.SplitN(strings.TrimSpace(s), "\n", 2)[0]
	if len(line) > 60 {
		line = line[:60]
	}
	return line
}

func firstNonEmpty(a, b string) string {
	line := strings.SplitN(strings.TrimSpace(a), "\n", 2)[0]
	if line != "" {
		return line
	}
	return strings.SplitN(strings.TrimSpace(b), "\n", 2)[0]
}
