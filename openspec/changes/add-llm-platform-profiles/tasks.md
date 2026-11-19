# Tasks

## 1. Safety net and platform registry

- [ ] 1.1 Add a golden test that renders the embedded default set for fixed `Name`/`Description` and compares a path + SHA-256 manifest checked in under `internal/templates/testdata/`; generate the fixture from the current tree before any template moves; verify `go test ./internal/templates/...` passes
- [ ] 1.2 Run `openspec init --tools <id>` in a temp directory for `claude`, `codex`, `gemini`, `opencode` and record each tool's explore command and generated-file directories in the design notes; verify by listing the produced files
- [ ] 1.3 Create `internal/platform` with the `Profile` type, the four profiles, and a lookup that returns an error listing the supported ids; verify table-driven tests cover every id, an unknown id, and that every profile's OpenSpec tool id is in the `openspec init --help` list
- [ ] 1.4 Document the package and the "add a platform" steps in `docs/` or the package doc comment; verify `golangci-lint run ./...` accepts the exported doc comments

## 2. Slot-aware rendering

- [ ] 2.1 Extend `templates.Render` to take the selected profile and rewrite `_slots/<slot>/...` paths (mapped to the destination, unmapped skipped, `_slots` never emitted) before the duplicate-target check; verify tests for mapped, unmapped, single-file slot, neutral file, and collision cases (collision leaves no destination)
- [ ] 2.2 Verify a set with no `_slots/` renders identically for every profile (test using an in-memory `fstest.MapFS`, including a legacy `.claude/foo.md` entry)
- [ ] 2.3 Update `tools/render-template` to accept `-platform` (default `claude`); verify `go run ./tools/render-template -name x -desc y <dir>` output matches the golden manifest
- [ ] 2.4 Move the Claude-only files into `_slots/` (`entrypoint`, `agents`, `commands`, `settings`, `team`, `rules`) with the `claude` mapping; verify the golden test from 1.1 and `make dogfood-check` stay green with no manifest change

## 3. Platform-aware templates

- [ ] 3.1 Add the platform fields to `templates.Data` and use each in at least one template; verify `TestEmbeddedTemplatesUseEveryDataField` passes and fails when a field is deliberately left unused
- [ ] 3.2 Adapt `README.md.tmpl`, `intend.md.tmpl`, `Makefile.tmpl` (lint-ignore globs), and `AGENTS.md.tmpl` (rules list only when the tool does not auto-load rules); verify rendered output per platform in tests (README tool and command, no `claude` for gemini, rules named in `AGENTS.md` for codex, no `.claude/` in `AGENTS.md` for claude)
- [ ] 3.3 Add the `gemini` entry file content with a line that is exactly `@AGENTS.md`; verify the import semantics against the Gemini CLI documentation and record the source in the design notes
- [ ] 3.4 Add a test that renders each non-Claude platform and asserts no `CLAUDE.md`, `.claude/`, `docs/team.md`, or the text `Claude Code`; verify it passes for `codex`, `gemini`, `opencode`
- [ ] 3.5 Make the README state that issue sync is manual on platforms without hooks; verify the codex scaffold has `scripts/sync-issues.sh` and no settings file

## 4. `viber init` wiring

- [ ] 4.1 Add `--platform` to `init` and `wizard.Answers`, validated first in `runInit` via `UsageError` (exit 2, lists ids, nothing created); verify tests for default, explicit, unknown platform, and that no directory exists after a usage error
- [ ] 4.2 Add a select field type to the wizard with `claude` preselected and ask for the platform when the flag is absent; verify wizard unit tests for default, change, and flag-supplied (prompt skipped)
- [ ] 4.3 Pass the OpenSpec tool id to `openspec.Init` as `--tools <id>`; verify with the existing command-stub test that the argument list contains `--tools codex` for `--platform codex`, and that a failure error names `openspec init`
- [ ] 4.4 Print platform-specific next steps and the one-line notice for `--from` sets without `_slots/`; verify output tests for each platform and for the legacy remote set
- [ ] 4.5 Update `internal/cmd/docs/init.txt` (flag, platforms, BREAKING note), regenerate the man page, and add the flag and platform list to `README.md`; verify `go test ./internal/cmd/...` (including the help test) passes

## 5. Doctor

- [ ] 5.1 Replace the `claude` check with `agent CLI` (platform-detected, any-supported fallback, names the CLI found, per-platform remediation); verify tests for the four scenarios in the doctor delta (codex-only, scaffold CLI missing, found, none)
- [ ] 5.2 Change scaffold detection to `openspec/` + `AGENTS.md` with the platform markers, and omit the settings, hook-scripts, and issue-sync checks when `.claude/` is absent; verify tests with a claude fixture, a gemini fixture, and a codex fixture, and update `testdata/doctor-*.txt`
- [ ] 5.3 Update `internal/cmd/docs/doctor.txt`; verify `go test ./internal/doctor/... ./internal/cmd/...` passes

## 6. Integration and release notes

- [ ] 6.1 End-to-end: for each platform run `viber init --no-tui` into a temp directory with a stubbed `openspec`, then `viber doctor` against the result; verify the tree matches the specs and doctor reports no errors for the scaffold files
- [ ] 6.2 Run `openspec validate add-llm-platform-profiles`, `golangci-lint run ./...`, `go test ./...`, `make lint-md`, and `make dogfood-check`; verify all are green
- [ ] 6.3 Add the CHANGELOG entry and the README "What's new" section per `.claude/rules/release-readme.md` (call out the `openspec init` interactivity change and the `agent CLI` rename); verify the README section and CHANGELOG bullets are the same text

## Workflow follow-up

- Archive `resolve-agents-md-collision` first, then this change after review.
- Verify the archived result with `openspec list --specs`.
