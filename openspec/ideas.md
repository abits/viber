# Feature ideas

Backlog of ideas that are not yet OpenSpec changes. OpenSpec models only specs (what the system does) and changes (work in flight), so ideas live here until one is picked up.

To pick one up: run `openspec new change "<name>"` (or `/opsx:propose`), carry the idea's rationale into `proposal.md`, and delete its row here. Implemented ideas are not listed; see `openspec/changes/archive/`.

Origin: feature brainstorm against other scaffolding tools, 2026-10-03. Letters are the original idea IDs; C (project checks in `viber doctor`) shipped as `add-project-doctor-checks`.

## Overview

| ID | Idea | Model | Benefit | Effort | Depends on |
| --- | --- | --- | --- | --- | --- |
| K | Resolve the `AGENTS.md` collision | agents.md convention | high | low | - |
| F | `viber init --dry-run` | Angular schematics, Yeoman | medium | low | - |
| A | `.viber.yml` manifest, then `viber diff`, then `viber upgrade` | Copier, Cruft | high | medium to high | - |
| B | Ship agents, skills, rules as a Claude Code plugin | Claude Code plugins | high | medium | decide against A |
| G | `--data-file` / `--defaults` for non-interactive init | Copier | medium | low | A (same file format) |
| E | Stack presets (`--stack go\|python\|ts\|none`) | create-t3-app, Spring Initializr | high | medium | A |
| D | `viber add agent\|rule\|skill <name>` | shadcn/ui, Rails generators | medium | medium | A |
| L | Devcontainer and Claude Code cloud SessionStart hook | Dev Containers | medium | low | - |
| M | `.mcp.json` presets in the wizard | - | low | low | - |
| I | `viber template lint` for template authors | pytest-cookies | low | low | - |
| N | `viber init --intent "..."` drafts `intend.md` via `claude -p` | - | medium | low | - |
| H | Template manifest with declared questions | Copier, Backstage | medium | high | A |
| J | Post-init tasks, remote sets only with `--trust` | Cookiecutter hooks, Copier | low | medium | H |
| O | Pluggable spec backend (`openspec\|speckit\|none`) | GitHub Spec Kit | low | high | - |

Suggested order: K, F, A (manifest and `diff` only). Decide A versus B in an ADR before building `upgrade`. Defer H and J: they push viber toward a general template framework, where Copier is already mature.

## Notes

### K: `AGENTS.md` collision

`AGENTS.md` is a cross-tool convention (Codex, Cursor, Gemini CLI, and others read it as a README for coding agents). viber uses it to describe the sub-agent team, so other agents find team prose instead of build and test instructions. Option: move the team description to `docs/team.md` and render `AGENTS.md` as an agent-neutral summary of `CLAUDE.md`. Touches the template, the dogfood file list, and `README.md`.

### F: `--dry-run`

Print the files `init` would write, and with `--force` a diff against what exists. Cheap safety for merging into an existing directory.

### A: Manifest, diff, upgrade

`init` writes `.viber.yml` (template source and ref, viber version, `name` and `desc`) from a new step after `RenderTemplates`, not from a template, so remote sets get it too and `templates.Data` needs no new fields. Unlocks: the drift check in `viber doctor` (deferred from `add-project-doctor-checks`), `viber diff`, and `viber upgrade` as a three-way merge (old render as base, new render as theirs, project as ours; `git merge-file` per file). `make dogfood` could read `DOGFOOD_NAME` and `DOGFOOD_DESC` from it. Fix the manifest format only when `diff` or `upgrade` needs it.

### B: Plugin distribution

Alternative to A for the agent team: central updates, no drift, but harder to customize locally. A hybrid is possible (agents as a plugin, `CLAUDE.md` and `intend.md` copied).

### G: Answers file

Reproducible scaffolding in CI. Should be the same format as `.viber.yml`.

### E: Stack presets

The rules already list `go test` / `pytest` / `npm test` side by side, a sign that one generic template is at its limit. A preset renders language-specific rules, Makefile targets, the Stop-hook command, and lint config. New `templates.Data` fields must be referenced by a template (`TestEmbeddedTemplatesUseEveryDataField`).

### D: Components after init

`viber add agent data-engineer`, `viber add rule python-typing`, from the embedded set or a `--from` repo. Code is copied and owned, as with shadcn/ui. Needs a component manifest format.

### L: Cloud readiness

`.devcontainer/devcontainer.json` and a SessionStart hook that installs `openspec`, `gh`, and the linters, so the scaffold's Stop hook works in Claude Code on the web.

### M: MCP presets

Optional `.mcp.json` entries (for example GitHub, Context7) chosen in the wizard.

### I: Template lint

Expose the `TestEmbeddedTemplatesUseEveryDataField` check, duplicate-target detection, and archive limits as a command for authors of `--from` template sets.

### N: Draft `intend.md`

Calls Claude Code headless to draft Problem, Goals, and Constraints. Adds a second runtime dependency, against the constraint in `intend.md`; acceptable only as an optional flag that degrades gracefully.

### H: Declared questions

Remote template sets declare their own prompts in a manifest instead of being limited to `name` and `desc`. Large step toward "Copier in Go"; weigh against the single-maintainer constraint.

### J: Post-init tasks

Run commands after rendering (`npm install`, `make repo-init`). Without an explicit opt-in, `--from` would become remote code execution: allow tasks from the embedded set only, or require `--trust` for remote sets, as Copier does.

### O: Spec backend

Broadens the audience but dilutes the profile. Only on concrete demand.

## Sources

- Copier: <https://copier.readthedocs.io/>
- Cruft: <https://cruft.github.io/cruft/>
- Cookiecutter hooks: <https://cookiecutter.readthedocs.io/en/stable/advanced/hooks.html>
- Yeoman: <https://yeoman.io/authoring/composability.html>
- Angular schematics: <https://angular.dev/tools/cli/schematics>
- shadcn/ui CLI: <https://ui.shadcn.com/docs/cli>
- create-t3-app: <https://create.t3.gg/>
- Spring Initializr: <https://start.spring.io/>
- Backstage software templates: <https://backstage.io/docs/features/software-templates/>
- GitHub Spec Kit: <https://github.com/github/spec-kit>
- Claude Code plugins: <https://code.claude.com/docs/en/plugins>
- AGENTS.md: <https://agents.md/>
