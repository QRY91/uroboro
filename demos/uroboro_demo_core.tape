# uroboro_demo_core.tape
# Core workflow: capture → status → publish (mobile-friendly)

# Output GIF file
Output landing-page/assets/uroboro_demo_core.gif

# Mobile-Friendly Terminal Setup
Set Theme "Dracula"
Set FontFamily "JetBrains Mono"
Set FontSize 18
Set Width 1200
Set Height 800
Set TypingSpeed 50ms
Set LoopOffset "8%"
Set LineHeight 1.4

# --- Demo: Capture Development Work (10 seconds) ---
Hide
Type "cd /home/qry/stuff/projects/uroboro" Enter Sleep 500ms
Type "clear" Enter Sleep 500ms
Show

Type "# Core Workflow Demo - Three Commands That's It" Sleep 1s Enter Sleep 1s
Type "# Command 1: CAPTURE (10 seconds)" Sleep 1s Enter Sleep 1s

Type "uroboro capture 'Implemented WebSocket real-time notifications'" Sleep 1.5s Enter
Sleep 3s

Type "# Same command, short format:" Sleep 1s Enter
Type "uro -c 'Fixed memory leak in connection handler'" Sleep 1.5s Enter  
Sleep 3s

# --- Demo: Check Status (complete overview) ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Command 3: STATUS (complete overview)" Sleep 1s Enter Sleep 1s

Type "uro -s" Sleep 1s Enter
Sleep 4s

Type "# Same data, full format:" Sleep 1s Enter
Type "uroboro status" Sleep 1s Enter
Sleep 4s

# --- Demo: Publish Content (2 minutes) ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Command 2: PUBLISH (2 minutes)" Sleep 1s Enter Sleep 1s

Type "uroboro publish --blog" Sleep 1.5s Enter
Sleep 8s

Type "# Short format also works:" Sleep 1s Enter
Type "uro -p --devlog" Sleep 1.5s Enter
Sleep 6s

Hide
Type "clear" Enter Sleep 1s
Show
Type "echo '🎯 Core workflow: Both uroboro and uro work seamlessly'" Sleep 1.5s Enter
Type "echo '✅ Professional tools, not amateur prototypes'" Sleep 1s Enter
Sleep 3s 