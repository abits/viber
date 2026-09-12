// Package cmd wires viber's Cobra command tree and owns the process-wide
// error reporting and exit-code contract:
//
//	0  success
//	1  runtime error (I/O, network, subprocess)
//	2  usage error (bad flags or arguments)
//
// Commands never print errors themselves; they return them, and [Execute]
// renders them exactly once.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/abits/viber/internal/version"
)

// Process exit codes. See the package comment for the contract they implement.
const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

// usageErr marks an error as caused by bad input rather than a runtime
// failure. It carries the command whose usage should be shown alongside it.
type usageErr struct {
	err error
	cmd *cobra.Command
}

func (u usageErr) Error() string { return u.err.Error() }
func (u usageErr) Unwrap() error { return u.err }

// UsageError marks err as a usage error, so that it exits with code 2 and is
// reported together with cmd's usage text. A nil cmd falls back to the root
// command.
func UsageError(cmd *cobra.Command, err error) error {
	return usageErr{err: err, cmd: cmd}
}

// usageArgs adapts a Cobra positional-argument validator so that a wrong
// argument count is reported as a usage error (exit 2, with usage shown)
// rather than as a runtime failure.
func usageArgs(fn cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := fn(cmd, args); err != nil {
			return UsageError(cmd, err)
		}
		return nil
	}
}

// exitCode maps an error returned by the command tree onto a process exit code.
func exitCode(err error) int {
	switch {
	case err == nil:
		return exitOK
	case errors.As(err, new(usageErr)):
		return exitUsage
	default:
		return exitFailure
	}
}

// Execute builds the command tree, runs it against a context cancelled on
// SIGINT, and returns the process exit code. It never calls os.Exit, so that
// callers (and tests) stay in control.
func Execute(v, c, d string) int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	return execute(ctx, newRootCmd(version.Info{Version: v, Commit: c, Date: d}))
}

// execute runs an already-built root command. It is the seam the command tests
// use to capture output, which Execute cannot offer because it owns its root.
func execute(ctx context.Context, root *cobra.Command) int {
	err := root.ExecuteContext(ctx)
	if err == nil {
		return exitOK
	}
	// Single point of truth for user-facing error output: SilenceErrors keeps
	// Cobra quiet, the runners only report step status, and everything that
	// reaches here is printed exactly once.
	w := root.ErrOrStderr()
	fmt.Fprintln(w, "Error:", err)
	var u usageErr
	if errors.As(err, &u) {
		cmd := u.cmd
		if cmd == nil {
			cmd = root
		}
		fmt.Fprint(w, cmd.UsageString())
	}
	return exitCode(err)
}

func newRootCmd(info version.Info) *cobra.Command {
	root := &cobra.Command{
		Use:   "viber",
		Short: "Scaffold a vibe-coding project wired for Claude Code + OpenSpec.",
		Long: `NAME
    viber - scaffold a vibe-coding project wired for Claude Code + OpenSpec

SYNOPSIS
    viber [command] [flags]

DESCRIPTION
    viber creates a new project with an OpenSpec-ready layout, a stakeholder
    intend.md, and a Claude Code sub-agent team covering the software
    development lifecycle.

    Run 'viber init --help' for the primary subcommand.

EXAMPLES
    Interactive scaffold:
        viber init

    Non-interactive scaffold (CI-friendly):
        viber init myproj --no-tui --name=myproj --desc="a side project"

    Use a remote template set:
        viber init myproj --from me/viber-templates@main

ENVIRONMENT
    GITHUB_TOKEN
        Optional. Sent as a bearer token when --from downloads a private
        or heavily rate-limited GitHub tarball.

EXIT STATUS
    0    success
    1    runtime error (I/O, network, subprocess)
    2    usage error (bad flags or arguments)

SEE ALSO
    OpenSpec:      https://github.com/Fission-AI/OpenSpec
    Claude Code:   https://claude.com/claude-code
`,
		Version:       info.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("{{.Version}}\n")
	root.Flags().BoolP("version", "V", false, "print version and exit")
	// Flag errors are usage errors like any other; execute prints them.
	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return UsageError(cmd, err)
	})
	root.AddCommand(newInitCmd())
	root.AddCommand(newVersionCmd())
	root.AddCommand(newUpdateCmd())
	root.AddCommand(newGenManCmd())
	return root
}
