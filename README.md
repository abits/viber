# viber

Scaffold a **vibe-coding** project wired for [Claude Code](https://claude.com/claude-code) + [OpenSpec](https://github.com/Fission-AI/OpenSpec), with a Claude Code sub-agent team covering the software development lifecycle.

One command turns an empty directory into a project you can hand to `/opsx:explore`: a filled-in scaffold, initialized git repo, an `intend.md` you fill to describe what you're building, and nine role-scoped sub-agents ready to hand off along the SDLC: PM → designer / architect → tech-lead → engineers → QA / security.

## What's new — v0.4.3

Released 2026-10-05. Full release: <https://github.com/abits/viber/releases/tag/v0.4.3>.

- **Release-driven README rule.** `.claude/rules/release-readme.md` (shipped both in viber and in every scaffolded project) requires a `## What's new — vX.Y.Z` section at the top of README on every release.
- **VERSION reconciled** with the published tag so `make bump-patch` produces the correct next version.

Earlier releases: see [CHANGELOG.md](CHANGELOG.md).

## Install

**From a GitHub release:**

    viber update           # if you already have viber installed
    # or download from https://github.com/abits/viber/releases

**From source:**

    git clone https://github.com/abits/viber
    cd viber
    make install           # copies to $HOME/bin/viber (0755, overwrites)

**With `go install`:**

    go install github.com/abits/viber/cmd/viber@latest

## Requirements

- [OpenSpec](https://github.com/Fission-AI/OpenSpec) — `npm install -g @fission-ai/openspec`
- [Claude Code](https://claude.com/claude-code) — for the `/opsx:explore` slash command and the sub-agent team.
- `git` on `PATH`.
- Optional, for the scaffold's GitHub scripts (`make repo-init`, `make sync-issues`): [GitHub CLI](https://cli.github.com) (`gh auth login`) and [jq](https://jqlang.github.io/jq/).

Run `viber doctor` to verify that all required and optional tools are installed.

## Quick start

Interactive:

    viber init

Positional (skips the wizard):

    viber init myproj                    # dir defaults to ./myproj
    viber init myproj ~/code/myproj      # explicit dir

Fully-flagged (CI-friendly):

    viber init --no-tui \
        --name=myproj \
        --desc="a spec-driven side project"

`init` checks for `openspec` before it writes anything, so a missing dependency never leaves a
half-created directory behind.

Then in the new project:

    cd myproj
    # fill in intend.md
    claude
    /opsx:explore

## What `init` creates

    myproj/
    ├── intend.md              # Problem / Goals / Constraints — fill this first
    ├── README.md
    ├── CLAUDE.md
    ├── AGENTS.md              # sub-agent team overview and flow
    ├── Makefile               # `make lint-md` / `repo-init` / `sync-issues`
    ├── .markdownlint.yaml
    ├── .gitignore
    ├── .git/                  # initialized
    ├── scripts/
    │   ├── repo-init.sh       # gh repo create --private --push
    │   └── sync-issues.sh     # openspec tasks.md -> GitHub Issues
    ├── .claude/settings.json  # tasks.md sync hook; agent teams (off by default)
    ├── .claude/commands/      # /repo-init, /sync-issues
    ├── .claude/agents/
    │   ├── architect.md
    │   ├── backend-engineer.md
    │   ├── designer.md
    │   ├── devops-engineer.md
    │   ├── frontend-engineer.md
    │   ├── product-manager.md
    │   ├── qa-engineer.md
    │   ├── security-reviewer.md
    │   └── tech-lead.md
    ├── .claude/rules/         # AI directives auto-loaded every turn
    └── openspec/              # created by `openspec init`

## Subcommands

| Command | Purpose |
| --- | --- |
| `viber doctor` | Check the required and optional tools viber and its scaffolds depend on; inside a scaffold also runs project checks (settings.json validity, hook scripts, origin remote, issue-sync wiring, `.env` safety). Prints remediation hints and exits non-zero on required failures. `--json` emits a flat array where each entry carries a `group` ("tools" or "project"). |
| `viber init [name] [dir]` | Initialize and populate a new agentic coding project (interactive by default). |
| `viber version` | Print version, commit, and build date. |
| `viber update` | Download the latest GitHub release, verify it against `checksums.txt`, and overwrite `~/bin/viber` (Linux/macOS). |
| `viber completion {bash,zsh,fish,powershell}` | Emit a shell completion script. |
| `viber help [command]` | Help for any command. |

Every command has full GNU-style help (`NAME · SYNOPSIS · DESCRIPTION · OPTIONS · EXAMPLES · ENVIRONMENT · EXIT STATUS`). Run `viber <cmd> --help`.

Exit codes: `0` success · `1` runtime error · `2` usage error.

## Remote template sets

The default templates are embedded in the binary. Point `--from` at any public GitHub repo to fetch alternative templates as a tarball:

    viber init myproj --from myorg/viber-templates@main

Optional: set `GITHUB_TOKEN` for private repos or higher rate limits.

A template set is rendered file by file: files ending in `.tmpl` are executed as Go `text/template` against the project name and description, with the suffix stripped; all other files are copied verbatim. Two entries that would land on the same path, such as `README.md` and `README.md.tmpl`, are rejected rather than one silently overwriting the other. Archives with absolute or `..` paths, or larger than 8 MiB uncompressed, are rejected too.

## Development

    make build          # -> bin/viber (ldflags-injected version)
    make test           # go test ./...
    make lint           # golangci-lint run ./...
    make fmt            # goimports -w .
    make tidy           # go mod tidy
    make man            # -> man/viber*.1
    make completions    # -> completions/viber.{bash,zsh,fish}
    make lint-md        # markdownlint-cli2 over **/*.md

### viber is scaffolded with viber

This repository carries the same scaffold `viber init` generates: `intend.md`, `openspec/`, the sub-agent team in `.claude/agents/` (see `AGENTS.md`), the `/repo-init` and `/sync-issues` commands, and their scripts. New work starts as an OpenSpec change (`/opsx:propose` or `/opsx:explore` in Claude Code), and its `tasks.md` is mirrored into GitHub Issues.

The template-owned files are generated, not edited in place:

    make dogfood        # re-render them from internal/templates/default/
    make dogfood-check  # fail if any of them drifted (runs in CI)

## Releasing

`master` only changes through pull requests, so a release is a PR followed by a tag:

1. On a branch, bump `VERSION`, prepend a `## vX.Y.Z — YYYY-MM-DD` entry to [CHANGELOG.md](CHANGELOG.md), and replace this README's `## What's new — vX.Y.Z` section with the same bullets; open a PR. See `.claude/rules/release-readme.md` for the exact shape.
2. Merge it once CI is green.
3. Tag the merge commit and push the tag, or create the release in the GitHub UI with tag `vX.Y.Z` targeting `master`:

       git fetch origin master
       git tag vX.Y.Z origin/master
       git push origin vX.Y.Z

`make bump-patch` / `bump-minor` / `bump-major` still bump, commit, and tag in one step for a local, direct-to-`master` workflow.

The `release` workflow (`.github/workflows/release.yml`) runs GoReleaser on tag push and publishes cross-compiled binaries (linux/darwin/windows × amd64/arm64) to GitHub Releases.

Local dry-run:

    make release        # requires GITHUB_TOKEN
    # or:
    goreleaser release --snapshot --clean --skip=publish

## License

MIT — see [LICENSE](LICENSE).
