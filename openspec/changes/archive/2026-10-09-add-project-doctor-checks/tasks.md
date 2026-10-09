# Tasks

## 1. Generalize the check model

- [x] 1.1 Add `Group` field (`"tools"` or `"project"`) to `Check` and `Result` in `internal/doctor/doctor.go`, set every entry in `DefaultChecks()` to `Group: "tools"`, extend `jsonResult` to serialize `group`, and update any existing golden/table tests so they declare the group explicitly. Verify: `go test ./internal/doctor/...` stays green and `TestDefaultChecks` asserts every row's `Group` is `"tools"`.
- [x] 1.2 Add a `Probe func(ctx context.Context, prior Results) Result` field to `Check`, introduce a `Results` lookup type keyed by `Name`, and dispatch to `Probe` from `probe()` when non-nil (falling back to the existing `Args` path otherwise). Verify: new unit test in `internal/doctor/doctor_test.go` runs a `Probe`-based `Check` against a fake `prior` and asserts the returned `Result` is recorded verbatim.

## 2. Scaffold detection

- [x] 2.1 Add `DetectProject(ctx context.Context, cwd string) (root string, inGit bool, ok bool)` in a new `internal/doctor/project.go`: run `git rev-parse --show-toplevel` through `execCommand` under `probeTimeout`, fall back to `cwd` on any failure, then stat `CLAUDE.md`, `openspec/`, and `.claude/` under that root. Detection must not emit a `Result`. Verify: new `internal/doctor/project_test.go` covers scaffold-at-root inside git, scaffold reached from a subdirectory, scaffold outside git (`inGit == false`, `ok == true`), and a directory missing one of the three sentinels (`ok == false`), all using injected `execCommand` and a `t.TempDir`.

## 3. Project check probes

- [x] 3.1 Implement a `settings.json valid` probe that reads `root/.claude/settings.json` and runs `json.Unmarshal`, reporting a byte offset from `*json.SyntaxError` when available and treating a missing file as an error. Verify: subtests in `internal/doctor/project_test.go` for valid JSON (ok), malformed JSON (error naming the offset), and missing file (error with remediation pointing at `.claude/settings.json`).
- [x] 3.2 Implement an `intend.md present` probe that only `os.Stat`s `root/intend.md`. Content is never inspected. Verify: subtests for present (ok) and absent (warning with remediation naming `intend.md`).
- [x] 3.3 Implement a `hook scripts present` probe (`DependsOn: ".claude/settings.json valid"`) that parses the already-read settings, extracts every `scripts/[A-Za-z0-9._/-]+` reference from `hooks.*[].hooks[].command`, resolves each against `root`, and stats it. Verify: subtests cover all-present (ok), one missing script (warning naming the path), and `DependsOn` skip when settings parsing failed.
- [x] 3.4 Implement an `origin on github.com` probe that shells out `git remote get-url origin` via `execCommand`, accepting `https://github.com/`, `git@github.com:`, and `ssh://git@github.com/` prefixes, and returns `StatusSkipped` when `inGit` is false. Verify: subtests per accepted URL form (ok), missing remote (warning with remediation `make repo-init`), non-github remote (warning), and no git work tree (skipped).
- [x] 3.5 Implement a `.env not tracked` probe that runs `git ls-files --error-unmatch .env` via `execCommand`: exit 0 means the file is tracked, which is reported as an error with remediation `git rm --cached .env`. Returns `StatusSkipped` when `inGit` is false. Verify: subtests for untracked (ok), tracked (error), and no git (skipped).
- [x] 3.6 Implement a `.env ignored` probe (`DependsOn: ".env not tracked"`) that runs `git check-ignore -q .env` via `execCommand`: non-zero exit is a warning. Returns `StatusSkipped` when `inGit` is false; also skipped by `DependsOn` when `.env` is tracked. Verify: subtests for ignored (ok), not ignored (warning), no git (skipped), tracked (skipped via `DependsOn`).
- [x] 3.7 Implement an `issue sync active` probe that aggregates results from `prior`, inspecting in order: `jq`, `gh authentication`, `hook scripts present`, `origin on github.com`. On the first non-ok prerequisite, return a warning that names it and reuses its `Remediation`; otherwise return ok. Verify: subtests for all-ok (passes), each prerequisite failing individually (warning names that prerequisite), and several missing at once (names the earliest in order).

## 4. Wire project checks into viber doctor

- [x] 4.1 Add `ProjectChecks(root string, inGit bool) []Check` in `internal/doctor/project.go` returning the seven project checks from Group 3 in spec order, with `Group: "project"`, correct `Required` flags (only `settings.json valid` and `.env not tracked` are required), and `DependsOn` edges as specified. Verify: table test in `project_test.go` asserts the count, order, groups, required flags, and dependency edges.
- [x] 4.2 In `internal/cmd/doctor.go`, call `doctor.DetectProject(ctx, cwd)` before building the check list and append `doctor.ProjectChecks(root, inGit)` to `doctor.DefaultChecks()` only when detection reports `ok`. Verify: `internal/cmd/doctor_test.go` runs the command against a `t.TempDir` scaffold (mocked `execCommand`) and asserts project rows are included; a non-scaffold run contains only the seven tool rows.

## 5. Rendering and documentation

- [x] 5.1 Before changing the human renderer, capture the current tool-only output as a golden file in `internal/cmd/testdata/`. Then update the renderer in `internal/cmd/doctor.go` so a run with any `project` result prints a blank line and `Project <root>` before the first project line, and the summary counts errors and warnings across both groups. Verify: the pre-change golden stays byte-identical for a non-scaffold run; a new golden captures the scaffold output.
- [x] 5.2 Extend `jsonResult` so the `--json` output includes `"group"` on every entry while remaining a single flat top-level array. Verify: new subtests in `internal/cmd/doctor_test.go` parse the JSON for a scaffold run and assert every tool entry has `"group":"tools"`, every project entry has `"group":"project"`, and the top level is still an array.
- [x] 5.3 Update `internal/cmd/docs/doctor.txt` (the `Long` help text) and the `## Subcommands` row for `viber doctor` in `README.md` to describe the project-check group and the new `group` field in JSON output. Verify: `./bin/viber doctor --help` prints the project-check description; `make dogfood-check` stays green (any shared template content is updated at the source).

## 6. End-to-end verification

- [x] 6.1 Add an integration test in `internal/cmd/doctor_test.go` that runs the Cobra `doctor` command through its `Execute` seam inside a `t.TempDir` scaffold with a mocked `execCommand` and asserts: exit 0 on a healthy scaffold, exit 1 on malformed `.claude/settings.json`, exit 1 on a tracked `.env`, and exit 0 when `.env` is present but not ignored (warning only).
