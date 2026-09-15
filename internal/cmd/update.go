package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/abits/viber/internal/ghfetch"
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
		// text lives in docs/update.txt
		Long: updateLong,
		Args: usageArgs(cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Validate input before checking platform support, so a bad
			// --repo is still reported as a usage error (exit 2) on Windows
			// instead of being masked by the runtime error below.
			owner, repo, err := splitRepoSpec(repoSpec)
			if err != nil {
				return UsageError(cmd, err)
			}
			if runtime.GOOS == "windows" {
				return fmt.Errorf("update is not supported on Windows in this version")
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

// splitRepoSpec parses an "owner/name" GitHub repository reference. --repo
// has no concept of a ref, so any "@ref" suffix ghfetch.ParseRepoSpec finds
// is simply discarded.
func splitRepoSpec(s string) (owner, repo string, err error) {
	owner, repo, _, err = ghfetch.ParseRepoSpec(s)
	if err != nil {
		return "", "", fmt.Errorf("--repo must be owner/name (got %q)", s)
	}
	return owner, repo, nil
}
