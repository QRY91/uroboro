#!/bin/bash
# Quick setup script for dogfooding the uroboro VS Code extension

echo "🐍 Setting up uroboro VS Code extension for dogfooding..."

# Check if we're in the right directory
if [ ! -f "package.json" ]; then
    echo "❌ Run this from the vscode-extension directory"
    exit 1
fi

# Initialize npm if needed
if [ ! -f "package-lock.json" ]; then
    echo "📦 Installing dependencies..."
    npm install
fi

# Compile TypeScript (ignore type errors for now - we're dogfooding!)
echo "🔨 Compiling extension..."
npx tsc --noEmitOnError false

# Check if VS Code is available
if command -v code &> /dev/null; then
    echo "🚀 Ready to test! Run:"
    echo "  code --extensionDevelopmentPath=$(pwd)"
    echo ""
    echo "Or press F5 in VS Code to launch Extension Development Host"
else
    echo "⚠️  VS Code not found in PATH"
    echo "Manual setup:"
    echo "1. Open VS Code"
    echo "2. File → Open Folder → $(pwd)"
    echo "3. Press F5 to launch Extension Development Host"
fi

echo ""
echo "🎯 Quick test commands once VS Code is running:"
echo "  Ctrl+Shift+U  → Quick capture"
echo "  Ctrl+Shift+C  → Quick capture (vim mode)"
echo "  Ctrl+Shift+P  → Quick publish blog (vim mode)"
echo "  Ctrl+Shift+S  → Show status"
echo ""
echo "📝 Make sure uroboro is installed in your Python environment:"
echo "  cd .. && pip install -e ." 