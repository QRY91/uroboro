# uroboro_demo.tape
# The Self-Documenting Content Pipeline - Sacred Workflow

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
Type "clear" Enter Sleep 500ms
Show

# --- Demo: First Time Experience ---
Type "# Welcome to uroboro - The Self-Documenting Content Pipeline 🐍" Sleep 1.5s Enter Sleep 2s
Type "uroboro --help" Sleep 1s Enter
Sleep 5s

Hide
Type "clear" Enter Sleep 1s
Show

# --- Demo: Sacred Command 1 - CAPTURE (10 seconds) ---
Type "# Sacred Command 1: CAPTURE (10 seconds)" Sleep 1.5s Enter Sleep 1s
Type "uroboro capture 'Implemented user authentication with OAuth2 and secure sessions'" Sleep 1.5s Enter
Sleep 4s

Type "uroboro capture 'Fixed critical payment processing bug - prevented data corruption'" Sleep 1.5s Enter
Sleep 4s

Type "uroboro capture 'Added real-time notifications using WebSockets - 3x faster than polling'" Sleep 1.5s Enter
Sleep 4s

# --- Demo: Sacred Command 3 - STATUS (complete overview) ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Sacred Command 3: STATUS (complete overview)" Sleep 1s Enter Sleep 1s
Type "uroboro status" Sleep 1s Enter
Sleep 5s

# --- Demo: Sacred Command 2 - PUBLISH (2 minutes) ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Sacred Command 2: PUBLISH (2 minutes)" Sleep 1.5s Enter Sleep 1s
Type "uroboro publish --blog" Sleep 1.5s Enter
Sleep 8s

# --- Demo: Show the generated content ---
Hide
Type "ls output/posts/" Enter Sleep 2s
Type "head -20 output/posts/dev-update-*.md" Enter Sleep 5s
Show

# --- End with the truth ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "echo '🎯 Three commands beats seventeen - thats the sacred workflow'" Sleep 1.5s Enter
Sleep 2s
Type "echo '🐍 uroboro: The tool that documents itself'" Sleep 1s Enter
Sleep 3s 