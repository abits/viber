package steps

import (
	"context"
	"io/fs"
	"os"

	"github.com/abits/viber/internal/gitrepo"
	"github.com/abits/viber/internal/openspec"
	"github.com/abits/viber/internal/templates"
)

type stepFn struct {
	name string
	run  func(ctx context.Context) error
}

func (s stepFn) Name() string                  { return s.name }
func (s stepFn) Run(ctx context.Context) error { return s.run(ctx) }

func Mkdir(dir string) Step {
	return stepFn{
		name: "create project directory",
		run: func(ctx context.Context) error {
			return os.MkdirAll(dir, 0o755)
		},
	}
}

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

func RenderTemplates(src fs.FS, dir string, data templates.Data, force bool) Step {
	return stepFn{
		name: "render templates",
		run: func(ctx context.Context) error {
			return templates.Render(src, dir, data, force)
		},
	}
}

func VerifyOpenspec() Step {
	return stepFn{
		name: "verify openspec installation",
		run: func(ctx context.Context) error {
			return openspec.Verify(ctx)
		},
	}
}
