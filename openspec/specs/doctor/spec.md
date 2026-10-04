# doctor Specification

## Purpose

Lets a developer verify in one step that every external tool viber and its scaffolds depend on is present and usable, and tells them how to fix what is not.

## Requirements

### Requirement: Doctor checks the scaffold's tool dependencies

`viber doctor` SHALL check each of the following and report one result per check, in this order: `git`, `openspec`, `gh`, `gh` authentication, `jq`, `markdownlint-cli2`, `claude`. It SHALL NOT modify any file, install anything, or require network access beyond what the checked tools themselves need to answer.

#### Scenario: All tools present

- **WHEN** every listed tool is on `PATH` and `gh` is authenticated
- **THEN** doctor prints one passing line per check, each including the tool's detected version where the tool reports one

#### Scenario: Doctor changes nothing

- **WHEN** doctor runs in any directory
- **THEN** no file in that directory or elsewhere is created, modified, or deleted

### Requirement: Checks are classified as required or optional

`git` and `openspec` SHALL be required checks, because `viber init` cannot complete without them. `gh`, `gh` authentication, `jq`, `markdownlint-cli2`, and `claude` SHALL be optional checks, because each only powers an individual scaffold feature. A failed required check SHALL be reported as an error; a failed optional check SHALL be reported as a warning.

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

### Requirement: Every failed check names its fix

Each warning or error line SHALL include a concrete remediation: the command that installs the tool or resolves the problem (for example `npm install -g @fission-ai/openspec` or `gh auth login`), or a documentation URL when no single command applies.

#### Scenario: Remediation shown

- **WHEN** `markdownlint-cli2` is not on `PATH`
- **THEN** its warning line includes `npm install -g markdownlint-cli2`

### Requirement: Exit status reflects required checks only

Doctor SHALL exit `0` when every required check passes, even if optional checks produced warnings. It SHALL exit `1` when at least one required check fails. It SHALL exit `2` on a usage error, such as an unknown flag or a positional argument. A final summary line SHALL state the number of errors and warnings.

#### Scenario: Only warnings

- **WHEN** all required checks pass and two optional checks fail
- **THEN** doctor exits `0` and the summary reports 0 errors and 2 warnings

#### Scenario: Required failure

- **WHEN** `git` is not on `PATH`
- **THEN** doctor exits `1`

#### Scenario: Unexpected argument

- **WHEN** the user runs `viber doctor extra`
- **THEN** doctor exits `2` and prints its usage

### Requirement: Machine-readable output

With `--json`, doctor SHALL print a single JSON document to standard output instead of the human-readable lines. It SHALL contain one entry per check with at least the check name, whether it is required, its status (`ok`, `warning`, `error`, or `skipped`), the detected version when known, and the remediation when the check did not pass. Exit codes SHALL be the same as without `--json`.

#### Scenario: JSON output

- **WHEN** the user runs `viber doctor --json` and `jq` is missing
- **THEN** standard output is valid JSON, the `jq` entry has status `warning` and a non-empty remediation, and nothing else is printed to standard output

### Requirement: Doctor is cancellable

Doctor SHALL stop promptly when interrupted, and a single tool that hangs SHALL NOT block the command indefinitely.

#### Scenario: Hanging tool

- **WHEN** a checked tool does not return within a bounded time
- **THEN** that check is reported as failed with a timeout message, and doctor continues with the next check
