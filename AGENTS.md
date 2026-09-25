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
- `@agent-<role>` shorthand in a Claude Code message (e.g. `@agent-architect`)
- Free-form: mention the role and hand-off explicitly

## Agent teams mode (experimental)

`.claude/settings.json` sets `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`, so Claude Code can launch
roles named above as independent teammate sessions — each with its own context window, a shared
self-claiming task list, and direct inter-agent messaging — instead of ephemeral subagents. The
same file also sets `subagentPromptCacheTtl: 1h` (in-process teammates otherwise fall back to a
5-minute cache window) and pre-approves `make lint-md` so a running team doesn't stall on repeated
permission prompts in the lead session; a `TaskCompleted` hook blocks marking a task done while
`.env` is tracked in git, as a last-resort gate on `.claude/rules/security.md`.

**Most of the flow above is a poor fit for a team.** Agent teams add coordination overhead and
cost significantly more tokens than a single session, and pay off only when the work is genuinely
independent. The PM → architect → tech-lead hand-offs are sequential and single-owner — invoke
them one at a time (`@agent-<role>`, above) rather than asking for a team. The two places a team is
worth its cost:

- **Implementation**: `backend-engineer`, `frontend-engineer`, and `devops-engineer` can own
  disjoint files and run in parallel. Ask explicitly, e.g. "spawn a team: backend-engineer,
  frontend-engineer, and devops-engineer to implement tasks.md in parallel."
- **Early, ambiguous exploration**: `product-manager`, `designer`, and `architect` debating an
  under-specified `intend.md` from different angles before anyone commits to a proposal.

Enabling agent teams also changes ordinary delegation: a subagent Claude names on its own now
launches as a teammate too, so a plain single-role hand-off can unexpectedly balloon into a full
team session. If that happens, tell Claude to spawn a subagent instead of a teammate, or set
`CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=0` for the session.

Further trade-offs to weigh:

- Interactive sessions only; it has no effect under `-p`/headless runs (roles still spawn as
  ordinary subagents there).
- No session resumption: `/resume` and `/rewind` drop in-process teammates.
- One team per session, no nested teams, and per-teammate permissions are fixed at spawn time.
- Display mode defaults to `in-process` (works in any terminal); split panes need tmux or iTerm2,
  which this scaffold doesn't assume are installed, so it's left unset rather than forced on.

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

## Slash commands

Two project-local slash commands wire this scaffold into GitHub. They're thin wrappers over the scripts in `scripts/`, so `make repo-init` / `make sync-issues` do the same thing from a shell.

- `/repo-init` — turns the local scaffold into a GitHub-hosted repo (`git init` if needed, initial commit if none, `gh repo create --private --push`). Idempotent: re-running is safe.
- `/sync-issues` — mirrors every `- [ ] N.M Task text` line under `openspec/changes/*/tasks.md` into a GitHub Issue and reconciles state (ticked → closed, un-ticked → reopened, deleted/renamed → orphan closed). One-way (tasks.md is source of truth). Labels every issue `openspec` + `openspec/<change-id>`.

`.claude/settings.json` also carries a `PostToolUse` hook that runs `sync-issues.sh --auto` in the background whenever `openspec/changes/*/tasks.md` is written, so `/opsx:apply` progress reflects into GitHub without manual re-syncing.

Prereqs for both: `gh` (authenticated via `gh auth login`), `jq`. See `README.md`.
