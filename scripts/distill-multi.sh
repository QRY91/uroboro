#!/usr/bin/env bash
set -euo pipefail

# distill-multi.sh — Run uroboro distill across multiple repos
# Concatenates git extracts from each repo + global uro extracts into one JSONL file

UROBORO="${UROBORO:-uroboro}"
DAYS=""
SINCE=""
CORRELATE=false
OUT_DIR="."
REPOS=()

usage() {
    cat >&2 <<EOF
Usage: $(basename "$0") [OPTIONS] REPO1 [REPO2 ...]

Options:
  --days N        Limit to last N days
  --since DATE    Limit to after date (2006-01-02)
  --correlate     Join git↔uro captures by ±30min window
  --out DIR       Output directory (default: current directory)
  --help          Show this help

Examples:
  $(basename "$0") --days 180 --correlate ~/projects/uroboro ~/projects/qryzone
  $(basename "$0") --since 2025-01-01 --out /tmp/style-data ~/projects/*
EOF
    exit 1
}

# Parse args
while [[ $# -gt 0 ]]; do
    case "$1" in
        --days)     DAYS="$2"; shift 2 ;;
        --since)    SINCE="$2"; shift 2 ;;
        --correlate) CORRELATE=true; shift ;;
        --out)      OUT_DIR="$2"; shift 2 ;;
        --help|-h)  usage ;;
        -*)         echo "Unknown option: $1" >&2; usage ;;
        *)          REPOS+=("$1"); shift ;;
    esac
done

if [[ ${#REPOS[@]} -eq 0 ]]; then
    echo "Error: at least one repo path required" >&2
    usage
fi

mkdir -p "$OUT_DIR"

DATE=$(date +%Y-%m-%d)
OUTFILE="$OUT_DIR/extracts-${DATE}.jsonl"
MANIFEST="$OUT_DIR/extract-manifest.json"

# Build common flags
TIME_FLAGS=()
[[ -n "$DAYS" ]]  && TIME_FLAGS+=(--days "$DAYS")
[[ -n "$SINCE" ]] && TIME_FLAGS+=(--since "$SINCE")

# Clear output file
> "$OUTFILE"

GIT_TOTAL=0
REPO_NAMES=()

# Extract git from each repo
for repo in "${REPOS[@]}"; do
    repo=$(realpath "$repo")
    name=$(basename "$repo")
    REPO_NAMES+=("$name")

    echo "Extracting git: $name..." >&2
    count=$($UROBORO distill --source git --repo "$repo" "${TIME_FLAGS[@]}" 2>/dev/null | tee -a "$OUTFILE" | wc -l)
    GIT_TOTAL=$((GIT_TOTAL + count))
    echo "  $count commits" >&2
done

# Extract uro captures (global, once)
echo "Extracting uro captures..." >&2
URO_COUNT=$($UROBORO distill --source uro "${TIME_FLAGS[@]}" 2>/dev/null | tee -a "$OUTFILE" | wc -l)
echo "  $URO_COUNT captures" >&2

# Correlate if requested (re-run with --correlate per repo)
CORRELATED=0
if $CORRELATE; then
    echo "Correlating..." >&2
    # Re-extract with --source all --correlate to get correlations
    # Overwrite output with correlated version
    > "$OUTFILE"
    GIT_TOTAL=0
    for repo in "${REPOS[@]}"; do
        repo=$(realpath "$repo")
        $UROBORO distill --source git --repo "$repo" "${TIME_FLAGS[@]}" 2>/dev/null >> "$OUTFILE"
        count=$($UROBORO distill --source git --repo "$repo" "${TIME_FLAGS[@]}" 2>/dev/null | wc -l)
        GIT_TOTAL=$((GIT_TOTAL + count))
    done
    # For uro: extract with correlation against each repo's git commits
    # Use the first repo for correlation (simplification for now)
    first_repo=$(realpath "${REPOS[0]}")
    URO_PART=$($UROBORO distill --source all --repo "$first_repo" "${TIME_FLAGS[@]}" --correlate 2>/dev/null)
    # Append only uro records (git records already written above)
    echo "$URO_PART" | grep '"source":"uroboro"' >> "$OUTFILE" || true
    URO_COUNT=$(echo "$URO_PART" | grep -c '"source":"uroboro"' || true)
    CORRELATED=$(echo "$URO_PART" | grep -c '"correlated_git_hash":"[^"]\+' || true)
    echo "  $CORRELATED correlated" >&2
fi

TOTAL_LINES=$(wc -l < "$OUTFILE")

# Write manifest
REPOS_JSON=$(printf '"%s",' "${REPO_NAMES[@]}" | sed 's/,$//')
cat > "$MANIFEST" <<EOF
{
  "date": "$DATE",
  "repos": [$REPOS_JSON],
  "git_count": $GIT_TOTAL,
  "uro_count": $URO_COUNT,
  "correlated_count": $CORRELATED,
  "total_records": $TOTAL_LINES,
  "days": ${DAYS:-null},
  "since": ${SINCE:+\"$SINCE\"}${SINCE:-null},
  "output": "$OUTFILE"
}
EOF

echo "" >&2
echo "Done: $TOTAL_LINES records ($GIT_TOTAL git + $URO_COUNT uro)" >&2
echo "Output: $OUTFILE" >&2
echo "Manifest: $MANIFEST" >&2
