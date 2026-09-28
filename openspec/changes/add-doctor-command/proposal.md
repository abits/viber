# Proposal

## Why

A viber scaffold depends on up to six external tools (`openspec`, `git`, `gh` with a login, `jq`, `markdownlint-cli2`, `claude`), but only `openspec` is checked, and only by `viber init`. Every other gap surfaces mid-workflow as a failing `make repo-init`, a silent no-op of the `tasks.md` sync hook, or a `make lint-md` that cannot find its binary. A single developer has nobody to ask, so the first failure costs a debugging detour. This was observed repeatedly while dogfooding viber on itself.

## What Changes

- New command `viber doctor` that checks every external tool a scaffold relies on and prints one line per check: status, tool, detected version or problem, and, on failure, the exact command that fixes it.
- Checks are split into **required** (`openspec`, `git`) and **optional** (`gh`, `gh` authentication, `jq`, `markdownlint-cli2`, `claude`). Optional tools only power individual scaffold features, so their absence is a warning, not an error.
- Exit code follows the existing contract: `0` when every required check passes (warnings allowed), `1` when a required check fails, `2` for usage errors.
- A `--json` flag emits the same results machine-readably, so scripts and CI can consume them.
- GNU-style help text in `internal/cmd/docs/doctor.txt`, matching the other commands; README "Subcommands" table gains a row.

Out of scope for this change: checks inside a scaffolded project (hook health, lockfile), auto-installing tools, and Windows-specific install hints beyond what the tools document themselves.

## Capabilities

### New Capabilities

- `doctor`: environment diagnostics for viber and the scaffolds it generates, covering which tools are checked, how results are classified and reported, and how the outcome maps to exit codes.

### Modified Capabilities

None. `viber init` keeps its own `openspec` pre-check unchanged.

## Impact

- **Code**: new `internal/cmd/doctor.go` and `internal/cmd/docs/doctor.txt`; a new `internal/doctor` package holding the check definitions and runner, so `cmd` stays a thin wiring layer and the checks are unit-testable with a fake command executor (as `internal/openspec` already does).
- **Dependencies**: none added; checks shell out via `os/exec`.
- **Docs**: README "Subcommands" table and the "Requirements" section, which can then point to `viber doctor`.
- **Compatibility**: purely additive; no existing command changes behavior.
