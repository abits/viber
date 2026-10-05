## Release — README announcement

Every published GitHub release has a matching `## What's new — vX.Y.Z` section at the top of `README.md`, directly under the intro and above the previous release's section (most recent first).

- Add the section **before** the tag is pushed. If the tag has already shipped without one, add it in a follow-up `docs: announce vX.Y.Z in README` commit immediately after.
- The section includes: release date (`Released YYYY-MM-DD.`), a link to the release page (`<https://github.com/<owner>/<repo>/releases/tag/vX.Y.Z>`), and 2–5 bullets covering user-visible changes.
- Derive the bullets from `git log vPREV..vX.Y.Z --oneline --no-merges`. Skip bumps, pure refactors, and dependency chores unless they fix something the user noticed.
- Keep each bullet one sentence. Lead with the behaviour (`viber update now verifies checksums`), not the file (`updater/updater.go changed`).
- Never retro-edit a previous release's bullets. Mistakes get a follow-up release note, not an in-place rewrite.
