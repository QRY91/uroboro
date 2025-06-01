# uroboro_demo_tamagoro.tape
# tamagoro system: gamified local model training

# Output GIF file
Output assets/uroboro_demo_tamagoro.gif

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

# --- Demo: Spawn Your First Egg ---
Type "# tamagoro: gamified AI training 🥚" Sleep 1s Enter Sleep 1s
Type "uro egg --spawn academic-writer --type academic" Sleep 1.5s Enter
Sleep 3s

# --- Demo: Check Egg Status ---
Type "uro egg --stats academic-writer" Sleep 1.5s Enter
Sleep 4s

# --- Demo: Auto-feeding in Action ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Auto-feeding through normal usage" Sleep 1s Enter Sleep 1s
Type "uro generate --academic --feed-egg academic-writer" Sleep 1.5s Enter
Sleep 4s

# --- Demo: Check Growth ---
Type "uro egg --stats academic-writer" Sleep 1.5s Enter
Sleep 4s

# --- Demo: Health Monitoring ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Built-in overingestion protection" Sleep 1s Enter Sleep 1s
Type "uro egg --health" Sleep 1.5s Enter
Sleep 4s

# --- Demo: Farm Overview ---
Type "uro egg --list" Sleep 1.5s Enter
Sleep 3s

Hide
Type "clear" Enter Sleep 1s
Show
Type "# 🥚 tamagoro: where AI meets tamagotchi" Sleep 1s Enter Sleep 1s
Type "echo '100% local, 100% private, 100% yours'" Sleep 1s Enter
Sleep 2s 