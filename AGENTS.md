# Agent team

This project ships a Claude Code sub-agent team covering the software development lifecycle. Each agent has a scoped responsibility, expected inputs, expected outputs, and explicit hand-off targets.

## Flow

    intend.md
        ↓
    product-manager  →  proposal.md
        ↓
    architect        →  design.md · ADRs · tasks.md
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

## Roles

| Role              | File                                    | Focus                                                    |
|-------------------|-----------------------------------------|----------------------------------------------------------|
| Product manager   | `.claude/agents/product-manager.md`     | Requirements, user stories, acceptance criteria          |
| Architect         | `.claude/agents/architect.md`           | System design, ADRs, tech choices, task breakdown        |
| Backend engineer  | `.claude/agents/backend-engineer.md`    | Server/API implementation + unit tests                   |
| Frontend engineer | `.claude/agents/frontend-engineer.md`   | UI implementation + component tests                      |
| DevOps engineer   | `.claude/agents/devops-engineer.md`     | CI/CD, IaC, release plumbing                             |
| Security reviewer | `.claude/agents/security-reviewer.md`   | Threat model + review before merge                       |
| QA engineer       | `.claude/agents/qa-engineer.md`         | Test plan + integration/E2E tests                        |
