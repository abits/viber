# Tasks

## 1. Templates

- [x] 1.1 `git mv internal/templates/default/AGENTS.md.tmpl internal/templates/default/docs/team.md.tmpl`, change only its `#` heading, and verify `git diff -M --stat` shows a rename
- [x] 1.2 Write a new `internal/templates/default/AGENTS.md.tmpl` with name, description, `intend.md`, OpenSpec workflow, and `make lint-md`, free of Claude-specific terms; verify with the content test from 1.5
- [x] 1.3 Rewrite `CLAUDE.md.tmpl`: `@AGENTS.md` import line, Team section pointing to `docs/team.md`, drop the duplicated project, `intend.md`, and lint lines; verify with the content test from 1.5
- [x] 1.4 Point the marketplace line in all nine `.claude/agents/*.md.tmpl` and the team line in `README.md.tmpl` to `docs/team.md`; verify `grep -rn AGENTS.md internal/templates/default` matches only `CLAUDE.md.tmpl`
- [x] 1.5 Add `docs/team.md` to the layout list in `internal/templates/embed_test.go` and add a render-and-grep test covering every scenario in `specs/agent-instructions/spec.md`; verify `go test ./internal/templates/...` passes

## 2. viber's own repo (dogfooding)

- [x] 2.1 Swap `AGENTS.md` for `docs/team.md` in `DOGFOOD_FILES`, run `make dogfood`, and verify `make dogfood-check` passes and `docs/team.md` exists
- [x] 2.2 Rewrite the repo `AGENTS.md` as a superset of the rendered template that also carries viber's Layout, Invariants, and Conventions; verify `make lint-md` passes
- [x] 2.3 Rewrite the repo `CLAUDE.md` to import `@AGENTS.md`, keep only Claude-specific sections (hooks, Team, Skills; Dogfooding moves to `AGENTS.md` because it binds every agent), point Team to `docs/team.md`, and list `AGENTS.md` among the hand-merged supersets; verify a new Claude Code session in the repo shows the invariants (`/memory` lists `AGENTS.md`)
- [x] 2.4 Update `README.md`: layout tree (`AGENTS.md` as agent instructions, new `docs/team.md`) and the self-scaffold paragraph; verify `make lint-md` passes

## 3. Integration

- [x] 3.1 Run `go run ./cmd/viber init` into a temp directory and verify the rendered `AGENTS.md`, `CLAUDE.md`, and `docs/team.md` read correctly end to end, and that `make lint`, `make test`, and `make lint-md` pass on the branch

## Workflow follow-up

- Add a CHANGELOG/README release note with the manual migration for existing scaffolds when the next version is cut (per `.claude/rules/release-readme.md`).
- Remove idea K from `openspec/ideas.md` if PR #57 has merged by then.
- Archive the change after review with `/opsx:archive`.
