package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newGenManCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "gen-man <dir>",
		Short:  "Generate man pages into <dir>.",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hdr := &doc.GenManHeader{Title: "VIBER", Section: "1"}
			return doc.GenManTree(cmd.Root(), hdr, args[0])
		},
	}
}
