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

# Check for project-specific tag convention
TAGS_FILE="$CWD/.claude/uroboro.tags"
if [ -f "$TAGS_FILE" ]; then
  echo ""
  echo "PROJECT CAPTURE CONVENTION ($PROJECT):"
  cat "$TAGS_FILE"
  echo ""
  echo "Follow this tagging scheme for all uro_capture calls in this session."
fi

# On startup/resume (not compact/clear), suggest uro_recap
if [ "$SOURCE" = "startup" ] || [ "$SOURCE" = "resume" ]; then
  echo ""
  echo "Consider running uro_recap to load recent context for $PROJECT."
fi

exit 0
