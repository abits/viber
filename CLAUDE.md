# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`viber` is a Go CLI that scaffolds a "vibe coding" project wired for Claude Code + OpenSpec.
`viber init` verifies that `openspec` is installed, creates a directory, initializes git, renders
a template set into it, and runs `openspec init`.

viber is scaffolded with itself. Read `intend.md` for the problem, goals, and constraints.
Requirements and tech decisions are drafted via OpenSpec (`/opsx:explore`); the outputs live
under `openspec/`.

## Layout

    cmd/viber/            main; injects version/commit/date via ldflags
    internal/cmd/         Cobra command tree; owns error reporting + exit codes
    internal/wizard/      Bubble Tea prompt for the values init needs
    internal/steps/       ordered, side-effecting units of work + two runners
    internal/templates/   embedded and remote template sets, rendering
    internal/updater/     self-update from GitHub releases
    internal/{gitrepo,openspec,tty,version}/  thin helpers

The dependency direction is `cmd -> steps -> {templates, gitrepo, openspec}`. Keep it acyclic;
`steps` is what lets the TUI and the plain runner share one definition of the work.

## Invariants

- **Exit codes**: `0` success, `1` runtime error, `2` usage error. `internal/cmd` is the single
  place that prints user-facing errors — commands and step runners return errors, they do not
  print them. Mark input errors with `UsageError(cmd, err)` so they exit 2 and show usage.
- **Ordering in `runInit`**: side-effect-free checks (`steps.VerifyOpenspec`) come first, then
  `RenderTemplates` (which creates the destination itself, atomically when it doesn't already
  exist - see `templates.Render`), then `GitInit`. A render failure on a fresh scaffold now
  leaves nothing behind; don't reorder `GitInit` before `RenderTemplates`, it requires the
  directory to already exist.
- **`templates.Data`**: every field must be referenced by at least one template.
  `TestEmbeddedTemplatesUseEveryDataField` enforces this — do not add a field "for later".
- **Context**: `steps.Runner` checks `ctx.Err()` between steps, and both Bubble Tea programs are
  built with `tea.WithContext`. Keep new long-running work cancellable.

## Conventions

- Format with `goimports -w` (fallback: `gofmt -w`). A `PostToolUse` hook auto-formats `.go` files
  after every Write/Edit — don't format by hand.
- A `Stop` hook runs `golangci-lint run ./... && go test ./...` at the end of each turn. Treat a red
  result as blocking. Lint config: `.golangci.yml` (v2, `default: standard`).
- Exported identifiers carry doc comments; see <https://go.dev/doc/comment>.
- `make test` / `make lint` / `make build`; `make bump-patch` commits and tags a release.

## Dogfooding

Files listed in `DOGFOOD_FILES` (`Makefile`) are owned by `internal/templates/default/`: edit the
template, then run `make dogfood`. CI runs `make dogfood-check` and fails on drift. `CLAUDE.md`,
`README.md`, `Makefile`, `.gitignore`, and `.claude/settings.json` are repo-owned supersets of
their templates; when a template for one of them changes, merge the change here by hand.

## Team

Nine role-specific sub-agents live in `.claude/agents/`. See `AGENTS.md` for the flow and hand-off
patterns.

## Markdown

Lint prose with `make lint-md` (`markdownlint-cli2`, config `.markdownlint.yaml`).

## Skills

- `/commit-push-pr` — user-invoked: stage, commit, push, open a PR via `gh`.
- `/repo-init`, `/sync-issues` — scaffold commands; see `AGENTS.md`. The `PostToolUse` hook in
  `.claude/settings.json` mirrors `openspec/changes/*/tasks.md` into GitHub Issues on write.
