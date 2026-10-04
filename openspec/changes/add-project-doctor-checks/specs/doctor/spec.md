# Spec Delta

## ADDED Requirements

### Requirement: Doctor detects a scaffolded project

Doctor SHALL determine a project root: the git top-level directory when the working directory is inside a git work tree, otherwise the working directory. The project root SHALL count as a scaffold when it contains a `CLAUDE.md` file, an `openspec` directory, and a `.claude` directory. Doctor SHALL run project checks only for a scaffold; outside one, its output SHALL be identical to doctor without project checks.

#### Scenario: Run from a subdirectory of a scaffold

- **WHEN** the user runs `viber doctor` in `myproj/internal/foo` and `myproj` is the git top level containing `CLAUDE.md`, `openspec/`, and `.claude/`
- **THEN** doctor runs the project checks against `myproj`

#### Scenario: Run outside a scaffold

- **WHEN** the user runs `viber doctor` in a directory whose project root lacks any of `CLAUDE.md`, `openspec/`, or `.claude/`
- **THEN** doctor runs only the tool checks and prints no project heading or project entries

#### Scenario: Scaffold without git

- **WHEN** the working directory is not inside a git work tree and contains `CLAUDE.md`, `openspec/`, and `.claude/`
- **THEN** doctor treats the working directory as the project root and runs the project checks

### Requirement: Doctor checks that the scaffold's setup works

For a scaffold, doctor SHALL run these project checks after the tool checks, in this order: `.claude/settings.json` is valid JSON, `intend.md` is present, hook scripts are present, `origin` remote is on github.com, issue sync is active, `.env` is not tracked, `.env` is ignored. The settings check and the `.env`-tracked check SHALL be required; the others SHALL be optional. Doctor SHALL NOT report on workflow state such as the content of `intend.md` or the progress of OpenSpec changes.

#### Scenario: Healthy scaffold

- **WHEN** doctor runs in a scaffold with valid settings, `intend.md`, the hook scripts, a github.com `origin`, an authenticated `gh`, `jq` installed, and an ignored, untracked `.env`
- **THEN** every project check passes

#### Scenario: Invalid settings file

- **WHEN** `.claude/settings.json` is not valid JSON
- **THEN** the settings check is reported as an error naming the parse problem, and doctor exits `1`

#### Scenario: Tracked .env

- **WHEN** `.env` is tracked by git
- **THEN** the `.env`-tracked check is reported as an error whose remediation is `git rm --cached .env`, and doctor exits `1`

#### Scenario: .env not ignored

- **WHEN** `.env` is neither tracked nor matched by any git ignore rule
- **THEN** the `.env`-ignored check is reported as a warning, and doctor exits `0` if no required check failed

#### Scenario: Missing hook script

- **WHEN** a hook command in `.claude/settings.json` references `scripts/sync-issues.sh` and that file does not exist
- **THEN** the hook-scripts check is reported as a warning naming the missing path

#### Scenario: No GitHub remote

- **WHEN** the repository has no `origin` remote, or `origin` does not point to github.com
- **THEN** the remote check is reported as a warning whose remediation is `make repo-init`

#### Scenario: Unfilled intend.md is not a failure

- **WHEN** `intend.md` exists but still contains only the template's placeholder comments
- **THEN** the `intend.md` check passes

#### Scenario: Git checks skipped without a work tree

- **WHEN** the project root is not inside a git work tree
- **THEN** the remote, `.env`-tracked, and `.env`-ignored checks are reported as skipped

### Requirement: Issue sync check names its first missing prerequisite

The issue sync check SHALL pass only when `jq` is installed, `gh` is authenticated, the hook script is present, and the `origin` remote is on github.com. When any prerequisite did not pass, the check SHALL be reported as a warning that names the first failed prerequisite in that order and reuses that prerequisite's remediation, rather than repeating the prerequisite's own failure.

#### Scenario: gh not authenticated

- **WHEN** `jq` is installed, the hook script and a github.com `origin` exist, and `gh` is not authenticated
- **THEN** the issue sync check is a warning stating that it is inactive because `gh` authentication failed, with remediation `gh auth login`

#### Scenario: Several prerequisites missing

- **WHEN** `jq` is missing and there is no `origin` remote
- **THEN** the issue sync check names `jq` as the cause

### Requirement: Project checks are reported as their own group

In human-readable output, doctor SHALL print the project checks after the tool checks under a heading that names the project root. The final summary line SHALL count errors and warnings across both groups.

#### Scenario: Grouped output

- **WHEN** doctor runs in a scaffold at `/home/me/myproj`
- **THEN** a heading containing `/home/me/myproj` separates the tool lines from the project lines, and a single summary line follows the project lines

## MODIFIED Requirements

### Requirement: Machine-readable output

With `--json`, doctor SHALL print a single JSON document to standard output instead of the human-readable lines. It SHALL be a flat array containing one entry per check with at least the check name, the group it belongs to (`tools` or `project`), whether it is required, its status (`ok`, `warning`, `error`, or `skipped`), the detected version when known, and the remediation when the check did not pass. Exit codes SHALL be the same as without `--json`.

#### Scenario: JSON output

- **WHEN** the user runs `viber doctor --json` and `jq` is missing
- **THEN** standard output is valid JSON, the `jq` entry has status `warning` and a non-empty remediation, and nothing else is printed to standard output

#### Scenario: JSON output in a scaffold

- **WHEN** the user runs `viber doctor --json` inside a scaffold
- **THEN** every tool entry has group `tools`, every project entry has group `project`, and all entries are elements of the same top-level array

### Requirement: Doctor is cancellable

Doctor SHALL stop promptly when interrupted, and a single check that hangs, whether a tool probe or a git query made by a project check, SHALL NOT block the command indefinitely.

#### Scenario: Hanging tool

- **WHEN** a checked tool does not return within a bounded time
- **THEN** that check is reported as failed with a timeout message, and doctor continues with the next check

#### Scenario: Hanging git query

- **WHEN** a git query made by a project check does not return within a bounded time
- **THEN** that project check is reported as failed with a timeout message, and doctor continues with the next check
