package templates

import (
	"io/fs"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestParseFrom(t *testing.T) {
	cases := []struct {
		name             string
		in               string
		owner, repo, ref string
		wantErr          bool
	}{
		{"owner/repo defaults to main", "me/tpl", "me", "tpl", "main", false},
		{"explicit ref", "me/tpl@v2", "me", "tpl", "v2", false},
		{"ref with slashes", "me/tpl@feat/x", "me", "tpl", "feat/x", false},
		{"sha ref", "me/tpl@0f1e2d3", "me", "tpl", "0f1e2d3", false},
		{"empty ref after @ is rejected", "me/tpl@", "", "", "", true},
		{"missing slash", "metpl", "", "", "", true},
		{"empty owner", "/tpl", "", "", "", true},
		{"empty repo", "me/", "", "", "", true},
		{"empty string", "", "", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			owner, repo, ref, err := parseFrom(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseFrom(%q) = %q,%q,%q; want error", tc.in, owner, repo, ref)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFrom(%q): %v", tc.in, err)
			}
			if owner != tc.owner || repo != tc.repo || ref != tc.ref {
				t.Errorf("parseFrom(%q) = %q,%q,%q; want %q,%q,%q",
					tc.in, owner, repo, ref, tc.owner, tc.repo, tc.ref)
			}
		})
	}
}

// TestEmbeddedTemplatesRender executes the real embedded template set so a
// mistyped template field fails here instead of in a user's terminal.
func TestEmbeddedTemplatesRender(t *testing.T) {
	src, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	dst := t.TempDir()
	data := Data{Name: "myproj", Description: "a spec-driven side project"}
	if err := Render(src, dst, data, false); err != nil {
		t.Fatalf("embedded templates failed to render: %v", err)
	}
}

// TestEmbeddedTemplateSetLayout pins which files a scaffold produces, so that
// adding or renaming a template is a visible, intentional diff rather than a
// silent change to everyone's new projects.
func TestEmbeddedTemplateSetLayout(t *testing.T) {
	src, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	dst := t.TempDir()
	if err := Render(src, dst, Data{Name: "myproj", Description: "d"}, false); err != nil {
		t.Fatal(err)
	}

	want := []string{
		".claude/agents/architect.md",
		".claude/agents/backend-engineer.md",
		".claude/agents/designer.md",
		".claude/agents/devops-engineer.md",
		".claude/agents/frontend-engineer.md",
		".claude/agents/product-manager.md",
		".claude/agents/qa-engineer.md",
		".claude/agents/security-reviewer.md",
		".claude/agents/tech-lead.md",
		".claude/commands/repo-init.md",
		".claude/commands/sync-issues.md",
		".claude/rules/ai-output-style.md",
		".claude/rules/git-hygiene.md",
		".claude/rules/pre-commit-checklist.md",
		".claude/rules/security.md",
		".claude/rules/testing.md",
		".claude/settings.json",
		".gitignore",
		".markdownlint.yaml",
		"AGENTS.md",
		"CLAUDE.md",
		"Makefile",
		"README.md",
		"intend.md",
		"scripts/repo-init.sh",
		"scripts/sync-issues.sh",
	}
	got := renderedFiles(t, dst)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("scaffold layout changed:\n got = %v\nwant = %v", got, want)
	}
}

// TestEmbeddedTemplatesUseEveryDataField is the guard for the defect this test
// file was written for: Data once carried Module, Remote and Year, which the
// CLI collected from the user and no template ever referenced. A field that no
// template uses is a question asked for nothing.
func TestEmbeddedTemplatesUseEveryDataField(t *testing.T) {
	src, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	bodies := templateBodies(t, src)

	for _, f := range reflect.VisibleFields(reflect.TypeOf(Data{})) {
		t.Run(f.Name, func(t *testing.T) {
			needle := "{{." + f.Name
			for _, body := range bodies {
				if strings.Contains(body, needle) {
					return
				}
			}
			t.Errorf("Data.%s is referenced by no embedded template; "+
				"either use it in the template set or remove the field", f.Name)
		})
	}
}

func renderedFiles(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := fs.WalkDir(dirFS(root), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			out = append(out, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(out)
	return out
}

func templateBodies(t *testing.T, src fs.FS) []string {
	t.Helper()
	var out []string
	err := fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path.Base(p), ".tmpl") {
			return err
		}
		b, readErr := fs.ReadFile(src, p)
		if readErr != nil {
			return readErr
		}
		out = append(out, string(b))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Fatal("no .tmpl files found in the embedded set")
	}
	return out
}

func dirFS(root string) fs.FS { return os.DirFS(root) }
