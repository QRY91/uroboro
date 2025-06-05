#!/bin/bash

# Filter uroboro-specific captures and generate devlog
# Usage: ./scripts/filter-uroboro-devlog.sh [days]

DAYS=${1:-7}
DATA_DIR="$HOME/.local/share/uroboro/daily"
TEMP_FILE="/tmp/uroboro-filtered.md"

echo "🔍 Filtering uroboro-specific captures from last $DAYS days..."

# Find uroboro-related captures
find "$DATA_DIR" -name "*.md" -mtime -$DAYS -exec grep -l -i \
    "uroboro\|landing page\|cross-platform\|XDG\|patch.notes\|CLI.*uro\|WebSocket.*uro\|go build\|internal/\|README.*uro" {} \; | \
while read file; do
    echo "Processing $file..."
    # Extract uroboro-specific entries
    grep -i -A5 -B1 \
        "uroboro\|landing page\|cross-platform\|XDG\|patch.notes\|CLI.*uro\|internal/\|GetDataDir\|common/dirs" \
        "$file" >> "$TEMP_FILE"
done

echo "📋 Generating uroboro-specific devlog..."

# Use the filtered content to generate devlog
if [ -s "$TEMP_FILE" ]; then
    echo "✅ Found uroboro-specific activity"
    echo "--- UROBORO DEVELOPMENT LOG ---"
    echo ""
    cat "$TEMP_FILE" | head -50  # Show first 50 lines of filtered content
    echo ""
    echo "--- END FILTERED CAPTURES ---"
    rm "$TEMP_FILE"
else
    echo "❌ No uroboro-specific activity found in last $DAYS days"
fi 