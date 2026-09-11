package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

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
		name, module, desc, remote, from string
		force, noTUI                     bool
	)
	cmd := &cobra.Command{
		Use:   "init [name] [dir]",
		Short: "Initialize and populate a new agentic coding project.",
		Long: `NAME
    viber init - initialize and populate a new agentic coding project

SYNOPSIS
    viber init [name] [dir] [flags]

DESCRIPTION
    Scaffolds a new project directory populated from an embedded template set,
    initializes a git repository, verifies that 'openspec' is installed, and
    runs 'openspec init' so Claude Code's /opsx:explore command is available.

    With no positional arguments and a terminal on stdin, an interactive
    Bubble Tea wizard collects the required values.

    With positional arguments, viber runs non-interactively:
      viber init <name>          → dir defaults to ./<name>
      viber init <name> <dir>    → fully specified

    Flags can also supply values (useful in CI): --name, --module, --desc,
    --remote, --from, --force. --no-tui forces non-interactive mode even
    on a terminal.

EXAMPLES
    Interactive wizard:
        viber init

    Name only (dir defaults to ./myproj):
        viber init myproj

    Name and directory:
        viber init myproj ~/code/myproj

    Fully flagged (CI):
        viber init --no-tui \
            --name=myproj \
            --module=github.com/me/myproj \
            --desc="a spec-driven side project"

    Overwrite an existing directory:
        viber init myproj --force

    Fetch templates from a GitHub repo instead of the embedded set:
        viber init myproj --from me/viber-templates@main

EXIT STATUS
    0    success
    1    runtime error (openspec missing, network, I/O)
    2    usage error (missing required values in non-interactive mode)
`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			seed := wizard.Answers{
				Name:        name,
				Module:      module,
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
	f.StringVar(&module, "module", "", "module path, e.g. github.com/you/name")
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

	var ans wizard.Answers
	if interactive {
		got, err := wizard.Run(ctx, seed)
		if err != nil {
			return err
		}
		ans = got
	} else {
		if seed.Name == "" || seed.Dir == "" {
			return UsageError(errors.New("name and destination directory are required in non-interactive mode (provide as positional args or via --name and a dir)"))
		}
		ans = seed
		ans.Dir = filepath.Clean(ans.Dir)
	}

	if err := checkDest(ans.Dir, ans.Force); err != nil {
		return err
	}

	src, err := templates.Resolve(ctx, ans.From)
	if err != nil {
		return err
	}

	data := templates.Data{
		Name:        ans.Name,
		Module:      ans.Module,
		Description: ans.Description,
		Remote:      ans.Remote,
		Year:        time.Now().Year(),
	}

	stepList := []steps.Step{
		steps.Mkdir(ans.Dir),
		steps.GitInit(ans.Dir),
		steps.RenderTemplates(src, ans.Dir, data, ans.Force),
		steps.VerifyOpenspec(),
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
