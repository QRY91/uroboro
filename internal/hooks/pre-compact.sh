#!/bin/bash
# uroboro enforcement hook: PreCompact
# Injects capture checkpoint reminder before context compaction.
# stdout becomes agent context.
# Opt-in: reads ~/.config/uroboro/enforcement.json

CONFIG="$HOME/.config/uroboro/enforcement.json"
if [ -f "$CONFIG" ]; then
  ENABLED=$(jq -r '.pre_compact.enabled // false' "$CONFIG" 2>/dev/null)
  if [ "$ENABLED" != "true" ]; then
    exit 0
  fi
else
  # Not configured = not enabled (opt-in)
  exit 0
fi

cat <<'CHECKPOINT'
UROBORO CHECKPOINT: Context is about to be compacted. If you made any decisions, trade-offs, or notable observations that haven't been captured yet, call uro_decision or uro_capture NOW before context is lost.
CHECKPOINT

exit 0
