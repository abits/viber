# Proposal

## Why

`viber doctor` checks the tools a scaffold needs but not the scaffold itself, and the scaffold's own wiring fails silently by design: the `tasks.md` sync hook in `.claude/settings.json` discards its output and ends in `|| true`, so a missing `scripts/sync-issues.sh`, a missing GitHub remote, or a broken `settings.json` produces no error anywhere. The `add-doctor-command` proposal deferred these in-project checks explicitly; with that change archived, the gap is the next thing a developer hits.

## What Changes

- When `viber doctor` runs inside a scaffolded project, it additionally runs a group of **project checks** after the existing tool checks. Outside a scaffold, its behavior is unchanged.
- A directory counts as a scaffold when its project root (the git top-level directory, or the working directory outside git) contains `CLAUDE.md`, `openspec/`, and `.claude/`. This is a heuristic; an explicit manifest is deferred to a separate change.
- Project checks answer only "does the setup work?":
  - `.claude/settings.json` parses as JSON (required: an invalid file disables every hook).
  - `intend.md` is present.
  - Every script a hook in `settings.json` references exists.
  - The `origin` remote points to github.com.
  - The issue sync is active: aggregates `jq`, `gh` authentication, the hook script, and the remote, and names the first missing prerequisite.
  - `.env` is not tracked by git (required: a tracked secret is an active leak).
  - `.env` is ignored by git.
- Human output prints the project checks under their own heading with the project root; `--json` entries gain a `group` field (`tools` or `project`). The JSON document stays a flat array, so existing consumers keep working.
- Exit codes keep the existing contract: a failed required project check exits `1`, warnings exit `0`.

Out of scope: workflow and process hints (unfilled `intend.md` placeholders, completed but unarchived changes); drift against the template and the `.viber.yml` manifest it needs; an explicit project-directory argument; any `--fix` mode. Doctor stays read-only.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `doctor`: adds scaffold detection and the project check group; extends the JSON entry with `group`; generalizes the cancellation and read-only guarantees to project checks.

## Impact

- **Code**: `internal/doctor` gains a probe abstraction beside the exec-based checks, multi-dependency support for `DependsOn`, scaffold detection, and the project check table; `internal/cmd/doctor.go` renders the second group and the `group` field.
- **Docs**: `internal/cmd/docs/doctor.txt` and the README "Subcommands" row describe the project checks.
- **Dependencies**: none added; git queries shell out through the existing executor seam.
- **Compatibility**: additive. Output outside a scaffold is byte-identical; JSON gains one field.
