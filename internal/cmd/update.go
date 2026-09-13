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
		// text lives in docs/update.txt
		Long: updateLong,
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
