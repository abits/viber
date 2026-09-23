#!/usr/bin/env bash
# repo-init.sh — turn this local scaffold into a GitHub-hosted repository.
#
# Steps: verify git+gh, init repo if needed, create initial commit if the
# tree has no commits yet, then `gh repo create --source . --push`. Safe to
# re-run: if origin already points at a GitHub repo, just pushes HEAD.
#
# Flags:
#   --name NAME           repo name (default: current directory name)
#   --desc DESC           description (default: first paragraph of README.md)
#   --visibility V        private (default) | public | internal
#   --public              shortcut for --visibility public
#   --remote NAME         remote name (default: origin)
#   --dry-run             print intended actions; make no changes
#   -h, --help            show this help

set -euo pipefail

VISIBILITY=private
REMOTE=origin
NAME=""
DESC=""
DRY_RUN=0

usage() {
  sed -n '2,/^set -euo/p' "$0" | sed 's/^# \{0,1\}//' | sed '/^set -euo/d'
  exit "${1:-0}"
}

while (( "$#" )); do
  case "$1" in
    --name)       NAME="$2"; shift 2 ;;
    --desc)       DESC="$2"; shift 2 ;;
    --visibility) VISIBILITY="$2"; shift 2 ;;
    --public)     VISIBILITY=public; shift ;;
    --remote)     REMOTE="$2"; shift 2 ;;
    --dry-run|-n) DRY_RUN=1; shift ;;
    -h|--help)    usage 0 ;;
    *)            echo "repo-init: unknown flag: $1" >&2; usage 2 ;;
  esac
done

case "$VISIBILITY" in
  private|public|internal) ;;
  *) echo "repo-init: --visibility must be private|public|internal (got $VISIBILITY)" >&2; exit 2 ;;
esac

command -v git >/dev/null || { echo "repo-init: git not installed" >&2; exit 1; }
command -v gh  >/dev/null || { echo "repo-init: gh not installed (https://cli.github.com)" >&2; exit 1; }
gh auth status >/dev/null 2>&1 || { echo "repo-init: gh not authenticated. Run: gh auth login" >&2; exit 1; }

run() {
  if (( DRY_RUN )); then
    printf 'DRY: %s\n' "$*"
  else
    "$@"
  fi
}

# Repo name defaults to current directory.
if [[ -z "$NAME" ]]; then
  NAME=$(basename "$PWD")
fi

# Description defaults to README.md's first non-blank paragraph (capped 350 chars).
if [[ -z "$DESC" && -f README.md ]]; then
  DESC=$(awk '
    BEGIN { buf="" }
    /^#/ { next }
    NF == 0 { if (buf != "") exit; next }
    { buf = (buf == "" ? $0 : buf " " $0) }
    END { print buf }
  ' README.md | cut -c1-350)
fi

# Ensure this is a git repo.
if ! git rev-parse --git-dir >/dev/null 2>&1; then
  echo "repo-init: initializing git repository"
  run git init --quiet
fi

# Create initial commit if the tree has no commits yet.
if ! git rev-parse HEAD >/dev/null 2>&1; then
  echo "repo-init: creating initial commit"
  run git add -A
  run git commit --quiet -m "chore: initial scaffold from viber"
fi

# If origin already exists and points at GitHub, just push.
if git remote get-url "$REMOTE" >/dev/null 2>&1; then
  ORIGIN_URL=$(git remote get-url "$REMOTE")
  if [[ "$ORIGIN_URL" == *"github.com"* ]]; then
    echo "repo-init: $REMOTE already points at $ORIGIN_URL — pushing"
    BRANCH=$(git rev-parse --abbrev-ref HEAD)
    run git push -u "$REMOTE" "$BRANCH"
    exit 0
  fi
  echo "repo-init: $REMOTE exists but is not a github.com URL ($ORIGIN_URL); aborting" >&2
  exit 1
fi

OWNER=$(gh api user --jq .login)
FULL="$OWNER/$NAME"
echo "repo-init: creating $FULL ($VISIBILITY)"

CREATE_ARGS=(repo create "$FULL" --source=. --remote="$REMOTE" --push "--$VISIBILITY")
if [[ -n "$DESC" ]]; then
  CREATE_ARGS+=(--description "$DESC")
fi

run gh "${CREATE_ARGS[@]}"

if (( DRY_RUN )); then
  echo "DRY: would print URL and next steps"
  exit 0
fi

URL=$(gh repo view "$FULL" --json url --jq .url)
echo
echo "Created: $URL"
echo
echo "Next steps:"
echo "  make sync-issues     # mirror openspec/changes/*/tasks.md into GitHub Issues"
