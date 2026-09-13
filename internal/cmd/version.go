package cmd

import "github.com/spf13/cobra"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the viber version.",
		// text lives in docs/version.txt
		Long: versionLong,
		Args: usageArgs(cobra.NoArgs),
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println(cmd.Root().Version)
		},
	}
}
