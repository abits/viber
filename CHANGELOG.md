# Changelog

All notable, user-visible changes land here. The README carries only the most recent entry; this file is the full history.

Entries follow the shape documented in `.claude/rules/release-readme.md`: a `## vX.Y.Z — <date>` heading, a link to the GitHub release, and 2–5 bullets derived from `git log vPREV..vX.Y.Z --oneline --no-merges`. Newest first; never retro-edit a shipped entry (mistakes get a follow-up release note).

## v0.4.3 — 2026-10-05

Release: <https://github.com/abits/viber/releases/tag/v0.4.3>.

- **Release-driven README rule.** `.claude/rules/release-readme.md` (shipped both in viber and in every scaffolded project) requires a `## What's new — vX.Y.Z` section at the top of README on every release.
- **VERSION reconciled** with the published tag so `make bump-patch` produces the correct next version.

## v0.4.2 — 2026-10-04

Release: <https://github.com/abits/viber/releases/tag/v0.4.2>.

- **`viber doctor` command added.** Runs health checks on the environment (openspec installation, Go toolchain, `gh` auth, etc.), reporting what's missing or misconfigured. Exits cleanly on interrupt; bounded child-process probes so a hanging check can't lock the whole run.
- Internal: planned and archived the OpenSpec change that drove the doctor feature.

## v0.4.1 — 2026-09-26

Release: <https://github.com/abits/viber/releases/tag/v0.4.1>.

- **Agent teams config in every scaffold, off by default.** `.claude/settings.json` now carries `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=0`, a `make lint-md` permission, and a `TaskCompleted` hook that blocks while `.env` is tracked in git. `AGENTS.md` explains when a team is worth its token cost and how to opt in.
- **Agent hand-offs use the documented `@agent-<role>` mention**, and each role has a `color` in its front matter.
- **Templates can no longer silently shadow each other.** A template set with both `foo` and `foo.tmpl` used to render whichever came last; `viber init` now fails and names both entries (`templates.ErrDuplicateTarget`).

## v0.4.0 — 2026-09-24

Release: <https://github.com/abits/viber/releases/tag/v0.4.0>.

- **viber is scaffolded with viber.** The repository now carries the same scaffold `viber init` generates (`intend.md`, `openspec/`, agents, rules, `/repo-init` and `/sync-issues`). `make dogfood` re-renders the template-owned files; CI's new `scaffold` job runs `make dogfood-check` and fails on drift.
- **`make lint-md` passes on a fresh scaffold.** `.markdownlint.yaml` lets front matter (`title`, `name`, `description`) stand in for the first-line H1, so generated agents, commands, and skills no longer fail MD041, and `lint-md` skips the files `openspec init` generates.
- **Scaffolded `CLAUDE.md` counts the agent team correctly** (nine, not seven).

## v0.3.1 — 2026-09-23

Release: <https://github.com/abits/viber/releases/tag/v0.3.1>.

- **Git + GitHub integration in every scaffolded project.** `viber init` now generates `scripts/repo-init.sh` (one-shot `gh repo create --private --push`) and `scripts/sync-issues.sh` (one-way mirror of `openspec/changes/*/tasks.md` → GitHub Issues). Both are surfaced as `make repo-init` / `make sync-issues` targets and as Claude Code slash commands `/repo-init` / `/sync-issues`.
- **OpenSpec tasks auto-sync.** The scaffold ships a `.claude/settings.json` with a `PostToolUse` hook that runs `sync-issues.sh --auto` in the background whenever Claude writes to a `tasks.md`, so `/opsx:apply` progress reflects into GitHub Issues without manual re-syncing.
- **Idempotent issue fingerprinting.** Each task is identified by `sha1(change-id|normalised-text)[:12]` embedded in the issue body. Renumbering tasks (`1.3 → 2.1`) doesn't churn issues. Deleting or renaming a task closes the orphaned issue automatically on the next sync.

## v0.3.0 — 2026-09-21

Release: <https://github.com/abits/viber/releases/tag/v0.3.0>.

- **`viber update` now verifies release archives** against `checksums.txt` before overwriting the binary — no unverified swaps of the installed executable.
- **`viber init` is atomic on a fresh destination**: rendering happens in a temp dir and is renamed into place only after every file is written, so a broken template or full disk leaves nothing behind. Merge behaviour (`--force` onto an existing dir) is unchanged.
- **`internal/ghfetch`** extracted so `--from` (template tarballs) and `viber update` (release assets) share one GitHub-tarball code path.
- **Six extra linters wired in** (`revive`, `godot`, `errorlint`, `bodyclose`, `gosec`, `misspell`) with per-symbol doc coverage across the codebase.
- **`v0.3.0` release CI green on Linux, macOS, and Windows** (linux/darwin/windows × amd64/arm64 archives + `checksums.txt`).
