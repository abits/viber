# Proposal

## Why

`AGENTS.md` is a cross-tool convention: Codex, Cursor, Gemini CLI, and other coding agents read it as the project's instructions file. Every viber scaffold puts the Claude Code sub-agent team description there, so those agents load marketplace setup, a role table, and agent-teams trade-offs instead of what the project is and how to work in it. Claude Code itself reads `CLAUDE.md`, so the same project facts are invisible to every non-Claude agent.

## What Changes

- The scaffold's `AGENTS.md` becomes the agent-neutral instructions file: project name and description, `intend.md`, the OpenSpec workflow, and `make lint-md`. It names no Claude-specific feature.
- The scaffold's `CLAUDE.md` imports `AGENTS.md` with Claude Code's `@AGENTS.md` syntax and adds only Claude-specific guidance (sub-agent team, slash commands), so shared facts have one source.
- The current team text (marketplace prerequisite, flow, invocation, agent teams mode, roles, slash commands) moves unchanged to a new scaffold file, `docs/team.md`.
- Every scaffold reference to `AGENTS.md` that means "the team docs" points to `docs/team.md` instead: the nine agent templates, `CLAUDE.md`, and `README.md`.
- viber's own repo follows: `docs/team.md` is template-owned (dogfooded); `AGENTS.md` becomes a repo-owned superset of its template, so it can carry viber's layout and invariants for non-Claude agents.
- **BREAKING** (layout only): a fresh scaffold now contains `docs/team.md`, and `AGENTS.md` has different content. Existing scaffolds are not migrated; viber has no upgrade command.

## Capabilities

### New Capabilities

- `agent-instructions`: which instruction files a scaffold ships for coding agents, what each contains, and how the agent-neutral and Claude-specific files relate.

### Modified Capabilities

None. `doctor` detects a scaffold by `CLAUDE.md`, `openspec/`, and `.claude/`; none of these change.

## Impact

- Templates: `internal/templates/default/AGENTS.md.tmpl` (rewritten), `CLAUDE.md.tmpl`, `README.md.tmpl`, `.claude/agents/*.md.tmpl` (9 files), new `docs/team.md.tmpl`.
- Tests: layout list in `internal/templates/embed_test.go` gains `docs/team.md`.
- Build: `DOGFOOD_FILES` in `Makefile` swaps `AGENTS.md` for `docs/team.md`.
- Repo-owned files: `AGENTS.md`, `CLAUDE.md`, `README.md` (layout tree and dogfooding section).
- No Go code paths, dependencies, or CLI flags change.
