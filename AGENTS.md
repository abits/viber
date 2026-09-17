# Agent team

This project ships a Claude Code sub-agent team covering the software development lifecycle. Each agent has a scoped responsibility, expected inputs, expected outputs, and explicit hand-off targets.

## Plugin marketplace prerequisite

Every role in this project can reach into the [`sethdford/claude-skills`](https://github.com/sethdford/claude-skills) marketplace — 454 standards-grounded skills and 173 commands across 9 roles. Add it once:

    /plugin marketplace add sethdford/claude-skills

Each agent file lists the plugins it expects and the exact `/plugin install …` line — fine-grained per plugin, or the whole-role shortcut via `claude install github:sethdford/claude-skills/<role>`. Grab everything (all 57 plugins across all 9 roles) with:

    claude install github:sethdford/claude-skills

The cross-role `sdlc-cross-role` plugin gives you lifecycle commands that span every role — install it separately regardless of which roles you use:

    /plugin install sdlc-cross-role

Commands from `sdlc-cross-role`: `/full-lifecycle`, `/feature-kickoff`, `/quality-gate`, `/pre-launch-review`, `/design-review`, `/security-review`, `/post-incident-review`, `/sprint-ceremony`.

## Flow

    intend.md
        ↓
    product-manager  →  proposal.md
        ↓                       ↘
    designer                 architect        →  design.md · ADRs · tasks.md
        ↓                       ↓
        └──────── tech-lead ────┴──── plan · standards · sequencing
                       ↓
    ┌──────────────────┬───────────────────┬──────────────────┐
    backend-engineer   frontend-engineer   devops-engineer
        ↓                   ↓                   ↓
    qa-engineer  ←─────────┴──────────────  security-reviewer
        ↓
    merge

## How to invoke

- `Task` tool with `subagent_type: <role>`
- `@<role>` shorthand in a Claude Code message (e.g. `@architect`)
- Free-form: mention the role and hand-off explicitly

## Agent teams mode (experimental)

`.claude/settings.json` sets `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`, so Claude Code can launch
roles named above as independent teammate sessions — each with its own context window, a shared
self-claiming task list, and direct inter-agent messaging — instead of ephemeral subagents.

Trade-offs to weigh before relying on it:

- Interactive sessions only; it has no effect under `-p`/headless runs (roles still spawn as
  ordinary subagents there).
- No session resumption: `/resume` and `/rewind` drop in-process teammates.
- Token cost scales linearly with team size — each teammate is a full session.
- One team per session, no nested teams, and per-teammate permissions are fixed at spawn time.

See <https://code.claude.com/docs/en/agent-teams.md>. Set the env var to `0` in
`.claude/settings.json` to fall back to ordinary subagent-only behavior.

## Roles

| Role              | File                                    | Focus                                                       |
|-------------------|-----------------------------------------|-------------------------------------------------------------|
| Product manager   | `.claude/agents/product-manager.md`     | Requirements, user stories, acceptance criteria             |
| Designer          | `.claude/agents/designer.md`            | Research, UX strategy, design system, prototypes, a11y      |
| Architect         | `.claude/agents/architect.md`           | System design, ADRs, tech choices, task breakdown           |
| Tech lead         | `.claude/agents/tech-lead.md`           | Delivery plan, standards, sequencing, cross-team RFCs       |
| Backend engineer  | `.claude/agents/backend-engineer.md`    | Server/API implementation + unit tests                      |
| Frontend engineer | `.claude/agents/frontend-engineer.md`   | UI implementation + component tests                         |
| DevOps engineer   | `.claude/agents/devops-engineer.md`     | CI/CD, IaC, release plumbing                                |
| Security reviewer | `.claude/agents/security-reviewer.md`   | Threat model + review before merge                          |
| QA engineer       | `.claude/agents/qa-engineer.md`         | Test plan + integration/E2E tests                           |
