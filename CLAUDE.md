# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository. Project instructions shared with other coding agents live in `AGENTS.md`, imported here:

@AGENTS.md

## OpenSpec in Claude Code

Use `/opsx:explore` to think a change through, `/opsx:propose` to draft it, `/opsx:apply` to implement it, and `/opsx:archive` to finish it.

## Hooks

- A `PostToolUse` hook auto-formats `.go` files after every Write/Edit — don't format by hand.
- A `Stop` hook runs `golangci-lint run ./... && go test ./...` at the end of each turn. Treat a red
  result as blocking.

## Team

Nine role-specific sub-agents live in `.claude/agents/`. See `docs/team.md` for the flow, hand-off patterns, and the `/repo-init` and `/sync-issues` commands.

## Skills

- `/commit-push-pr` — user-invoked: stage, commit, push, open a PR via `gh`.
- `/repo-init`, `/sync-issues` — scaffold commands; see `docs/team.md`. The `PostToolUse` hook in
  `.claude/settings.json` mirrors `openspec/changes/*/tasks.md` into GitHub Issues on write.
