# Intend: viber

## Problem

Starting a project for AI-assisted ("vibe") coding with Claude Code takes the same setup every time: a `CLAUDE.md`, rules, a sub-agent team, OpenSpec for requirements, a GitHub repo, and a way to keep OpenSpec tasks visible as issues. Done by hand, that setup drifts between projects and is skipped when time is short, so the agents work without context and the specs never reach the tracker.

## Goals

- One command (`viber init`) produces a ready-to-use project: directory, git, rendered template set, `openspec init`.
- The scaffold is opinionated but replaceable: `--from owner/repo[@ref]` swaps in a remote template set.
- viber keeps itself current (`viber update`) and is scaffolded with its own template set; `make dogfood-check` proves it in CI.
- Every template field is used (`TestEmbeddedTemplatesUseEveryDataField`), so the prompt never asks for data it drops.

## Constraints

- Single static Go binary (CGO off) for linux/darwin/windows on amd64/arm64, released with GoReleaser.
- The only runtime dependency of `init` is `openspec`; `gh` and `jq` are needed only for the GitHub scripts in the scaffold.
- Exit codes are a contract: `0` success, `1` runtime error, `2` usage error.
- A failed render leaves nothing behind on a fresh scaffold; long-running work is cancellable via `context`.
- Maintained by one person with Claude Code: conventions are enforced by hooks, lint, and CI rather than review alone.
