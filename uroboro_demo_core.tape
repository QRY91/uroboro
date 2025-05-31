# uroboro_demo_core.tape
# Core workflow: capture → status → generate

# Output GIF file
Output assets/uroboro_demo_core.gif

# Terminal Setup
Set Theme "Dracula"
Set FontFamily "JetBrains Mono"
Set FontSize 22
Set Width 1100
Set Height 650
Set TypingSpeed 60ms
Set LoopOffset "10%"
Set LineHeight 1.3

# --- Setup Demo Environment ---
Hide
Type "cd /tmp && rm -rf uroboro_demo && mkdir uroboro_demo && cd uroboro_demo" Enter Sleep 1s
Type "cp -r /home/qry/stuff/projects/uroboro/* ." Enter Sleep 1s
Type "pip install -e . --quiet" Enter Sleep 3s
Type "mkdir -p my-project/.devlog" Enter Sleep 500ms
Type "echo '# My Web App' > my-project/README.md" Enter Sleep 500ms
Type "clear" Enter Sleep 500ms
Show

# --- Show initial help ---
Type "uro --help" Sleep 1s Enter
Sleep 4s

Hide
Type "clear" Enter Sleep 1s
Show

# --- Demo: Capture Development Work ---
Type "# Capturing development insights" Sleep 1s Enter Sleep 1s
Type "uro capture 'Implemented user authentication with OAuth2 and JWT tokens'" Sleep 1.5s Enter
Sleep 3s

Type "uro capture 'Fixed critical memory leak in WebSocket connections' --tags bugfix performance" Sleep 1.5s Enter  
Sleep 3s

Type "uro capture 'Added real-time collaboration features' --project my-project --tags feature realtime" Sleep 1.5s Enter
Sleep 3s

# --- Demo: Check Status ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "uro status" Sleep 1s Enter
Sleep 4s

# --- Demo: Generate Content ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Generate content from captured insights" Sleep 1s Enter Sleep 1s
Type "uro generate --preview --output devlog" Sleep 1.5s Enter
Sleep 6s

Hide
Type "clear" Enter Sleep 1s
Show
Type "# That's the core uroboro workflow!" Sleep 1s Enter Sleep 1s
Type "echo '🐍 Capture → Generate → Share'" Sleep 1s Enter
Sleep 2s 