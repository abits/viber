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

type usageErr struct{ err error }

func (u usageErr) Error() string { return u.err.Error() }
func (u usageErr) Unwrap() error { return u.err }

func UsageError(err error) error { return usageErr{err} }

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var u usageErr
	if errors.As(err, &u) {
		return 2
	}
	return 1
}

func Execute(v, c, d string) int {
	info := version.Info{Version: v, Commit: c, Date: d}
	root := newRootCmd(info)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	err := root.ExecuteContext(ctx)
	if err != nil {
		var u usageErr
		if !errors.As(err, &u) {
			fmt.Fprintln(root.ErrOrStderr(), "Error:", err)
		}
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
        viber init myproj --no-tui --name=myproj --module=github.com/me/myproj

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
	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		cmd.Println(cmd.UsageString())
		return UsageError(err)
	})
	root.AddCommand(newInitCmd())
	root.AddCommand(newVersionCmd())
	root.AddCommand(newUpdateCmd())
	root.AddCommand(newGenManCmd())
	return root
}
