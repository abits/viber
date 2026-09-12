package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/abits/viber/internal/updater"
)

func newUpdateCmd() *cobra.Command {
	var (
		repoSpec string
		dest     string
	)
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Download the latest release and overwrite the installed binary.",
		Long: `NAME
    viber update - download the latest release and overwrite the installed binary

SYNOPSIS
    viber update [--repo <owner/name>] [--dest <path>]

DESCRIPTION
    Queries the latest GitHub release for owner/repo (default: abits/viber),
    downloads the archive matching the current OS/architecture, extracts the
    'viber' binary, and atomically writes it to <dest> (default:
    $HOME/bin/viber). File mode is set to 0755.

    Windows is not supported in this version; use 'go install' or 'make install'
    on Windows.

EXAMPLES
    Update the default install:
        viber update

    Update from a fork:
        viber update --repo myorg/viber

    Install into a custom path:
        viber update --dest /usr/local/bin/viber

ENVIRONMENT
    GITHUB_TOKEN
        Optional. Sent as a bearer token when calling the GitHub API and when
        downloading release assets. Useful for private repos or when hitting
        anonymous rate limits.

EXIT STATUS
    0    success
    1    runtime error (network, extraction, filesystem)
    2    usage error (bad --repo)
`,
		Args: usageArgs(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if runtime.GOOS == "windows" {
				return fmt.Errorf("update is not supported on Windows in this version")
			}
			owner, repo, err := splitRepoSpec(repoSpec)
			if err != nil {
				return UsageError(cmd, err)
			}
			if dest == "" {
				d, err := updater.DefaultDest("viber")
				if err != nil {
					return err
				}
				dest = d
			}
			ctx := cmd.Context()
			cmd.Printf("Checking latest release for %s/%s...\n", owner, repo)
			rel, err := updater.Latest(ctx, owner, repo)
			if err != nil {
				return err
			}
			cmd.Printf("Latest: %s\n", rel.TagName)
			cmd.Printf("Downloading and installing to %s...\n", dest)
			newVer, err := updater.Install(ctx, rel, "viber", dest)
			if err != nil {
				return err
			}
			cmd.Printf("Installed %s to %s\n", newVer, dest)
			return nil
		},
	}
	cmd.Flags().StringVar(&repoSpec, "repo", "abits/viber", "GitHub release source, owner/name")
	cmd.Flags().StringVar(&dest, "dest", "", "install destination (default: $HOME/bin/viber)")
	return cmd
}

// splitRepoSpec parses an "owner/name" GitHub repository reference.
func splitRepoSpec(s string) (owner, repo string, err error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("--repo must be owner/name (got %q)", s)
	}
	return parts[0], parts[1], nil
}
