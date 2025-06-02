# uroboro_demo_git.tape
# Sacred workflow with git integration

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
Type "echo 'def authenticate_user(): pass' > auth.py" Enter Sleep 500ms
Type "git add auth.py" Enter Sleep 500ms
Type "git commit -m 'Add authentication system'" Enter Sleep 1s
Type "clear" Enter Sleep 500ms
Show

# --- Demo: Capture with Git Context ---
Type "# Sacred workflow with git integration" Sleep 1s Enter Sleep 1s
Type "uroboro capture 'Implemented OAuth2 authentication system with JWT tokens'" Sleep 1.5s Enter
Sleep 3s

Type "git add -A && git commit -m 'Add OAuth2 authentication'" Sleep 1.5s Enter Sleep 1s

Type "uroboro capture 'Fixed race condition in session management - prevents concurrent logins'" Sleep 1.5s Enter
Sleep 3s

# --- Demo: Status Shows Git Integration ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# STATUS shows git integration automatically" Sleep 1s Enter Sleep 1s
Type "uroboro status" Sleep 1.5s Enter
Sleep 5s

# --- Demo: Publish With Git Context ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# PUBLISH includes git context in content" Sleep 1s Enter Sleep 1s
Type "uroboro publish --blog" Sleep 1.5s Enter
Sleep 8s

# --- Demo: Show Generated Content ---
Hide
Type "ls output/posts/" Enter Sleep 2s
Type "echo '# Content includes git context automatically'" Sleep 1.5s Enter
Show

Hide
Type "clear" Enter Sleep 1s
Show
Type "echo '🔗 Sacred workflow + Git = Seamless documentation'" Sleep 1.5s Enter
Sleep 2s 