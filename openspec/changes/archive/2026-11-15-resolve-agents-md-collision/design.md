# Design

## Context

See proposal.md for motivation. Current state:

- `internal/templates/default/AGENTS.md.tmpl` holds only team text and uses no template data.
- `CLAUDE.md.tmpl` carries the project name and description, the `intend.md` pointer, a Team section pointing to `AGENTS.md`, and `make lint-md`.
- Nine `.claude/agents/*.md.tmpl` files and `README.md.tmpl` point to `AGENTS.md` for the marketplace step or the team.
- In viber's own repo, `AGENTS.md` is template-owned (listed in `DOGFOOD_FILES`), while `CLAUDE.md` and `README.md` are repo-owned supersets.
- `TestEmbeddedTemplatesUseEveryDataField` requires every `templates.Data` field to be used by some template; the scaffold layout test lists every rendered path.
- The renderer already emits nested paths (`.claude/agents/`), so `docs/team.md.tmpl` needs no code change.

## Goals / Non-Goals

**Goals:**

- One source for agent-neutral facts in a scaffold (`AGENTS.md`), imported by `CLAUDE.md`.
- Keep the team text byte-identical apart from the new location, so the move is reviewable as a rename.

**Non-Goals:**

- Migrating existing scaffolds. There is no upgrade command (backlog idea A).
- Rewriting or shortening the team text itself.
- New `templates.Data` fields.

## Decisions

### `CLAUDE.md` imports `AGENTS.md` instead of duplicating it

Claude Code expands `@path` imports in `CLAUDE.md`, so `@AGENTS.md` gives Claude the shared facts without a second copy. Alternatives: a standalone summary in both files (drifts, chosen against by the user), a symlink `CLAUDE.md -> AGENTS.md` (loses the Claude-only sections, and symlinks are fragile on Windows, which CI tests).

### Name and description move to `AGENTS.md.tmpl`

`{{.Name}}` and `{{.Description}}` render in `AGENTS.md` and no longer in `CLAUDE.md`. Both fields stay referenced (`AGENTS.md.tmpl`, plus `README.md.tmpl` already uses them), so the data-field test keeps passing without a test change.

### `docs/team.md` via `git mv` of the template

`AGENTS.md.tmpl` is renamed to `docs/team.md.tmpl` unchanged except for its top heading, then a new `AGENTS.md.tmpl` is written. This keeps history and makes the diff a rename plus a new file. `docs/` is chosen over `.claude/TEAM.md` because the team page is also for humans and the README links to it.

### Ownership in viber's repo

`docs/team.md` replaces `AGENTS.md` in `DOGFOOD_FILES`. `AGENTS.md` becomes a repo-owned superset like `CLAUDE.md`: it starts from the rendered template and adds viber's layout, invariants, and conventions, which are agent-neutral and today live only in `CLAUDE.md`. The repo `CLAUDE.md` then imports `@AGENTS.md` and keeps only Claude-specific sections (hooks, team, skills). Alternative: keep `AGENTS.md` dogfooded. Rejected because a non-Claude agent working on viber would then see generic scaffold text instead of viber's invariants, the same problem this change fixes.

### Content checks as template tests

The spec's scenarios are checked by a render-and-grep test in `internal/templates`, next to the layout test: render the embedded set to `t.TempDir()`, assert the required strings in `AGENTS.md`/`CLAUDE.md`, the forbidden strings in `AGENTS.md`, and that no `.claude/agents/*` file mentions `AGENTS.md`.

## Risks / Trade-offs

- [Other tools may not expand `@AGENTS.md`] → Only Claude Code reads `CLAUDE.md`, and it does expand it; other tools read `AGENTS.md` directly.
- [Users of older scaffolds keep the collision] → Note it in the release's CHANGELOG entry with the manual fix (move `AGENTS.md` to `docs/team.md`, re-run `viber init` into a scratch directory, copy the new `AGENTS.md`).
- [Repo `AGENTS.md` and its template drift] → Same exposure as `CLAUDE.md` today; the CLAUDE.md "Dogfooding" section lists `AGENTS.md` among the hand-merged supersets.

## Migration Plan

Template change ships in the next release. No rollback concern: revert the commit. Existing scaffolds are untouched.
