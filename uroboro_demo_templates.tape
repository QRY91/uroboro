# uroboro_demo_templates.tape
# Project templates demo

# Output GIF file
Output assets/uroboro_demo_templates.gif

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
Type "cd /tmp && rm -rf uroboro_templates_demo && mkdir uroboro_templates_demo && cd uroboro_templates_demo" Enter Sleep 1s
Type "cp -r /home/qry/stuff/projects/uroboro/* ." Enter Sleep 1s
Type "pip install -e . --quiet" Enter Sleep 3s
Type "clear" Enter Sleep 500ms
Show

# --- Demo: List Templates ---
Type "# uroboro project templates" Sleep 1s Enter Sleep 1s
Type "uro project --list" Sleep 1s Enter
Sleep 3s

# --- Demo: Create Web Project ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Create a new web application" Sleep 1s Enter Sleep 1s
Type "uro project create my-web-app --template web --name 'My Web Application' --description 'A modern React web app'" Sleep 2s Enter
Sleep 4s

# --- Demo: Explore Generated Structure ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Check what was generated" Sleep 1s Enter Sleep 1s
Type "ls -la my-web-app/" Sleep 1s Enter
Sleep 2s

Type "echo '--- .devlog/README.md ---'" Sleep 1s Enter Sleep 1s
Type "head -20 my-web-app/.devlog/README.md" Sleep 1.5s Enter
Sleep 4s

# --- Demo: Create API Project ---
Hide
Type "clear" Enter Sleep 1s
Show
Type "# Create an API project" Sleep 1s Enter Sleep 1s
Type "uro project create my-api --template api --name 'My API Service' --tech-stack FastAPI PostgreSQL Redis" Sleep 2s Enter
Sleep 4s

# --- Demo: Show API Project Structure ---
Type "ls -la my-api/" Sleep 1s Enter
Sleep 2s

Type "echo '--- API .gitignore ---'" Sleep 1s Enter Sleep 1s
Type "head -10 my-api/.gitignore" Sleep 1.5s Enter
Sleep 3s

Hide
Type "clear" Enter Sleep 1s
Show
Type "echo '📋 6 templates available: web, api, game, tool, research, mobile'" Sleep 1.5s Enter
Sleep 2s 