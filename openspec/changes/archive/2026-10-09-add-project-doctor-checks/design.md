# Design

## Context

`internal/doctor` models a check as a subprocess: `Check{Name, Args, Required, DependsOn, Remediation}`, run by `probe` through the package-level `execCommand` seam with a 5 s per-check timeout. `Run` walks the checks in order and marks a check `skipped` when its single `DependsOn` did not pass. `internal/cmd/doctor.go` renders the flat `[]Result` as text or JSON and returns `ErrRequiredFailed` when a required check failed.

Most project checks (see `specs/doctor/spec.md`) are not subprocesses: they read files or combine earlier results. Some do shell out to `git`. The issue sync check depends on tool checks (`jq`, `gh auth`) and on project checks (hook script, remote), so the two groups cannot run in isolation.

## Goals / Non-Goals

**Goals:**

- One check model and one runner for both groups, so dependencies cross groups without special cases.
- Tool-check output outside a scaffold stays byte-identical, guarded by a test.
- Every project check testable with a `t.TempDir` scaffold and the existing fake executor.

**Non-Goals:**

- Reading `.viber.yml` or comparing against templates. The manifest change will add a detection path and the drift check.
- Parsing hook commands as shell. Script references are found by pattern, not by evaluating the command.
- Supporting GitHub Enterprise remotes in the remote check.

## Decisions

### Generalize `Check` with a probe function

`Check` gains `Group` (`tools` or `project`) and `Probe func(ctx context.Context, prior Results) Result`. When `Probe` is nil, the existing `Args` path runs, so the seven tool checks keep their table entries unchanged. `prior` is a read-only lookup of results recorded so far, keyed by check name.

*Alternative*: a separate `ProjectCheck` type with its own runner. Rejected: the issue sync check needs tool results, so a second runner would have to receive the first runner's output anyway, and the renderer would have to merge two result types.

### Issue sync aggregates prior results instead of using `DependsOn`

The issue sync probe reads the results of `jq`, `gh auth`, the hook-script check, and the remote check from `prior` and reports the first one that did not pass, reusing its remediation. `DependsOn` stays a single string with its current meaning, which is to skip when a dependency failed.

*Alternative*: make `DependsOn` a `[]string`. Rejected after the spec settled: `DependsOn` produces `skipped`, but the spec wants a warning that names the cause. Skipping would hide the very fact the check exists to report.

### Detection lives in `internal/doctor`, called by `cmd`

`doctor.DetectProject(ctx, cwd) (root string, ok bool)` runs `git rev-parse --show-toplevel` through `execCommand` under the probe timeout. Any failure, whether git is missing or the directory is not a work tree, falls back to `cwd`. It then stats `CLAUDE.md`, `openspec/`, and `.claude/`. `runDoctor` appends `doctor.ProjectChecks(root, inGit)` to `DefaultChecks()` only when `ok`. Detection is not itself a check, so it never adds an output line.

*Alternative*: walk parent directories looking for `CLAUDE.md`. Rejected: a stray `CLAUDE.md` in `$HOME` would win over the repository's own root, and git already knows the answer.

### Checks without git report `skipped`

The remote and both `.env` checks need a work tree. `ProjectChecks` receives `inGit` from detection, and those probes return `skipped` when it is false. A synthetic `DependsOn` target would add an output line that no user asked for.

### Each check reads one thing

| Check | Source | Notes |
| --- | --- | --- |
| settings.json valid | `os.ReadFile` + `json.Unmarshal` | The error names the parse problem, with an offset from `*json.SyntaxError` when one is available. A missing file is an error too, because hooks cannot run. |
| intend.md present | `os.Stat` | Content is not inspected; see the spec's workflow exclusion. |
| hook scripts present | the parsed settings, every `hooks.*[].hooks[].command` | References matching `scripts/[A-Za-z0-9._/-]+` are resolved against the root. `DependsOn: "settings.json"`, so invalid settings skip this check rather than report twice. Pattern-based so remote template sets with other script names are covered. |
| origin on github.com | `git remote get-url origin` | Accepts `https://github.com/`, `git@github.com:`, `ssh://git@github.com/`. |
| issue sync active | `prior` | See above. |
| .env not tracked | `git ls-files --error-unmatch .env` | Exit 0 means tracked, which is the error. |
| .env ignored | `git check-ignore -q .env` | Works whether or not `.env` exists. `DependsOn: ".env not tracked"`, because a tracked file is never reported as ignored and the warning would repeat the error. |

### Rendering

Human mode prints a blank line and `Project  <root>` before the first `project` result, then the same per-line format. The summary counts both groups. JSON adds `group` to `jsonResult`. No heading is emitted when no project result exists, which keeps tool-only output byte-identical. A golden test captures today's output before the change and asserts it after.

## Risks / Trade-offs

- [The heuristic matches any repository with `CLAUDE.md`, `openspec/`, and `.claude/`, including ones viber did not create] → Every project check is either generic (settings, `.env`) or only a warning. The manifest change replaces the heuristic as the primary signal.
- [Script references hidden behind variables or `cd` gymnastics are missed] → The check can only produce false passes, never false failures. That is acceptable for a warning-level check.
- [GitHub Enterprise users get a remote warning although `gh` can sync] → Documented in the remediation text. The check stays optional.
- [Two extra git subprocesses lengthen doctor] → Each is a local read bounded by the existing per-check timeout. The total stays well under a second on a normal repository.
