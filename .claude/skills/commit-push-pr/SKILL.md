---
name: commit-push-pr
description: Stage current changes, create a well-crafted commit, push the branch, and open a PR via gh. User-invoked workflow.
disable-model-invocation: true
---

Apply the rules in `.claude/rules/git-hygiene.md` and `.claude/rules/pre-commit-checklist.md` throughout. They govern branch policy, commit message shape, and what must be true before staging.

Run the full stage → commit → push → PR flow in one pass.

1. Inspect state in parallel: `git status`, `git diff` (staged + unstaged), `git log -5 --oneline` (to match commit style).
2. Stage the intended files by name (not `git add -A`). Skip anything sensitive (`.env`, credentials, tokens).
3. Create a NEW commit with a concise message (subject ≤ 70 chars, body focused on the WHY). Do not `--amend` unless the user explicitly asked.
4. If the current branch has no upstream, push with `git push -u origin HEAD`. Otherwise `git push`.
5. Create the PR with `gh pr create`, using a short title (< 70 chars) and a HEREDOC body with `## Summary` (bullets) and `## Test plan` (checklist).
6. Return the PR URL.

`$ARGUMENTS` may contain extra instructions (e.g., a specific commit message or PR title). If empty, derive both from the diff.

Never skip hooks (`--no-verify`) or force push. If a pre-commit hook fails, fix the underlying issue and create a new commit — never `--amend` to hide the failure.
