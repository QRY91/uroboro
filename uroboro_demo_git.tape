# uroboro_demo_git.tape
# Git integration workflow demo

# Output GIF file
Output assets/uroboro_demo_git.gif

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
Type "cd /tmp && rm -rf uroboro_git_demo && mkdir uroboro_git_demo && cd uroboro_git_demo" Enter Sleep 1s
Type "cp -r /home/qry/stuff/projects/uroboro/* ." Enter Sleep 1s
Type "pip install -e . --quiet" Enter Sleep 3s
Type "git init -b main" Enter Sleep 1s
Type "git config user.name 'Demo User'" Enter Sleep 500ms
Type "git config user.email 'demo@uroboro.dev'" Enter Sleep 500ms
Type "echo '# My Project' > README.md" Enter Sleep 500ms
Type "git add README.md" Enter Sleep 500ms
Type "git commit -m 'Initial commit'" Enter Sleep 1s
Type "echo 'Some code changes' >> app.py" Enter Sleep 500ms
Type "git add app.py" Enter Sleep 500ms
Type "git commit -m 'Add core functionality'" Enter Sleep 1s
Type "clear" Enter Sleep 500ms
Show

# --- Demo: Git Status ---
Type "# uroboro git integration" Sleep 1s Enter Sleep 1s
Type "uro git" Sleep 1s Enter
Sleep 4s

# --- Demo: Install Git Hook ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Install git hook for auto-capture" Sleep 1s Enter Sleep 1s
Type "uro git --hook-install" Sleep 1.5s Enter
Sleep 3s

# --- Demo: Analyze Git History ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Analyze commit patterns" Sleep 1s Enter Sleep 1s
Type "uro git --analyze --days 7" Sleep 1.5s Enter
Sleep 4s

# --- Demo: Manual Commit Capture ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Manually capture recent commits" Sleep 1s Enter Sleep 1s
Type "uro git --capture-commits --days 1" Sleep 1.5s Enter
Sleep 4s

# --- Demo: Test Auto-Capture ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Test the hook with a new commit" Sleep 1s Enter Sleep 1s
Type "echo 'Hook test' >> test.md" Sleep 1s Enter Sleep 1s
Type "git add test.md" Sleep 1s Enter Sleep 1s
Type "git commit -m 'Test auto-capture hook'" Sleep 1.5s Enter
Sleep 3s

Hide
Type "clear" Enter Sleep 1s
Show
Type "echo '🔗 Git + uroboro = Automatic development documentation'" Sleep 1.5s Enter
Sleep 2s 