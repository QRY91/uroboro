#!/bin/bash
# uroboro enforcement hook: PostToolUse
# Counter-based periodic nudge during active work.
# Matched to Edit|Write|Bash in settings.json (high-signal tools).
# stdout becomes agent context after tool result.
# Opt-in: reads ~/.config/uroboro/enforcement.json

CONFIG="$HOME/.config/uroboro/enforcement.json"
if [ -f "$CONFIG" ]; then
  ENABLED=$(jq -r '.post_tool_nudge.enabled // false' "$CONFIG" 2>/dev/null)
  if [ "$ENABLED" != "true" ]; then
    exit 0
  fi
  THRESHOLD=$(jq -r '.post_tool_nudge.threshold // 15' "$CONFIG" 2>/dev/null)
else
  # Not configured = not enabled (opt-in)
  exit 0
fi

# Validate threshold
if ! [[ "$THRESHOLD" =~ ^[0-9]+$ ]] || [ "$THRESHOLD" -lt 1 ]; then
  THRESHOLD=15
fi

# Counter file (shared across sessions, resets at reboot via /tmp)
COUNTER_FILE="/tmp/uroboro-nudge-counter"

# Read current count
COUNT=0
if [ -f "$COUNTER_FILE" ]; then
  COUNT=$(cat "$COUNTER_FILE" 2>/dev/null) || COUNT=0
fi

# Increment
COUNT=$((COUNT + 1))

# Check threshold
if [ "$COUNT" -ge "$THRESHOLD" ]; then
  echo "UROBORO NUDGE: $COUNT tool calls since last capture check. If any decisions or trade-offs were involved, consider calling uro_decision or uro_capture."
  COUNT=0
fi

# Write back
echo "$COUNT" > "$COUNTER_FILE"

exit 0
