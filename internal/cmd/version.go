package cmd

import "github.com/spf13/cobra"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the viber version.",
		Long: `NAME
    viber version - print viber's version, commit, and build date

SYNOPSIS
    viber version

DESCRIPTION
    Prints the same string as 'viber --version'.
`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println(cmd.Root().Version)
		},
	}
}
