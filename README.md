# viber

Scaffold a **vibe-coding** project wired for [Claude Code](https://claude.com/claude-code) + [OpenSpec](https://github.com/Fission-AI/OpenSpec), with a Claude Code sub-agent team covering the software development lifecycle.

One command turns an empty directory into a project you can hand to `/opsx:explore`: a filled-in scaffold, initialized git repo, an `intend.md` you fill to describe what you're building, and nine role-scoped sub-agents ready to hand off along the SDLC: PM → designer / architect → tech-lead → engineers → QA / security.

## What's new — v0.3.0

Released 2026-09-21. Full release: <https://github.com/abits/viber/releases/tag/v0.3.0>.

- **`viber update` now verifies release archives** against `checksums.txt` before overwriting the binary — no unverified swaps of the installed executable.
- **`viber init` is atomic on a fresh destination**: rendering happens in a temp dir and is renamed into place only after every file is written, so a broken template or full disk leaves nothing behind. Merge behaviour (`--force` onto an existing dir) is unchanged.
- **`internal/ghfetch`** extracted so `--from` (template tarballs) and `viber update` (release assets) share one GitHub-tarball code path.
- **Six extra linters wired in** (`revive`, `godot`, `errorlint`, `bodyclose`, `gosec`, `misspell`) with per-symbol doc coverage across the codebase.
- **`v0.3.0` release CI green on Linux, macOS, and Windows** (linux/darwin/windows × amd64/arm64 archives + `checksums.txt`).

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
    ├── Makefile               # `make lint-md`
    ├── .markdownlint.yaml
    ├── .gitignore
    ├── .git/                  # initialized
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
    ├── .claude/rules/          # AI directives auto-loaded every turn
    └── openspec/              # created by `openspec init`

## Subcommands

| Command | Purpose |
|---|---|
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

## Development

    make build          # -> bin/viber (ldflags-injected version)
    make test           # go test ./...
    make lint           # golangci-lint run ./...
    make fmt            # goimports -w .
    make tidy           # go mod tidy
    make man            # -> man/viber*.1
    make completions    # -> completions/viber.{bash,zsh,fish}

## Releasing

    make bump-patch     # or bump-minor / bump-major — commits + tags vX.Y.Z
    git push --follow-tags

The `release` workflow (`.github/workflows/release.yml`) runs GoReleaser on tag push and publishes cross-compiled binaries (linux/darwin/windows × amd64/arm64) to GitHub Releases.

Local dry-run:

    make release        # requires GITHUB_TOKEN
    # or:
    goreleaser release --snapshot --clean --skip=publish

## License

MIT — see [LICENSE](LICENSE).
