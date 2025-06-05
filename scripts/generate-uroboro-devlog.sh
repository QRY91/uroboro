#!/bin/bash

# Generate uroboro-specific devlog from organized captures
# Usage: ./scripts/generate-uroboro-devlog.sh [days]

DAYS=${1:-7}
UROBORO_FILE="$HOME/.local/share/uroboro/by-project/uroboro.md"
TEMP_INPUT="/tmp/uroboro-recent.md"

echo "🔍 Generating uroboro-specific devlog for last $DAYS days..."

if [ ! -f "$UROBORO_FILE" ]; then
    echo "❌ Uroboro capture file not found: $UROBORO_FILE"
    echo "Run: python3 scripts/organize-captures.py first"
    exit 1
fi

# Get recent uroboro captures
echo "📄 Extracting recent uroboro captures..."

# Calculate cutoff date (DAYS ago)
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    CUTOFF_DATE=$(date -v-${DAYS}d +%Y-%m-%d)
else
    # Linux
    CUTOFF_DATE=$(date -d "$DAYS days ago" +%Y-%m-%d)
fi

echo "📅 Looking for captures since: $CUTOFF_DATE"

# Extract recent captures
awk -v cutoff="$CUTOFF_DATE" '
    /^## [0-9]{4}-[0-9]{2}-[0-9]{2}T/ {
        date_match = match($0, /([0-9]{4}-[0-9]{2}-[0-9]{2})/)
        if (date_match) {
            capture_date = substr($0, RSTART, RLENGTH)
            if (capture_date >= cutoff) {
                include = 1
            } else {
                include = 0
            }
        }
    }
    include { print }
' "$UROBORO_FILE" > "$TEMP_INPUT"

if [ ! -s "$TEMP_INPUT" ]; then
    echo "❌ No recent uroboro captures found in last $DAYS days"
    rm -f "$TEMP_INPUT"
    exit 1
fi

echo "✅ Found recent uroboro activity:"
echo "--- RECENT UROBORO CAPTURES ---"
head -50 "$TEMP_INPUT"
echo "--- END CAPTURES ---"

# Clean up
rm -f "$TEMP_INPUT"

echo ""
echo "🎯 Next steps:"
echo "1. Copy the captures above to create focused patch notes"
echo "2. Use this content for uroboro-only devlog generation"
echo "3. Update patch-notes.html with real uroboro progress" 