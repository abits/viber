# Spec Delta

## MODIFIED Requirements

### Requirement: Doctor checks the scaffold's tool dependencies

`viber doctor` SHALL check each of the following and report one result per check, in this order: `git`, `openspec`, `gh`, `gh` authentication, `jq`, `markdownlint-cli2`, `agent CLI`. The `agent CLI` check SHALL pass when the CLI of the project's platform is on `PATH` (or, when the platform cannot be told apart from the scaffold, when the CLI of any supported platform is on `PATH`), and outside a scaffold when the CLI of any supported platform is on `PATH`; its line SHALL name the CLI that was found. It SHALL NOT modify any file, install anything, or require network access beyond what the checked tools themselves need to answer.

#### Scenario: All tools present

- **WHEN** every listed tool is on `PATH` and `gh` is authenticated
- **THEN** doctor prints one passing line per check, each including the tool's detected version where the tool reports one

#### Scenario: Doctor changes nothing

- **WHEN** doctor runs in any directory
- **THEN** no file in that directory or elsewhere is created, modified, or deleted

#### Scenario: Codex user without Claude

- **WHEN** only `codex` of the supported platform CLIs is on `PATH`, outside a scaffold
- **THEN** the `agent CLI` check passes and names `codex`

#### Scenario: Scaffold platform CLI missing

- **WHEN** doctor runs in a scaffold with `.claude/` and `claude` is not on `PATH`, although `codex` is
- **THEN** the `agent CLI` check is reported as a warning whose remediation points at the Claude Code installation page

### Requirement: Checks are classified as required or optional

`git` and `openspec` SHALL be required checks, because `viber init` cannot complete without them. `gh`, `gh` authentication, `jq`, `markdownlint-cli2`, and `agent CLI` SHALL be optional checks, because each only powers an individual scaffold feature. A failed required check SHALL be reported as an error; a failed optional check SHALL be reported as a warning.

#### Scenario: Optional tool missing

- **WHEN** `jq` is not on `PATH` and every required tool is present
- **THEN** the `jq` line is reported as a warning, and the remaining checks still run

#### Scenario: Required tool missing

- **WHEN** `openspec` is not on `PATH`
- **THEN** the `openspec` line is reported as an error, and the remaining checks still run

#### Scenario: gh installed but not authenticated

- **WHEN** `gh` is on `PATH` but `gh auth status` reports no login
- **THEN** the `gh` check passes and the `gh` authentication check is reported as a warning

#### Scenario: gh missing

- **WHEN** `gh` is not on `PATH`
- **THEN** the `gh` check is a warning and the `gh` authentication check is reported as skipped, not as a second failure

### Requirement: Doctor detects a scaffolded project

Doctor SHALL determine a project root: the git top-level directory when the working directory is inside a git work tree, otherwise the working directory. The project root SHALL count as a scaffold when it contains an `openspec` directory, an `AGENTS.md` file, and at least one platform entry marker: a `CLAUDE.md` file with a `.claude` directory (platform `claude`), or a `GEMINI.md` file (platform `gemini`), or neither (platform `codex` or `opencode`, which doctor cannot tell apart). Doctor SHALL run project checks only for a scaffold; outside one, its output SHALL be identical to doctor without project checks.

#### Scenario: Run from a subdirectory of a scaffold

- **WHEN** the user runs `viber doctor` in `myproj/internal/foo` and `myproj` is the git top level containing `AGENTS.md`, `CLAUDE.md`, `openspec/`, and `.claude/`
- **THEN** doctor runs the project checks against `myproj`

#### Scenario: Run outside a scaffold

- **WHEN** the user runs `viber doctor` in a directory whose project root lacks `openspec/` or `AGENTS.md`
- **THEN** doctor runs only the tool checks and prints no project heading or project entries

#### Scenario: Non-Claude scaffold

- **WHEN** the project root contains `AGENTS.md` and `openspec/` but no `CLAUDE.md` or `.claude/`
- **THEN** doctor treats it as a scaffold and runs the project checks

#### Scenario: Scaffold without git

- **WHEN** the working directory is not inside a git work tree and contains `AGENTS.md`, `CLAUDE.md`, `openspec/`, and `.claude/`
- **THEN** doctor treats the working directory as the project root and runs the project checks

### Requirement: Doctor checks that the scaffold's setup works

For a scaffold, doctor SHALL run these project checks after the tool checks, in this order: `.claude/settings.json` is valid JSON, `intend.md` is present, hook scripts are present, `origin` remote is on github.com, issue sync is active, `.env` is not tracked, `.env` is ignored. The settings check and the `.env`-tracked check SHALL be required; the others SHALL be optional. For a scaffold without a `.claude` directory the settings check, the hook-scripts check, and the issue-sync check SHALL be omitted entirely, since the hook that drives them is not part of that scaffold, and the remaining checks SHALL keep their order and classification. Doctor SHALL NOT report on workflow state such as the content of `intend.md` or the progress of OpenSpec changes.

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

#### Scenario: Codex scaffold omits Claude hook checks

- **WHEN** doctor runs in a scaffold that has no `.claude/` directory
- **THEN** no settings, hook-scripts, or issue-sync entry appears in the output, and the other project checks run
