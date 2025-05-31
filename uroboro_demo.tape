# uroboro_demo.tape
# Essential uroboro workflow with proper timing

# Output GIF file
Output assets/uroboro_demo.gif

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
Type "mkdir -p .devlog" Enter Sleep 500ms
Type "clear" Enter Sleep 500ms
Show

# --- Demo: First Time Experience ---
Type "# Welcome to uroboro - The Self-Documenting Pipeline 🐍" Sleep 1.5s Enter Sleep 2s
Type "uro --help" Sleep 1s Enter
Sleep 5s

Hide
Type "clear" Enter Sleep 1s
Show

# --- Demo: Quick Capture ---
Type "# Capture development insights as you work" Sleep 1.5s Enter Sleep 1s
Type "uro capture 'Implemented user authentication with OAuth2'" Sleep 1.5s Enter
Sleep 4s

Type "uro capture 'Fixed critical payment processing bug' --tags bugfix critical" Sleep 1.5s Enter
Sleep 4s

# --- Demo: Status Check ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Check your recent activity" Sleep 1s Enter Sleep 1s
Type "uro status" Sleep 1s Enter
Sleep 5s

# --- Demo: Generate Content ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Transform captures into polished content" Sleep 1.5s Enter Sleep 1s
Type "uro generate --preview --output devlog" Sleep 1.5s Enter
Sleep 20s

# --- Demo: More Features ---
Hide
Type "clear" Enter Sleep 2s
Show
Type "# Explore more features" Sleep 1s Enter Sleep 1s
Type "uro project --list" Sleep 1s Enter
Sleep 4s

# --- End with flourish ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "echo '🐍 uroboro: Where development work becomes content'" Sleep 1.5s Enter
Sleep 2s
Type "echo 'Visit uroboro.dev for more!'" Sleep 1s Enter
Sleep 3s 