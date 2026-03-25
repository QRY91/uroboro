#!/bin/bash
# uroboro enforcement hook: SessionStart
# Injects uroboro reminder + project-specific tag conventions into agent context.
# stdout from SessionStart hooks becomes agent context.

INPUT=$(cat)
CWD=$(echo "$INPUT" | jq -r '.cwd')
SOURCE=$(echo "$INPUT" | jq -r '.source')
PROJECT=$(basename "$CWD")

# Always output base uroboro reminder
cat <<'REMINDER'
UROBORO ACTIVE. Use uro_capture, uro_decision, uro_blocker, uro_question throughout this session. Capture decisions silently — don't announce them.
REMINDER

# Detect whether CWD is itself a git repo
IS_GIT_REPO=false
if git -C "$CWD" rev-parse --show-toplevel >/dev/null 2>&1; then
  REPO_ROOT=$(git -C "$CWD" rev-parse --show-toplevel 2>/dev/null)
  # Reject dotfiles repo at $HOME
  if [ "$REPO_ROOT" != "$HOME" ]; then
    IS_GIT_REPO=true
  fi
fi

if [ "$IS_GIT_REPO" = "true" ]; then
  # Single-repo session: inject this repo's tag convention if present
  TAGS_FILE="$CWD/.claude/uroboro.tags"
  if [ -f "$TAGS_FILE" ]; then
    echo ""
    echo "PROJECT CAPTURE CONVENTION ($PROJECT):"
    cat "$TAGS_FILE"
    echo ""
    echo "Follow this tagging scheme for all uro_capture calls in this session."
  fi
else
  # Multi-repo parent session: scan immediate subdirs for tag conventions
  FOUND_REPOS=()
  for subdir in "$CWD"/*/; do
    [ -d "$subdir" ] || continue
    if git -C "$subdir" rev-parse --show-toplevel >/dev/null 2>&1; then
      TAGS_FILE="$subdir/.claude/uroboro.tags"
      if [ -f "$TAGS_FILE" ]; then
        FOUND_REPOS+=("$subdir")
      fi
    fi
  done

  if [ ${#FOUND_REPOS[@]} -gt 0 ]; then
    echo ""
    echo "MULTI-REPO SESSION: working directory contains multiple git repos."
    echo "Pass 'project' explicitly on every uro_capture/uro_decision call — infer it from the path of the file being edited."
    echo ""
    for subdir in "${FOUND_REPOS[@]}"; do
      subproject=$(basename "$subdir")
      echo "PROJECT CAPTURE CONVENTION ($subproject):"
      cat "$subdir/.claude/uroboro.tags"
      echo ""
    done
    echo "Follow the matching convention for each repo's files."
  else
    echo ""
    echo "MULTI-REPO SESSION: no uroboro.tags found in subdirectories. Pass 'project' explicitly based on which repo's files you are editing."
  fi
fi

# On startup/resume (not compact/clear), suggest uro_recap
if [ "$SOURCE" = "startup" ] || [ "$SOURCE" = "resume" ]; then
  echo ""
  echo "Consider running uro_recap to load recent context for $PROJECT."
fi

exit 0
