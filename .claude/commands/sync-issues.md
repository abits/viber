---
description: Mirror openspec/changes/*/tasks.md into GitHub Issues (create, close, reopen, reconcile orphans).
---

Run `bash scripts/sync-issues.sh $ARGUMENTS` in the current project.

The script is one-way (tasks.md → GitHub Issues), idempotent, and safe to re-run. It:

- Creates an issue for every `- [ ] N.M Task text` line under a non-archived `openspec/changes/*/tasks.md`.
- Closes the corresponding issue when the box gets ticked.
- Reopens if the box is un-ticked.
- Closes any orphan issue (labelled `openspec` but whose task no longer exists in any `tasks.md`) with a comment.
- Labels every issue `openspec` + `openspec/<change-id>`.

After the script finishes:

- Parse the final `Summary: created=… closed=… reopened=… orphaned=… unchanged=…` line and print it back to the user.
- If `created > 0`, list up to five newly created issue URLs (`gh issue list --label openspec --state open --limit 5 --json url --jq '.[].url'`).
- If the script exited non-zero, surface the error verbatim; do not retry silently.

Common flags to forward via `$ARGUMENTS`:

- `--dry-run` to preview without touching GitHub.
- `--change <id>` to limit the sync to one change directory.
