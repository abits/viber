package templates

import (
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// renderScaffold renders the embedded template set and returns a reader for
// the rendered files, keyed by slash-separated path relative to the scaffold.
func renderScaffold(t *testing.T, data Data) (dst string, read func(rel string) string) {
	t.Helper()
	src, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	dst = t.TempDir()
	if err := Render(src, dst, data, false); err != nil {
		t.Fatalf("render embedded templates: %v", err)
	}
	return dst, func(rel string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dst, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read rendered %s: %v", rel, err)
		}
		return string(b)
	}
}

func TestAgentsMDCarriesProjectInstructions(t *testing.T) {
	_, read := renderScaffold(t, Data{Name: "myproj", Description: "A demo"})

	agents := read("AGENTS.md")

	for _, want := range []string{"myproj", "A demo", "intend.md", "openspec/", "make lint-md"} {
		if !strings.Contains(agents, want) {
			t.Errorf("AGENTS.md does not contain %q", want)
		}
	}
}

func TestAgentsMDHasNoClaudeSpecificContent(t *testing.T) {
	_, read := renderScaffold(t, Data{Name: "myproj", Description: "A demo"})

	agents := read("AGENTS.md")

	for _, banned := range []string{".claude/", "/plugin", "@agent-", "/repo-init", "/sync-issues", "settings.json"} {
		if strings.Contains(agents, banned) {
			t.Errorf("AGENTS.md contains Claude-specific %q", banned)
		}
	}
}

func TestClaudeMDImportsAgentsMD(t *testing.T) {
	_, read := renderScaffold(t, Data{Name: "myproj", Description: "A demo"})

	claude := read("CLAUDE.md")

	if !slices.Contains(strings.Split(claude, "\n"), "@AGENTS.md") {
		t.Error("CLAUDE.md has no line that is exactly @AGENTS.md")
	}
}

func TestClaudeMDDoesNotDuplicateSharedFacts(t *testing.T) {
	_, read := renderScaffold(t, Data{Name: "myproj", Description: "A demo"})

	claude := read("CLAUDE.md")

	for _, dup := range []string{"make lint-md", "A demo"} {
		if strings.Contains(claude, dup) {
			t.Errorf("CLAUDE.md repeats %q, which AGENTS.md already carries", dup)
		}
	}
}

func TestTeamDocListsEveryAgent(t *testing.T) {
	dst, read := renderScaffold(t, Data{Name: "myproj", Description: "A demo"})
	entries, err := os.ReadDir(filepath.Join(dst, ".claude", "agents"))
	if err != nil {
		t.Fatalf("list rendered agents: %v", err)
	}

	team := read("docs/team.md")

	if len(entries) != 9 {
		t.Fatalf("got %d agent files, want 9", len(entries))
	}
	for _, e := range entries {
		ref := path.Join(".claude/agents", e.Name())
		if !strings.Contains(team, ref) {
			t.Errorf("docs/team.md does not list %s", ref)
		}
	}
}

func TestTeamReferencesPointToTeamDoc(t *testing.T) {
	dst, read := renderScaffold(t, Data{Name: "myproj", Description: "A demo"})
	entries, err := os.ReadDir(filepath.Join(dst, ".claude", "agents"))
	if err != nil {
		t.Fatalf("list rendered agents: %v", err)
	}

	for _, rel := range []string{"CLAUDE.md", "README.md"} {
		if !strings.Contains(read(rel), "docs/team.md") {
			t.Errorf("%s does not reference docs/team.md", rel)
		}
	}
	for _, e := range entries {
		rel := path.Join(".claude/agents", e.Name())
		body := read(rel)
		if strings.Contains(body, "AGENTS.md") {
			t.Errorf("%s references AGENTS.md; team docs live in docs/team.md", rel)
		}
		if strings.Contains(body, "marketplace add") && !strings.Contains(body, "docs/team.md") {
			t.Errorf("%s mentions the marketplace step without pointing to docs/team.md", rel)
		}
	}
}
