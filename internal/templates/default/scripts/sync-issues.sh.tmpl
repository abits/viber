#!/usr/bin/env bash
# sync-issues.sh — mirror openspec/changes/*/tasks.md into GitHub Issues.
#
# One-way: tasks.md is the source of truth. Every `- [ ] N.M Task text`
# line becomes a GitHub Issue tagged with a stable fingerprint. Ticking
# the box closes the issue; un-ticking reopens it. Deleting or renaming a
# task closes its orphaned issue with a comment. Re-runs are idempotent.
#
# Flags:
#   --dry-run    print intended actions, make no changes
#   --auto       quiet mode for the PostToolUse hook; failures exit 0
#   --change ID  limit the sync to a single change directory
#   -h, --help   show this help
#
# Requires: git, gh (authenticated), jq. In --auto mode, missing prereqs
# or a non-GitHub origin cause a silent exit 0 so the hook is safe.

set -euo pipefail

DRY_RUN=0
AUTO=0
FILTER_CHANGE=""

usage() {
  sed -n '2,/^set -euo/p' "$0" | sed 's/^# \{0,1\}//' | sed '/^set -euo/d'
  exit "${1:-0}"
}

while (( "$#" )); do
  case "$1" in
    --dry-run|-n) DRY_RUN=1; shift ;;
    --auto)       AUTO=1; shift ;;
    --change)     FILTER_CHANGE="$2"; shift 2 ;;
    -h|--help)    usage 0 ;;
    *)            echo "sync-issues: unknown flag: $1" >&2; usage 2 ;;
  esac
done

log()  { (( AUTO )) || printf '%s\n' "$*"; }
warn() { (( AUTO )) || printf 'sync-issues: %s\n' "$*" >&2; }

# In --auto mode any hard failure is downgraded to a silent exit 0.
soft_exit() {
  if (( AUTO )); then exit 0; else exit "${1:-1}"; fi
}

for tool in git gh jq; do
  command -v "$tool" >/dev/null || { warn "$tool not installed"; soft_exit 1; }
done
gh auth status >/dev/null 2>&1 || { warn "gh not authenticated"; soft_exit 1; }

ORIGIN=$(git remote get-url origin 2>/dev/null || true)
[[ -z "$ORIGIN" ]]        && { warn "no origin remote set (run repo-init first)"; soft_exit 1; }
[[ "$ORIGIN" == *github.com* ]] || { warn "origin is not a github.com repo"; soft_exit 1; }

CHANGES_DIR="openspec/changes"
if [[ ! -d "$CHANGES_DIR" ]]; then
  log "no $CHANGES_DIR/ directory; nothing to sync"
  exit 0
fi

