# Design

## Context

See proposal.md - Why. Current coupling, observed in the code:

- `internal/templates/default/` is one flat set. Claude-only files sit beside neutral ones: `CLAUDE.md.tmpl`, `docs/team.md.tmpl`, `.claude/{agents,commands,rules,settings.json}`.
- `templates.Data` is `{Name, Description}` and `Render` takes `(fs, dst, data, force)`; paths are never computed, only `.tmpl` is stripped.
- `runInit` ends with `openspec.Init`, which runs a bare, interactive `openspec init`; `printNextSteps` hardcodes `claude` and `/opsx:explore`.
- `internal/doctor` hardcodes the `claude` tool check and the `CLAUDE.md` + `openspec` + `.claude` sentinel triple.
- `openspec init --tools <id>` is non-interactive and accepts `claude`, `codex`, `gemini`, `opencode` among others (checked with `openspec init --help`, openspec 2.0.0).
- Invariants to keep: atomic render, `RenderTemplates` before `GitInit`, exit codes 0/1/2 with `internal/cmd` as the only printer, `TestEmbeddedTemplatesUseEveryDataField`, `make dogfood-check`.

## Goals / Non-Goals

**Goals:**

- One embedded template set serves all platforms; no per-platform copies of neutral content.
- The default (`claude`) scaffold is byte-identical to today's.
- A new platform is a registry entry (plus slot files only if it needs new ones).
- Existing remote `--from` sets keep working unchanged.

**Non-Goals:**

- Selecting several platforms in one scaffold (see Decision 5).
- Translating Claude sub-agents, commands, or hooks into other tools' formats (Cursor rules, Gemini extensions, opencode agents). Those platforms get the neutral core and the manual `make`/script path.
- Persisting the platform choice in a file inside the scaffold.
- Changing OpenSpec's own generated files; OpenSpec writes them for the selected tool.

## Decisions

### 1. A leaf package `internal/platform` holds the registry

A `Profile` per platform: id, display name, CLI binary, OpenSpec tool id, explore command, whether the tool auto-loads rule files, and a `Slots map[string]string` (slot name to destination path; absent means omitted). A function resolves an id to a `Profile` or an error that lists supported ids. It imports nothing from viber, so `cmd`, `wizard`, `templates`, `openspec`, and `doctor` can all depend on it without breaking `cmd -> steps -> {templates, gitrepo, openspec}`.

Alternative: keep profiles inside `internal/templates`. Rejected because `doctor` and `wizard` need the list and should not import the renderer.

### 2. Slots instead of per-platform trees or templated file names

Files that differ by platform live under `_slots/<slot>/...` in the template set. The renderer rewrites a path `_slots/<slot>/rest` to `<profile.Slots[slot]>/rest` (or to `<profile.Slots[slot]>` for a single-file slot), and skips the entry when the slot is unmapped. The path rewrite happens in the existing walk, before the existing duplicate-target check, so collisions still fail the same way and atomic publish is untouched.

Initial slots and mappings:

| Slot | claude | gemini | codex / opencode |
|------|--------|--------|------------------|
| `entrypoint` (`CLAUDE.md`/`GEMINI.md` body, imports `@AGENTS.md`) | `CLAUDE.md` | `GEMINI.md` | omitted |
| `agents` | `.claude/agents` | omitted | omitted |
| `commands` | `.claude/commands` | omitted | omitted |
| `settings` | `.claude/settings.json` | omitted | omitted |
| `team` | `docs/team.md` | omitted | omitted |
| `rules` | `.claude/rules` | `docs/rules` | `docs/rules` |

`entrypoint` is one template whose output differs by platform data (display name, import line, an `{{if}}` around the Claude-only sections about OpenSpec commands, hooks, and the team). If that makes the template unreadable, split it into two slots, `entrypoint-claude` and `entrypoint-gemini`; this is an implementation-time call that does not change the specs.

Alternatives considered:

- Separate template trees per platform: duplicates neutral files and drifts. Rejected.
- `{{if}}`/`{{range}}` only, with the file name also templated: Go templates cannot omit a file from a name, so an extra "empty output means skip" rule would be needed, and file names with braces are hostile on Windows and in `git`. Rejected.
- A manifest file in the set mapping files to platforms: another format to maintain and validate. Rejected for v1.

A slot-less set (every remote set today) never triggers the rewrite, which is what gives requirement "Template sets without slots keep rendering as before" for free. The explicit-`--platform`-with-`--from` notice is printed by `internal/cmd` after `Resolve` (it needs to know whether the fetched set has `_slots/`).

### 3. Platform data reaches templates through `templates.Data`

