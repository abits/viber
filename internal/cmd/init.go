package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/abits/viber/internal/gitrepo"
	"github.com/abits/viber/internal/openspec"
	"github.com/abits/viber/internal/steps"
	"github.com/abits/viber/internal/templates"
	"github.com/abits/viber/internal/tty"
	"github.com/abits/viber/internal/wizard"
)

func newInitCmd() *cobra.Command {
	var (
		name, desc, remote, from string
		force, noTUI             bool
	)
	cmd := &cobra.Command{
		Use:   "init [name] [dir]",
		Short: "Initialize and populate a new agentic coding project.",
		// text lives in docs/init.txt
		Long: initLong,
		Args: usageArgs(cobra.MaximumNArgs(2)),
		RunE: func(cmd *cobra.Command, args []string) error {
			seed := wizard.Answers{
				Name:        name,
				Description: desc,
				Remote:      remote,
				From:        from,
				Force:       force,
			}
			if len(args) > 0 && seed.Name == "" {
				seed.Name = args[0]
			}
			if len(args) > 1 {
				seed.Dir = args[1]
			}
			if seed.Name != "" && seed.Dir == "" {
				seed.Dir = "./" + seed.Name
			}
			return runInit(cmd, seed, noTUI)
		},
	}
	f := cmd.Flags()
	f.StringVar(&name, "name", "", "project name (required in --no-tui)")
	f.StringVar(&desc, "desc", "", "one-line description (shown in README/CLAUDE.md)")
	f.StringVar(&remote, "remote", "", "git remote to add as 'origin' (optional)")
	f.StringVar(&from, "from", "", "fetch templates from GitHub: owner/repo[@ref]")
	f.BoolVar(&force, "force", false, "overwrite existing files in the destination")
	f.BoolVar(&noTUI, "no-tui", false, "disable the interactive wizard; require all values as flags")
	return cmd
}

func runInit(cmd *cobra.Command, seed wizard.Answers, noTUI bool) error {
	ctx := cmd.Context()
	haveAll := seed.Name != "" && seed.Dir != ""
	canSkip := noTUI || haveAll
	interactive := !canSkip && tty.IsInteractive(os.Stdin)

	ans, err := resolveAnswers(ctx, cmd, seed, interactive)
	if err != nil {
		return err
	}

	if err := checkDest(ans.Dir, ans.Force); err != nil {
		return err
	}

	src, err := templates.Resolve(ctx, ans.From)
	if err != nil {
		return err
	}

	data := templates.Data{Name: ans.Name, Description: ans.Description}

	// VerifyOpenspec has no side effects and must stay first: every step after
	// it writes to disk, so failing later would strand a half-scaffolded dir.
	stepList := []steps.Step{
		steps.VerifyOpenspec(),
		steps.Mkdir(ans.Dir),
		steps.GitInit(ans.Dir),
		steps.RenderTemplates(src, ans.Dir, data, ans.Force),
	}

	if interactive {
		err = steps.RunSpinner(ctx, stepList)
	} else {
		err = steps.RunPlain(ctx, stepList, cmd.OutOrStdout())
	}
	if err != nil {
		return err
	}

	if err := openspec.Init(ctx, ans.Dir); err != nil {
		return err
	}

	if ans.Remote != "" {
		if err := gitrepo.AddRemote(ctx, ans.Dir, "origin", ans.Remote); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not add remote %q: %v\n", ans.Remote, err)
		}
	}

	printNextSteps(cmd.OutOrStdout(), ans.Dir)
	return nil
}

// resolveAnswers fills in whatever the caller did not supply, either by running
// the wizard or by rejecting an incomplete non-interactive invocation.
func resolveAnswers(ctx context.Context, cmd *cobra.Command, seed wizard.Answers, interactive bool) (wizard.Answers, error) {
	if interactive {
		return wizard.Run(ctx, seed)
	}
	if seed.Name == "" || seed.Dir == "" {
		return wizard.Answers{}, UsageError(cmd, errors.New(
			"name and destination directory are required in non-interactive mode "+
				"(provide as positional args or via --name and a dir)"))
	}
	seed.Dir = filepath.Clean(seed.Dir)
	return seed, nil
}

func checkDest(dir string, force bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if force || len(entries) == 0 {
		return nil
	}
	return fmt.Errorf("destination %q is not empty (use --force to overwrite)", dir)
}

func printNextSteps(w io.Writer, dir string) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Next steps:")
	fmt.Fprintf(w, "  cd %s\n", dir)
	fmt.Fprintln(w, "  # fill in intend.md")
	fmt.Fprintln(w, "  claude")
	fmt.Fprintln(w, "  /opsx:explore")
	fmt.Fprintln(w)
}
