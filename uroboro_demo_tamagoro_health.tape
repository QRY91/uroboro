# uroboro_demo_tamagoro_health.tape
# tamagoro overingestion protection and health monitoring

# Output GIF file
Output assets/uroboro_demo_tamagoro_health.gif

# Terminal Setup
Set Theme "Dracula"
Set FontFamily "JetBrains Mono"
Set FontSize 20
Set Width 1000
Set Height 600
Set TypingSpeed 50ms
Set LoopOffset "10%"
Set LineHeight 1.3

# --- Setup Demo Environment ---
Hide
Type "cd /tmp && rm -rf uroboro_demo && mkdir uroboro_demo && cd uroboro_demo" Enter Sleep 1s
Type "cp -r /home/qry/stuff/projects/uroboro/* ." Enter Sleep 1s
Type "pip install -e . --quiet" Enter Sleep 2s
Type "clear" Enter Sleep 500ms
Show

# --- Demo: Health System Overview ---
Type "# tamagoro health monitoring 🏥" Sleep 1s Enter Sleep 1s
Type "uro egg --health" Sleep 1.5s Enter
Sleep 5s

# --- Demo: Detailed Egg Health ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "uro egg --stats academic-writer --health" Sleep 1.5s Enter
Sleep 5s

# --- Demo: Protection Features ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Built-in overingestion protection:" Sleep 1s Enter Sleep 500ms
Type "echo '🔄 Content deduplication'" Sleep 800ms Enter Sleep 500ms
Type "echo '📊 Quality trend monitoring'" Sleep 800ms Enter Sleep 500ms
Type "echo '🚫 Rate limiting'" Sleep 800ms Enter Sleep 500ms
Type "echo '🛡️ Health checking'" Sleep 800ms Enter Sleep 500ms
Type "echo '🎯 Skill diversity requirements'" Sleep 800ms Enter
Sleep 3s

Hide
Type "clear" Enter Sleep 1s
Show
Type "# Solves the ouroboros problem 🐍" Sleep 1s Enter Sleep 1s
Type "echo 'AI training without data degradation'" Sleep 1s Enter
Sleep 2s 