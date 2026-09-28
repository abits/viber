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

// ErrRequiredFailed is returned by Run when at least one required check failed.
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

// Check describes one external-tool probe.
type Check struct {
	// Name is the human-readable label shown in output.
	Name string
	// Args is the command and arguments to run (e.g. ["git", "--version"]).
	Args []string
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

// probe runs c and returns its Result. The caller supplies the outer ctx;
// probe derives a per-check timeout from it.
func probe(ctx context.Context, c Check) Result {
	pctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := execCommand(pctx, c.Args[0], c.Args[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

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
func Run(ctx context.Context, checks []Check) []Result {
	results := make([]Result, 0, len(checks))
	// index by Name for dependency lookups.
	outcome := make(map[string]Status, len(checks))

	for _, c := range checks {
		if ctx.Err() != nil {
			break
		}
		if c.DependsOn != "" {
			if s, ok := outcome[c.DependsOn]; ok && s != StatusOK {
				r := Result{Check: c, Status: StatusSkipped}
				results = append(results, r)
				outcome[c.Name] = StatusSkipped
				continue
			}
		}
		r := probe(ctx, c)
		results = append(results, r)
		outcome[c.Name] = r.Status
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
