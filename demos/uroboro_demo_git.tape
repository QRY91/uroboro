# uroboro_demo_git.tape
# Git integration workflow (mobile-friendly)

# Output GIF file
Output landing-page/assets/uroboro_demo_git.gif

# Mobile-Friendly Terminal Setup
Set Theme "Dracula"
Set FontFamily "JetBrains Mono"
Set FontSize 18
Set Width 1200
Set Height 800
Set TypingSpeed 50ms
Set LoopOffset "8%"
Set LineHeight 1.4

# --- Demo: Git Integration ---
Hide
Type "cd /home/qry/stuff/projects/uroboro" Enter Sleep 500ms
Type "clear" Enter Sleep 500ms
Show

Type "# Git Integration Demo" Sleep 1s Enter Sleep 1s
Type "# Seamless git workflow integration" Sleep 1s Enter Sleep 1s

# Make a commit for demo
Type "echo '// WebSocket pooling optimization' >> temp_demo.js" Sleep 1s Enter
Type "git add temp_demo.js" Sleep 1s Enter
Type "git commit -m 'Add WebSocket connection pooling'" Sleep 1s Enter
Sleep 2s

Type "# Auto-capture from git commit:" Sleep 1s Enter
Type "uroboro capture --auto-git" Sleep 1.5s Enter
Sleep 3s

Type "# Manual capture with context:" Sleep 1s Enter
Type "uro -c 'Optimized connection pooling - 50% faster WebSockets'" Sleep 1.5s Enter
Sleep 3s

# --- Demo: Status Shows Git Integration ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# STATUS shows git integration automatically" Sleep 1s Enter Sleep 1s

Type "uro -s" Sleep 1.5s Enter
Sleep 4s

# --- Demo: Publish With Git Context ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# PUBLISH includes git context automatically" Sleep 1s Enter Sleep 1s

Type "uroboro publish --devlog" Sleep 1.5s Enter
Sleep 8s

# Cleanup
Hide
Type "rm temp_demo.js && git add -A && git commit -m 'Clean up demo'" Enter Sleep 1s
Type "clear" Enter Sleep 500ms
Show

Type "echo '🔗 Git integration: Seamless development documentation'" Sleep 1.5s Enter
Type "echo '📊 Both uroboro and uro work with git context'" Sleep 1s Enter
Sleep 3s 