`Data` gains the platform's display name, CLI command, explore command, and an `AutoLoadsRules` boolean, as flat fields or one embedded struct. Each is referenced by at least one template, so `TestEmbeddedTemplatesUseEveryDataField` keeps passing and keeps meaning something; any field not used by a template is not added. `README.md`, `intend.md`, `AGENTS.md` (rules list only when `AutoLoadsRules` is false), `Makefile` (lint ignore paths for the platform's generated OpenSpec files), and the `viber init` next steps use them.

Observed assumption: `Makefile.tmpl` currently ignores `.claude/skills/openspec-*/**` and `.claude/commands/opsx/**` from `markdownlint`. Other tools generate OpenSpec files elsewhere (for example `.codex/`, `.gemini/`, `.opencode/`); the platform profile therefore carries a list of lint-ignore globs, filled in by running `openspec init --tools <id>` into a temp directory during implementation.

### 4. `openspec init --tools <id>` replaces the interactive picker

`openspec.Init` takes the OpenSpec tool id and passes `--tools`. This is a deliberate behavior change (flagged **BREAKING** in the proposal): the scaffold and the generated OpenSpec commands must agree, and an interactive prompt in the middle of a non-interactive `viber init` was already awkward in CI. Extra tools remain available by running `openspec init --tools a,b` afterwards.

Alternative: pass `--tools` only when `--platform` is explicit and keep the picker otherwise. Rejected: two code paths and a default that can silently produce a mismatched scaffold.

### 5. Single platform per scaffold

Several platforms at once (a team mixing Claude and Codex) is plausible but multiplies slot collisions (two rules destinations, two entry files) and doctor detection ambiguity. The data model (a profile with a slot map) does not preclude it later; v1 does not accept a list.

### 6. Selection UX

`--platform` is a string flag validated early in `runInit`, before `Resolve` or any write, and reported with `UsageError` (exit 2, lists supported ids). The wizard gains a choice field; today every wizard field is free text, so a small select field type (arrow keys, default `claude`) is added next to `fields`. In non-interactive mode the default is `claude`.

### 7. Doctor

The `claude` check becomes `agent CLI`. Platform detection from an existing scaffold uses markers only (`.claude/` means claude; `GEMINI.md` means gemini; otherwise codex or opencode, indistinguishable), so no marker file is added to scaffolds; for the indistinguishable case the check accepts any supported CLI. Scaffold detection becomes `openspec/` + `AGENTS.md` (+ markers as above), and the three Claude-hook-dependent project checks are omitted when `.claude/` is absent. In `viber doctor --json` the tool entry `claude` is renamed `agent CLI`; consumers filtering on the old name must update (noted in the changelog).

## Risks / Trade-offs

- [Claude scaffold regresses while files move into `_slots/`] -> Before moving anything, record a manifest (path + SHA-256) of the rendered default scaffold as a golden test fixture; the refactor must keep it identical. `make dogfood-check` covers the dogfooded subset only.
- [Non-Claude users get less than Claude users (no sub-agents, no hook)] -> Intended and documented in `README.md`; the neutral core plus rules is useful on its own, and the slot map makes adding more per-platform files incremental.
- [Per-platform explore command and lint-ignore paths guessed wrong] -> Derive them from real `openspec init --tools <id>` output in a temp directory during implementation, and pin them with a test per profile.
- [Gemini `@AGENTS.md` import semantics] -> Verify against the Gemini CLI memory-import documentation during implementation (Context7 or the upstream docs); fall back to copying a pointer sentence if imports are not supported.
- [`doctor --json` check rename] -> Mention in CHANGELOG; the JSON is documented as keyed by check name only loosely, and no in-repo consumer exists.
- [Breaking `openspec init` interactivity] -> Called out in proposal, README, and changelog; escape hatch is the documented manual command.

## Migration Plan

1. Capture the golden manifest of today's default scaffold.
2. Introduce `internal/platform` and slot-aware rendering with only the `claude` profile; move files into `_slots/`; golden test and `make dogfood-check` stay green. This step is releasable on its own.
3. Add the other profiles, templated wording, `--platform`, wizard field, `openspec --tools`, next steps.
4. Update doctor.
5. Docs, then release with `CHANGELOG.md`/`README.md` entries per `.claude/rules/release-readme.md`.

Rollback: revert the release; existing scaffolds are not touched by viber after `init`, so nothing needs migrating on user machines.

## Open Questions

- Exact OpenSpec explore command strings and generated-file directories for `codex`, `gemini`, `opencode` (resolved empirically in task 1.2; changes only profile data).
- Whether to add `cursor` or `github-copilot` to the first batch: they read `AGENTS.md` or have their own rules directories, so they are straightforward to add as profiles once the registry exists. Deferred.
