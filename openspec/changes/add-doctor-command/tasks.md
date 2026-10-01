# Tasks

## 1. Check engine (`internal/doctor`)

- [x] 1.1 Create `internal/doctor` with `Check`, `Result`, `Status` (`ok`, `warning`, `error`, `skipped`), and an injectable `execCommand`. Verify with `go build ./...` and doc comments on every exported identifier (`golangci-lint run ./...`).
- [x] 1.2 Implement the probe: run `<tool> --version` under a 5 s per-check timeout. Map a missing binary (`*exec.Error`) to "not installed", any other failure to "not working" with the first stderr line, and extract the version via the tolerant regex. Verify with table tests using a fake `execCommand`: present, missing, non-zero exit, timeout, unparsable version.
- [x] 1.3 Implement `Run(ctx, checks)`. It must honor `DependsOn`, so `gh auth` is `skipped` when `gh` failed, classify failures by `Required`, and stop promptly when `ctx` is cancelled. Verify with tests for the dependent-skip scenario and a cancelled context.
- [x] 1.4 Define the default check table: `git` and `openspec` required; `gh`, `gh auth status --hostname github.com`, `jq`, `markdownlint-cli2`, and `claude` optional. Each gets the remediation text from the spec. Verify with a test that asserts order, required flags, and that every entry has a non-empty remediation.

## 2. Command (`internal/cmd`)

- [x] 2.1 Add `newDoctorCmd` with `usageArgs(cobra.NoArgs)` and a `--json` flag, register it in `root.go`, and write `internal/cmd/docs/doctor.txt` in the GNU help format of the other commands. Verify with `viber doctor --help` and a test that `viber doctor extra` exits 2.
- [x] 2.2 Render the human output: aligned status/name/version lines, remediation on an indented line, and a summary line with error and warning counts. No emoji. Return `doctor.ErrRequiredFailed` when a required check failed. Verify with command tests on a fake executor: exit 0 with warnings only, exit 1 with a required failure, and the summary counts from the spec.
- [x] 2.3 Render `--json` output. Verify with a test that stdout parses as JSON, a missing `jq` yields status `warning` with a remediation, and the exit codes match human mode.

## 3. Docs

- [x] 3.1 Add `viber doctor` to the README "Subcommands" table and point the "Requirements" section to it. Verify with `make lint-md`.
- [x] 3.2 Regenerate the man pages (`make man`) and confirm `viber doctor` appears. Run `make dogfood-check` to confirm no template-owned file changed.

## 4. Integration

- [x] 4.1 Run the built binary here: `make build && ./bin/viber doctor; echo $?`. In this environment `gh` is present and authenticated, so `git`, `openspec`, `gh`, and `gh auth` should all be `ok`; exit 0. Then run `PATH=/nonexistent ./bin/viber doctor` and expect exit 1 with `git` and `openspec` as errors.
