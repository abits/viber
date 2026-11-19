# Proposal

## Why

viber scaffolds are wired exclusively for Claude Code: the entry file is `CLAUDE.md`, the sub-agent team, rules, slash commands and hooks live under `.claude/`, and `viber init` ends with "run `claude`". Projects driven by another coding agent (Codex CLI, Gemini CLI, opencode) get a scaffold full of files their tool never reads, plus prose that tells them to use a tool they do not have. OpenSpec itself already supports all of these tools (`openspec init --tools <id>`), and the previous change (`resolve-agents-md-collision`) made `AGENTS.md` agent-neutral; the remaining Claude coupling is in the template set and in a handful of Go call sites.

## What Changes

- Add a platform concept: `viber init --platform <id>` (and a wizard prompt) selects the LLM coding platform. Initial set: `claude` (default), `codex`, `gemini`, `opencode`. Adding a platform later is one registry entry plus, where needed, template slots.
- Make the scaffold platform-dependent: Claude-only artifacts (`CLAUDE.md`, `.claude/agents|commands|rules|settings.json`, the sub-agent team doc, the issue-sync hook) are emitted only for `claude`; other platforms receive the neutral core plus the files their tool reads (for example `GEMINI.md` importing `AGENTS.md`), with rules placed where `AGENTS.md` can point to them.
- Adapt the templates: Claude wording in `README.md`, `intend.md`, `Makefile`, and `docs/` becomes platform-aware (tool name, CLI command, OpenSpec explore command), and the embedded set is reorganised so platform-specific files live in reserved `_slots/` entries that each platform maps to a destination path (or omits).
- Run `openspec init --tools <openspec-id>` for the chosen platform instead of the interactive tool picker, so the OpenSpec command files match the scaffold.
- Print platform-correct next steps after `init`.
- Make `viber doctor` platform-aware: the `claude` tool check becomes an "agent CLI" check, scaffold detection no longer requires `.claude/`, and Claude-only project checks run only for scaffolds that have `.claude/settings.json`.
- **BREAKING**: `openspec init` is no longer interactive during `viber init`; the OpenSpec tool selection is derived from `--platform`. Users who previously ticked additional tools in that prompt must run `openspec init --tools ...` themselves afterwards. The default scaffold file tree is otherwise unchanged.

## Capabilities

### New Capabilities

- `llm-platforms`: which LLM coding platforms viber supports, how one is selected at init, what each platform's scaffold contains, how OpenSpec is initialised for it, and what the next-steps output says.
- `template-slots`: the contract a template set uses to declare platform-specific files (reserved `_slots/` directory), including backward compatibility for template sets, such as remote `--from` sets, that do not use it.

### Modified Capabilities

- `doctor`: the `claude` tool check generalises to an agent-CLI check; scaffold detection and the project checks stop assuming Claude Code.

## Impact

- Code: `internal/cmd/init.go` (flag, next steps), `internal/wizard` (answers, a select field), `internal/templates` (`Data`, slot-aware rendering, a platform registry), `internal/openspec` (`--tools`), `internal/steps` (unchanged order), `internal/doctor` (checks, detection), `internal/cmd/docs/*.txt`, `tools/render-template`.
- Templates: files under `internal/templates/default/.claude/**`, `CLAUDE.md.tmpl`, `docs/team.md.tmpl` move into `_slots/`; `README.md.tmpl`, `intend.md.tmpl`, `AGENTS.md.tmpl`, `Makefile.tmpl` gain platform conditionals. `TestEmbeddedTemplatesUseEveryDataField` keeps guarding `Data`.
- Dogfooding: `DOGFOOD_FILES` paths are unchanged because viber renders the `claude` platform; `make dogfood-check` must stay green.
- Ordering: this change adds requirements about `AGENTS.md`/`CLAUDE.md` content that build on the unarchived `resolve-agents-md-collision` change; archive that one first.
- Docs: `README.md` (flag, supported platforms), `CHANGELOG.md` at release time.
