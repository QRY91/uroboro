#!/bin/bash
# capture.sh - Quick capture script for Cursor terminal

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if [ $# -eq 0 ]; then
    echo "Usage: ./capture.sh 'your content here' [project-name]"
    exit 1
fi

CONTENT="$1"
PROJECT="${2:-}"

if [ -n "$PROJECT" ]; then
    python3 src/aggregator.py capture "$CONTENT" --project "$PROJECT"
else
    python3 src/aggregator.py capture "$CONTENT"
fi 