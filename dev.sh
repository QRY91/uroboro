#!/bin/bash

# 🛡️ UROBORO DEVELOPMENT GUARDRAILS SCRIPT
# This script prevents regressions and ensures consistent development workflow

set -e  # Exit on any error

echo "🛡️ Starting Uroboro Development Guardrails..."

# STEP 1: Kill any existing processes
echo "📥 Killing existing uroboro processes..."
pkill -f uroboro || true
sleep 1

# STEP 2: Verify we're in the right directory
if [[ ! -f "go.mod" ]] || [[ ! -d "web" ]]; then
    echo "❌ ERROR: Must run from uroboro root directory"
    exit 1
fi

# STEP 3: Build Svelte app with verification
echo "🔨 Building Svelte app..."
cd web
if ! command -v pnpm &> /dev/null; then
    echo "❌ ERROR: pnpm not found. Please install pnpm first."
    exit 1
fi

pnpm run build

# Verify dist files exist
if [[ ! -f "dist/index.html" ]]; then
    echo "❌ ERROR: Svelte build failed - dist/index.html not found"
    exit 1
fi

echo "✅ Svelte app built successfully"
cd ..

# STEP 4: Check for embedded template conflicts
echo "🔍 Checking for embedded template conflicts..."
if grep -q "playBtn" internal/journey/server.go; then
    echo "⚠️  WARNING: Found embedded HTML template in server.go"
    echo "   This will override the Svelte app!"
    echo ""
    echo "🔧 Auto-fixing: Commenting out embedded template..."

    # Create backup
    cp internal/journey/server.go internal/journey/server.go.backup

    # Comment out the embedded template constants
    sed -i 's/^const journeyCSS = `/\/\/ DISABLED: const journeyCSS = `/' internal/journey/server.go
    sed -i 's/^const journeyJS = `/\/\/ DISABLED: const journeyJS = `/' internal/journey/server.go

    echo "✅ Embedded templates disabled"
fi

# STEP 5: Build Go binary
echo "🔨 Building Go binary..."
go build -o uroboro ./cmd/uroboro

# STEP 6: Start server
echo "🚀 Starting uroboro server..."
./uroboro publish --journey --days 7 --port 8080 &
SERVER_PID=$!

# STEP 7: Wait and verify
echo "⏳ Waiting for server to start..."
sleep 3

# Test if server is responding
if curl -s http://localhost:8080 > /dev/null; then
    echo "✅ Server started successfully"

    # Check what's being served
    TITLE=$(curl -s http://localhost:8080 | grep -o '<title>[^<]*' | head -1)
    echo "📄 Serving: $TITLE"

    if echo "$TITLE" | grep -q "Uroboro Journey Timeline"; then
        echo "🎉 SUCCESS: Modern Svelte app is being served!"
        echo ""
        echo "🌐 Access your timeline at: http://localhost:8080"
        echo "🔧 API endpoint at: http://localhost:8080/api/journey"
        echo ""
        echo "🛡️ Guardrails PASSED - no regressions detected"
    else
        echo "❌ FAILURE: Wrong template being served"
        echo "   Expected: Uroboro Journey Timeline"
        echo "   Got: $TITLE"
        kill $SERVER_PID
        exit 1
    fi
else
    echo "❌ ERROR: Server not responding"
    kill $SERVER_PID || true
    exit 1
fi

echo ""
echo "🎯 DEVELOPMENT READY!"
echo "   - Svelte app built and verified"
echo "   - Embedded conflicts resolved"
echo "   - Server running on port 8080"
echo "   - Process ID: $SERVER_PID"
echo ""
echo "💡 To stop server: kill $SERVER_PID"
echo "🔄 To restart: ./dev.sh"
