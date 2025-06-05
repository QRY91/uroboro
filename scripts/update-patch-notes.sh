#!/bin/bash

# Update patch notes with uroboro-only development content
# Usage: ./scripts/update-patch-notes.sh

echo "🚀 Updating uroboro patch notes..."

echo "📋 Step 1: Capture today's uroboro work"
echo "Use: ./uro-test capture 'your uroboro work description' --project 'uroboro' --tags 'feature,fix,docs'"
echo ""

echo "📊 Step 2: Generate uroboro-only devlog"
./scripts/filter-uroboro-devlog.sh 7

echo ""
echo "📝 Step 3: Manual update needed"
echo "Copy the filtered content above into landing-page/patch-notes.html"
echo "Replace the existing article content with real uroboro development updates"

echo ""
echo "🔄 Going forward workflow:"
echo "1. Always use --project 'uroboro' when capturing uroboro work"
echo "2. Run this script weekly to generate new patch notes"
echo "3. Update patch-notes.html with the real filtered content"
echo ""
echo "✅ Example capture command:"
echo "./uro-test capture 'your uroboro development work' --project 'uroboro' --tags 'relevant,tags'" 