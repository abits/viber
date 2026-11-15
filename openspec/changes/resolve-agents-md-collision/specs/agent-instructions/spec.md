# Spec Delta

## Purpose

Defines the instruction files a viber scaffold ships for coding agents, so that both Claude Code and other tools that follow the `AGENTS.md` convention get correct project guidance from the first session.

## ADDED Requirements

### Requirement: AGENTS.md carries agent-neutral project instructions

A scaffold SHALL contain an `AGENTS.md` at its root that states the project name and description, points to `intend.md` for problem, goals, and constraints, describes the OpenSpec workflow under `openspec/`, and names `make lint-md` for prose linting. `AGENTS.md` SHALL NOT describe Claude Code sub-agents, Claude Code plugins, slash commands, or Claude Code settings.

#### Scenario: Fresh scaffold

- **WHEN** the user runs `viber init myproj` with description "A demo"
- **THEN** `myproj/AGENTS.md` contains `myproj`, `A demo`, `intend.md`, `openspec/`, and `make lint-md`

#### Scenario: No Claude-specific content

- **WHEN** a scaffold is rendered from the embedded template set
- **THEN** its `AGENTS.md` mentions none of `.claude/`, `/plugin`, `@agent-`, `/repo-init`, `/sync-issues`, or `settings.json`

### Requirement: CLAUDE.md imports AGENTS.md

A scaffold's `CLAUDE.md` SHALL import `AGENTS.md` using Claude Code's `@AGENTS.md` import syntax on a line of its own, and SHALL NOT repeat the project description, the `intend.md` pointer, or the `make lint-md` instruction that `AGENTS.md` already carries.

#### Scenario: Import line present

- **WHEN** a scaffold is rendered from the embedded template set
- **THEN** its `CLAUDE.md` contains a line that is exactly `@AGENTS.md`

#### Scenario: No duplicated shared facts

- **WHEN** a scaffold is rendered from the embedded template set
- **THEN** its `CLAUDE.md` does not contain `make lint-md` and does not contain the project description

### Requirement: Team documentation lives in docs/team.md

A scaffold SHALL contain `docs/team.md` describing the Claude Code sub-agent team: the plugin marketplace prerequisite, the hand-off flow, how to invoke a role, agent teams mode, the roles table, and the project slash commands. `CLAUDE.md` SHALL point to `docs/team.md` for this information.

#### Scenario: Team file rendered

- **WHEN** the user runs `viber init myproj`
- **THEN** `myproj/docs/team.md` exists and lists all nine roles found in `myproj/.claude/agents/`

#### Scenario: CLAUDE.md points to the team file

- **WHEN** a scaffold is rendered from the embedded template set
- **THEN** its `CLAUDE.md` references `docs/team.md`

### Requirement: Team references point to docs/team.md

Every scaffold file that refers the reader to the team documentation, including each sub-agent definition under `.claude/agents/` and `README.md`, SHALL reference `docs/team.md`. No scaffold file other than `CLAUDE.md` SHALL reference `AGENTS.md` as the source of team or marketplace information.

#### Scenario: Agent definitions

- **WHEN** a scaffold is rendered from the embedded template set
- **THEN** every file under `.claude/agents/` that mentions the marketplace setup step references `docs/team.md` and none references `AGENTS.md`

#### Scenario: README

- **WHEN** a scaffold is rendered from the embedded template set
- **THEN** its `README.md` points to `docs/team.md` for the sub-agent team
