#!/bin/bash
# Sync uroboro captures to OpenViking using local Ollama models.
# Designed for overnight batch runs — no external APIs, zero cost.
#
# Usage:
#   ./scripts/sync-overnight.sh              # sync last 30 days
#   ./scripts/sync-overnight.sh --days 90    # sync last 90 days
#   ./scripts/sync-overnight.sh --since 2025-01-01  # since date
#
# Prerequisites:
#   - Ollama running with nomic-embed-text and qwen3:8b
#   - OpenViking venv at .venv-openviking/

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
VENV="$PROJECT_DIR/.venv-openviking"
LOG="/tmp/uroboro-sync-$(date +%Y%m%d-%H%M%S).log"

echo "=== uroboro → OpenViking sync ===" | tee "$LOG"
echo "Started: $(date)" | tee -a "$LOG"
echo "Log: $LOG"

# Check Ollama
if ! curl -sf http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo "ERROR: Ollama not running" | tee -a "$LOG"
    exit 1
fi

# Start OpenViking server if not running
if ! curl -sf http://localhost:1933/api/v1/observer/system > /dev/null 2>&1; then
    echo "Starting OpenViking server..." | tee -a "$LOG"
    source "$VENV/bin/activate"
    python3 -c "
import uvicorn
from openviking.server.app import create_app, load_server_config
config = load_server_config()
app = create_app(config)
uvicorn.run(app, host=config.host, port=config.port, log_level='warning')
" >> "$LOG" 2>&1 &
    OV_PID=$!
    sleep 5

    if ! curl -sf http://localhost:1933/api/v1/observer/system > /dev/null 2>&1; then
        echo "ERROR: OpenViking failed to start" | tee -a "$LOG"
        exit 1
    fi
    echo "OpenViking started (PID: $OV_PID)" | tee -a "$LOG"
fi

# Run sync (direct mode: files + Ollama embeddings, no VLM)
echo "Running sync..." | tee -a "$LOG"
"$PROJECT_DIR/uroboro" sync --direct "$@" 2>&1 | tee -a "$LOG"

echo "" | tee -a "$LOG"
echo "Finished: $(date)" | tee -a "$LOG"