# Collect active tasks.md files (skip archive).
mapfile -t TASK_FILES < <(
  find "$CHANGES_DIR" -mindepth 2 -maxdepth 3 -name tasks.md -not -path "$CHANGES_DIR/archive/*" \
    | sort
)
if (( ${#TASK_FILES[@]} == 0 )); then
  log "no active tasks.md files under $CHANGES_DIR"
  exit 0
fi

# Optional --change filter.
if [[ -n "$FILTER_CHANGE" ]]; then
  FILTERED=()
  for f in "${TASK_FILES[@]}"; do
    if [[ "$f" == "$CHANGES_DIR/$FILTER_CHANGE/tasks.md" ]]; then
      FILTERED+=("$f")
    fi
  done
  TASK_FILES=("${FILTERED[@]}")
  if (( ${#TASK_FILES[@]} == 0 )); then
    warn "no tasks.md matched --change=$FILTER_CHANGE"
    soft_exit 1
  fi
fi

run() {
  if (( DRY_RUN )); then
    printf 'DRY: %s\n' "$*"
  else
    "$@" >/dev/null
  fi
}

sha_short() {
  # Portable 12-char sha1 prefix.
  if command -v sha1sum >/dev/null; then
    printf '%s' "$1" | sha1sum | cut -c1-12
  else
    printf '%s' "$1" | shasum | cut -c1-12
  fi
}

# Ensure the base label exists once.
gh label create openspec --color '0e8a16' --description 'Tracked via openspec' --force >/dev/null 2>&1 || :

# Fetch current state of every openspec-labelled issue (one API call).
STATE_JSON=$(gh issue list --label openspec --state all --limit 500 \
  --json number,title,body,state 2>/dev/null || echo '[]')

# id -> "number|state|title" lookup.
declare -A ISSUE_BY_ID
while IFS=$'\t' read -r ID NUM STATE TITLE; do
  [[ -z "$ID" ]] && continue
  ISSUE_BY_ID["$ID"]="$NUM|$STATE|$TITLE"
done < <(
  jq -r '
    .[] |
    (.body // "") as $b |
    ($b | capture("<!-- viber-task-id: (?<id>[0-9a-f]+) -->"; "g") | .id) as $id |
    "\($id)\t\(.number)\t\(.state | ascii_downcase)\t\(.title)"
  ' <<<"$STATE_JSON"
)

# Track which ids we saw this run (for orphan reconcile at the end).
declare -A SEEN_ID

created=0; closed=0; reopened=0; unchanged=0; orphaned=0

process_task() {
  local change="$1" done_flag="$2" text="$3"
  # Normalise: strip leading "N.M " numbering, collapse internal whitespace.
  local normalised
  normalised=$(printf '%s' "$text" \
    | sed -E 's/^[0-9]+(\.[0-9]+)*[[:space:]]+//' \
    | tr -s '[:space:]' ' ' \
    | sed -E 's/^ +//; s/ +$//')
  [[ -z "$normalised" ]] && return 0

  local id
  id=$(sha_short "$change|$normalised")
  SEEN_ID["$id"]=1

  local change_label="openspec/$change"
  gh label create "$change_label" --color 'c5def5' --description "openspec change $change" --force >/dev/null 2>&1 || :

  local title="$normalised"
  local body
  body=$(printf 'Tracked by viber from \x60openspec/changes/%s/tasks.md\x60.\n\n<!-- viber-task-id: %s -->\n' "$change" "$id")

  local record="${ISSUE_BY_ID[$id]:-}"
  local num="" state=""
  if [[ -n "$record" ]]; then
    IFS='|' read -r num state _ <<<"$record"
  fi

  if [[ "$done_flag" == " " ]]; then
    # Unchecked → issue should be open.
    case "$state" in
      "")     log "create #new  [$change] $title"
              run gh issue create --title "$title" --body "$body" --label openspec --label "$change_label"
              created=$((created+1)) ;;
      open)   unchanged=$((unchanged+1)) ;;
      closed) log "reopen #$num [$change] $title"
              run gh issue reopen "$num"
              reopened=$((reopened+1)) ;;
    esac
  else
    # Checked → issue should be closed (or never created).
    case "$state" in
      "")     unchanged=$((unchanged+1)) ;;
      open)   log "close  #$num [$change] $title"
              run gh issue close "$num" --comment "closed by viber sync-issues: task marked done in openspec/changes/$change/tasks.md"
              closed=$((closed+1)) ;;
      closed) unchanged=$((unchanged+1)) ;;
    esac
  fi
}

for f in "${TASK_FILES[@]}"; do
  # Extract the <change-id> from openspec/changes/<change-id>/tasks.md.
  change=$(dirname "$f")
  change=${change#"$CHANGES_DIR/"}

  while IFS= read -r line; do
    if [[ "$line" =~ ^[[:space:]]*-[[:space:]]*\[([\ xX])\][[:space:]]+(.+)$ ]]; then
      flag="${BASH_REMATCH[1]}"
      # Normalise `X` to `x`; space stays space.
      [[ "$flag" == "X" ]] && flag=x
      process_task "$change" "$flag" "${BASH_REMATCH[2]}"
    fi
  done < "$f"
done

# Orphan reconcile: any issue whose id we didn't see this run and is currently open → close it.
for id in "${!ISSUE_BY_ID[@]}"; do
  if [[ -z "${SEEN_ID[$id]:-}" ]]; then
    IFS='|' read -r num state title <<<"${ISSUE_BY_ID[$id]}"
    if [[ "$state" == "open" ]]; then
      log "orphan #$num $title"
      run gh issue close "$num" --comment "closed by viber sync-issues: task no longer present in openspec/"
      orphaned=$((orphaned+1))
    fi
  fi
done

# Summary.
if (( AUTO )); then
  # In --auto mode print only if something changed.
  if (( created + closed + reopened + orphaned > 0 )); then
    printf 'sync-issues: created=%d closed=%d reopened=%d orphaned=%d unchanged=%d\n' \
      "$created" "$closed" "$reopened" "$orphaned" "$unchanged"
  fi
else
  printf '\nSummary: created=%d closed=%d reopened=%d orphaned=%d unchanged=%d\n' \
    "$created" "$closed" "$reopened" "$orphaned" "$unchanged"
fi
