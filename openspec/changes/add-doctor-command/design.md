# Design

## Context

`internal/openspec.Verify` already checks one tool by running `openspec --version` through a package-level `execCommand` variable that tests replace with a fake. `internal/cmd` owns all user-facing error output and the exit-code contract (see `CLAUDE.md` Invariants). `internal/steps` runs ordered, side-effecting units of work for `init`. Doctor is read-only and has no ordering dependencies between checks, so it does not fit the steps model.

## Goals / Non-Goals

**Goals:**

- One declarative table of checks, so adding a tool is one entry plus one test.
- Checks testable without the real tools installed.
- Output readable on a plain terminal and in CI logs; no TUI.

**Non-Goals:**

- Minimum-version enforcement. Versions are reported, not judged; no current scaffold feature needs a floor.
- Parallel execution. Seven short subprocesses finish well under a second; ordering stays deterministic.
- Reusing doctor from `viber init`. `init` keeps its existing `steps.VerifyOpenspec`; unifying them is a later refactor.

## Decisions

### New package `internal/doctor`, thin command in `internal/cmd`

`internal/doctor` defines a `Check` (name, required flag, probe, remediation) and `Run(ctx, checks) []Result`. `internal/cmd/doctor.go` only parses flags, calls `Run`, renders, and maps the outcome onto the exit code.

*Alternative*: put the checks into `internal/cmd`. Rejected: it would mix process concerns with probing logic and make the checks hard to unit-test. It would also break the layering in `CLAUDE.md`, where `cmd` delegates work to packages.

### Probing through an injectable executor

Each probe runs `<tool> --version` (or `gh auth status`) via a package-level `execCommand = exec.CommandContext`. This is the same seam `internal/openspec` uses. A missing binary is detected with `errors.As(err, *exec.Error)` and becomes "not installed". Any other non-zero exit becomes "installed but not working", and the probe's first stderr line is shown.

*Alternative*: `exec.LookPath` only. Rejected: it cannot tell a working `gh` from an unauthenticated one, and it cannot report versions.

### Version extraction

The detected version is the first match of `\d+\.\d+(\.\d+)?` in the tool's `--version` output. When nothing matches, the first line is shown verbatim, trimmed to 60 characters. Tools format this line differently (`git version 2.x`, `gh version 2.x (date)`, bare `1.13.2`), so one tolerant pattern beats one parser per tool.

### Per-check timeout

Each probe runs under `context.WithTimeout(ctx, 5*time.Second)`, derived from the command context. This satisfies the spec's cancellation requirement and matches `CLAUDE.md`'s "keep long-running work cancellable". `gh auth status` can make a network call, and 5 s bounds it.

### Dependent check: gh authentication

The `gh auth` check declares `DependsOn: "gh"`. When `gh` failed, it is reported as `skipped`, not as a second failure (see the spec).

### Exit code via a sentinel error

When a required check fails, `doctor` renders all results and then returns `doctor.ErrRequiredFailed`. `internal/cmd` already maps any non-usage error to exit 1 and prints it. The message is the summary line, so the user sees it exactly once. Unknown flags and positional args use `usageArgs(cobra.NoArgs)` and exit 2, like the other commands.

### Output format

Human mode prints one aligned line per check: status, name, version or problem, and on a second indented line the remediation. Status markers are the plain words `ok`, `warn`, `error`, and `skip`, not symbols. That follows the no-emoji rule in `.claude/rules/ai-output-style.md` and keeps logs greppable. `--json` marshals the results slice. Human output goes to stdout. Errors from `Execute` still go to stderr, as today.

## Risks / Trade-offs

- [Remediation commands differ per OS (`brew` vs `apt`)] → For OS-specific tools (`gh`, `jq`), give the vendor's install URL. For npm-distributed tools (`openspec`, `markdownlint-cli2`), give the `npm install -g` command, which works everywhere.
- [`claude --version` output or binary name may change] → The check is optional and reports the raw first line on parse failure, so a format change degrades to less pretty output, not a false error.
- [`gh auth status` exits non-zero when any configured host is logged out, even if github.com is fine] → Scope the probe with `--hostname github.com`.
