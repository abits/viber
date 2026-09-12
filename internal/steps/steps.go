package steps

import (
	"context"
	"io/fs"
	"os"

	"github.com/abits/viber/internal/gitrepo"
	"github.com/abits/viber/internal/openspec"
	"github.com/abits/viber/internal/templates"
)

// dirMode is the permission bit set used for every directory viber creates.
const dirMode = 0o755

type stepFn struct {
	name string
	run  func(ctx context.Context) error
}

func (s stepFn) Name() string                  { return s.name }
func (s stepFn) Run(ctx context.Context) error { return s.run(ctx) }

// VerifyOpenspec checks that the openspec binary is on PATH. It has no side
// effects and belongs first in a scaffold run, so that a missing dependency
// fails before anything is written to disk.
func VerifyOpenspec() Step {
	return stepFn{
		name: "verify openspec installation",
		run:  openspec.Verify,
	}
}

// Mkdir creates the destination directory, including any missing parents.
func Mkdir(dir string) Step {
	return stepFn{
		name: "create project directory",
		run: func(context.Context) error {
			return os.MkdirAll(dir, dirMode)
		},
	}
}

// GitInit initializes a git repository in dir unless one is already there.
func GitInit(dir string) Step {
	return stepFn{
		name: "initialize git repository",
		run: func(ctx context.Context) error {
			if gitrepo.IsRepo(dir) {
				return nil
			}
			return gitrepo.Init(ctx, dir)
		},
	}
}

// RenderTemplates expands the template set src into dir.
func RenderTemplates(src fs.FS, dir string, data templates.Data, force bool) Step {
	return stepFn{
		name: "render templates",
		run: func(context.Context) error {
			return templates.Render(src, dir, data, force)
		},
	}
}
