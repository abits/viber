package doctor

import (
	"context"
	"os/exec"
	"time"
)

// ExecCommandForTest replaces the package-level executor for the duration of a
// test. It returns a cleanup function that restores the original. Intended for
// use by external test packages that cannot access the unexported execCommand.
func ExecCommandForTest(fn func(context.Context, string, ...string) *exec.Cmd) func() {
	orig := execCommand
	execCommand = fn
	return func() { execCommand = orig }
}

// ProbeTimeoutForTest replaces probeTimeout for the duration of a test.
// It returns a cleanup function that restores the original.
func ProbeTimeoutForTest(d time.Duration) func() {
	orig := probeTimeout
	probeTimeout = d
	return func() { probeTimeout = orig }
}
