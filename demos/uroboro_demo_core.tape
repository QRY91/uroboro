# uroboro_demo_core.tape
# Sacred workflow: capture → publish → status

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
Type "clear" Enter Sleep 500ms
Show

# --- Show the sacred commands ---
Type "uroboro --help" Sleep 1s Enter
Sleep 4s

Hide
Type "clear" Enter Sleep 1s
Show

# --- Demo: Capture Development Work (10 seconds) ---
Type "# Sacred Command 1: CAPTURE (10 seconds)" Sleep 1s Enter Sleep 1s
Type "uroboro capture 'Implemented user authentication with OAuth2 and JWT tokens'" Sleep 1.5s Enter
Sleep 3s

Type "uroboro capture 'Fixed critical memory leak in WebSocket connections - cut memory usage by 40%'" Sleep 1.5s Enter  
Sleep 3s

Type "uroboro capture 'Added real-time collaboration features using WebSockets'" Sleep 1.5s Enter
Sleep 3s

# --- Demo: Check Status (complete overview) ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Sacred Command 3: STATUS (complete overview)" Sleep 1s Enter Sleep 1s
Type "uroboro status" Sleep 1s Enter
Sleep 4s

# --- Demo: Publish Content (2 minutes) ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Sacred Command 2: PUBLISH (2 minutes)" Sleep 1s Enter Sleep 1s
Type "uroboro publish --blog" Sleep 1.5s Enter
Sleep 8s

Hide
Type "clear" Enter Sleep 1s
Show
Type "# That's the sacred workflow - three commands that's it!" Sleep 1s Enter Sleep 1s
Type "echo '🎯 capture → publish → status'" Sleep 1s Enter
Sleep 2s 