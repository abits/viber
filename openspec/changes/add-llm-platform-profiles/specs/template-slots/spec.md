# Spec Delta

## Purpose

Defines how a template set declares files that exist only for some LLM platforms, so one template set can serve several platforms and existing template sets keep working unchanged.

## ADDED Requirements

### Requirement: Platform-specific files live in reserved slots

A template set MAY contain a top-level `_slots/` directory. Each entry directly below it SHALL name a slot. For the selected platform, a slot SHALL be rendered to the destination path the platform maps that slot to; a slot the platform does not map SHALL be omitted without error. The `_slots/` directory itself SHALL never appear in the output.

#### Scenario: Mapped slot

- **WHEN** a template set has `_slots/rules/git-hygiene.md.tmpl` and the platform maps `rules` to `.claude/rules`
- **THEN** the output contains `.claude/rules/git-hygiene.md` and no `_slots/` path

#### Scenario: Unmapped slot

- **WHEN** the platform does not map the slot `agents`
- **THEN** nothing from `_slots/agents/` is rendered and rendering succeeds

#### Scenario: Slot that is a single file

- **WHEN** a template set has `_slots/entrypoint.md.tmpl` and the platform maps `entrypoint` to `GEMINI.md`
- **THEN** the output contains `GEMINI.md` rendered from that template

### Requirement: Entries outside slots are platform-independent paths

Files outside `_slots/` SHALL render to the same relative path for every platform, as before, with the `.tmpl` suffix removed and `.tmpl` files executed as templates.

#### Scenario: Neutral file

- **WHEN** a template set has `AGENTS.md.tmpl` outside `_slots/`
- **THEN** `AGENTS.md` is rendered for every platform

### Requirement: Template sets without slots keep rendering as before

A template set with no `_slots/` directory, which includes every remote set fetched with `--from` today, SHALL render exactly as it did before platforms existed, whatever platform is selected. `viber init` SHALL print a one-line notice when `--platform` is given explicitly together with `--from` and the fetched set has no `_slots/`, saying that the set is rendered as is.

#### Scenario: Legacy remote set

- **WHEN** the user runs `viber init myproj --from me/tpl --platform codex` and `me/tpl` has no `_slots/`
- **THEN** every file of the set is rendered at its own path and the notice is printed

### Requirement: Slot destinations cannot collide silently

When a slot destination equals the path of another rendered entry, rendering SHALL fail with the existing duplicate-target error naming both sources, and nothing SHALL be left behind on a fresh scaffold.

#### Scenario: Collision

- **WHEN** a platform maps a slot to `AGENTS.md` and the set also has `AGENTS.md.tmpl` outside `_slots/`
- **THEN** rendering fails with an error naming both source paths and no destination directory exists afterwards

### Requirement: Templates can read the selected platform

Every template SHALL be executed with the selected platform's display name, the command that starts its CLI, its OpenSpec explore command, and whether its tool auto-loads rule files, so wording can adapt without a separate template per platform. Every one of these values SHALL be used by at least one template in the embedded set.

#### Scenario: Wording differs by platform

- **WHEN** the same template renders once for `claude` and once for `codex`
- **THEN** the platform display name and CLI command in the output differ accordingly
