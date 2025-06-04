# uroboro_demo.tape
# Complete uroboro workflow - Professional quality assurance (mobile-friendly)

# Output GIF file
Output landing-page/assets/uroboro_demo.gif

# Mobile-Friendly Terminal Setup
Set Theme "Dracula"
Set FontFamily "JetBrains Mono"
Set FontSize 18
Set Width 1200
Set Height 800
Set TypingSpeed 50ms
Set LoopOffset "8%"
Set LineHeight 1.4

# --- Demo: Complete Overview ---
Hide
Type "cd /home/qry/stuff/projects/uroboro" Enter Sleep 500ms
Type "clear" Enter Sleep 500ms
Show

Type "# uroboro: The Self-Documenting Content Pipeline 🐍" Sleep 1.5s Enter Sleep 1s
Type "# Professional tools, not amateur prototypes" Sleep 1s Enter Sleep 1s

# --- Demo: Dual Command Interface ---
Type "# Both uroboro and uro work seamlessly:" Sleep 1s Enter Sleep 1s

Type "uroboro capture 'Fixed memory leak in connection handler'" Sleep 1.5s Enter
Sleep 3s

Type "uro -c 'Added rate limiting to prevent abuse'" Sleep 1.5s Enter
Sleep 3s

Type "# XDG-compliant unified data storage:" Sleep 1s Enter
Type "uro -c 'All data in ~/.local/share/uroboro/daily/'" Sleep 1.5s Enter
Sleep 3s

# --- Demo: Status Shows Professional Development ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Complete development pipeline overview:" Sleep 1s Enter Sleep 1s

Type "uroboro status" Sleep 1s Enter
Sleep 5s

# --- Demo: Quality Assurance in Action ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Quality assurance: Full test coverage + CI/CD" Sleep 1s Enter Sleep 1s
Type "make test" Sleep 1s Enter
Sleep 4s

Type "# Professional publishing with local AI:" Sleep 1s Enter
Type "uro -p --blog" Sleep 1.5s Enter
Sleep 8s

# --- Demo: Show the transformation ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "echo '✅ Removed 2,497 lines of Python code (60.8% cleanup)'" Sleep 1.5s Enter
Type "echo '🛡️ Full unit test coverage + GitHub Actions CI'" Sleep 1s Enter
Type "echo '📁 XDG compliance: ~/.local/share/uroboro/daily/'" Sleep 1s Enter
Type "echo '⚡ Dual interface: uroboro and uro from same codebase'" Sleep 1s Enter
Sleep 2s

Type "echo '🎯 Three commands beats seventeen - Focus over features'" Sleep 1.5s Enter
Type "echo '🐍 uroboro: Documents itself while helping you document everything else'" Sleep 1s Enter
Sleep 4s 