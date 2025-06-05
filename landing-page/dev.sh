#!/bin/bash

echo "🌐 Uroboro Landing Page Dev Server"
echo "=================================="

PORT=8000

# Check for available dev servers (in order of preference)
if command -v bun &> /dev/null; then
    echo "🔥 Using Bun (fastest)"
    bun --serve --port $PORT .
elif command -v deno &> /dev/null; then
    echo "🦕 Using Deno"
    deno run --allow-net --allow-read -L debug https://deno.land/std@0.208.0/http/file_server.ts --port $PORT . --header "Cache-Control:no-cache, no-store, must-revalidate" --header "Pragma:no-cache" --header "Expires:0"
elif command -v php &> /dev/null; then
    echo "🐘 Using PHP"
    php -S localhost:$PORT
elif command -v python3 &> /dev/null; then
    echo "🐍 Using Python"
    python3 -m http.server $PORT
elif command -v python &> /dev/null; then
    echo "🐍 Using Python"
    python -m http.server $PORT
else
    echo "❌ No dev server found!"
    echo "Install one of: bun, deno, php, or python"
    exit 1
fi 