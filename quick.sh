#!/bin/bash

# 🚀 QUICK REBUILD & RUN
# Fast iteration script for development

set -e

echo "🚀 Quick rebuild starting..."

# Kill existing
pkill -f uroboro || true

# Build frontend
echo "🔨 Building Svelte..."
cd web && pnpm run build && cd ..

# Build backend
echo "🔨 Building Go..."
go build -o uroboro ./cmd/uroboro

# Start server
echo "🚀 Starting server..."
echo "✅ Quick rebuild complete!"
echo "🌐 http://localhost:8080"
echo "🛑 Press Ctrl+C to stop server"
echo ""
./uroboro publish --journey --days 7 --port 8080
