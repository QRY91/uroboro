#!/bin/bash
# uroboro enforcement hook: Stop
# Audits for zero-capture sessions with significant work.
# Stop hook stdout is NOT injected into conversation — uses JSON
# {"decision":"block","reason":"..."} to nudge the agent.
# Checks stop_hook_active to prevent infinite loops.

INPUT=$(cat)
STOP_HOOK_ACTIVE=$(echo "$INPUT" | jq -r '.stop_hook_active // "false"')
TRANSCRIPT=$(echo "$INPUT" | jq -r '.transcript_path')

# Expand ~ to $HOME (bash doesn't expand tilde in variables)
TRANSCRIPT="${TRANSCRIPT/#\~/$HOME}"

# If we already blocked once this turn, don't block again (prevent loops)
if [ "$STOP_HOOK_ACTIVE" = "true" ]; then
  exit 0
fi

# Need a transcript to audit
if [ -z "$TRANSCRIPT" ] || [ ! -f "$TRANSCRIPT" ]; then
  exit 0
fi

# Count tool calls and uroboro calls in transcript
TOOL_CALLS=$(grep -c '"tool_use"' "$TRANSCRIPT" 2>/dev/null) || TOOL_CALLS=0
URO_CALLS=$(grep -cE '"uro_(capture|decision|blocker|question)"' "$TRANSCRIPT" 2>/dev/null) || URO_CALLS=0

# Threshold: if >20 tool calls but 0 uro calls, nudge
if [ "$TOOL_CALLS" -gt 20 ] && [ "$URO_CALLS" -eq 0 ]; then
  cat <<'NUDGE'
{"decision":"block","reason":"SESSION AUDIT: This session had significant work (>20 tool calls) but zero uroboro captures. If any decisions, trade-offs, or notable observations were made, capture them now with uro_decision or uro_capture before the session context is lost."}
NUDGE
fi

exit 0
