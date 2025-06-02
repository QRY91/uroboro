# uroboro VS Code Extension

**Seamless development insight capture without breaking flow.**

## 🎯 Core Features

### Quick Capture (Ctrl+Shift+U)
- Instant insight capture from anywhere in VS Code
- No context switching, stays in flow state
- Auto-updates status bar with recent activity

### Code Context Capture
- Right-click selected code → "Capture with Code Context"
- Automatically includes file name and line number
- Perfect for documenting solutions and patterns

### Status Integration
- Live status bar showing recent captures
- Click to view full uroboro status
- Quick access to publish commands

### Publish from Editor
- Command palette: "uroboro: Publish Content"
- Choose blog post, social content, or dev log
- Auto-opens generated blog posts in editor

## 🚀 Getting Started

### Prerequisites
1. Have uroboro installed: `pip install -e .` (from uroboro repo)
2. VS Code 1.74.0 or newer

### Installation (Development)
1. Copy this `vscode-extension` folder to your VS Code extensions directory
2. Run `npm install` in the extension folder
3. Run `npm run compile` to build
4. Press F5 to launch Extension Development Host

### Usage

**Quick capture while coding:**
```
Ctrl+Shift+U → Type insight → Enter
✅ Captured: Fixed race condition in auth flow
```

**Capture code context:**
```
1. Select problematic/interesting code
2. Right-click → "Capture with Code Context" 
3. Describe your insight
✅ Captured: This pattern prevents memory leaks (in auth.py:42)
```

**Publish from editor:**
```
Ctrl+Shift+P → "uroboro: Quick Publish Blog"
🔄 Generating content...
✅ Blog post published
```

## ⚙️ Configuration

Access via Settings → Extensions → uroboro:

- **Python Path**: Path to Python with uroboro installed
- **Auto Git Capture**: Automatically capture git commits  
- **Show Status Bar**: Display uroboro status in status bar
- **Capture Template**: Format for code context captures

## 🎪 Commands

| Command | Keybinding | Description |
|---------|------------|-------------|
| `uroboro.capture` | `Ctrl+Shift+U` | Quick insight capture |
| `uroboro.captureSelection` | - | Capture with code context |
| `uroboro.status` | `Ctrl+Shift+S` | Show status |
| `uroboro.publish` | - | Choose content type to publish |
| `uroboro.quickPublish` | - | Instant blog post |

## 🔧 Development

This extension bridges VS Code with the uroboro CLI tool:

- **Zero dependencies**: Uses your existing uroboro installation
- **Local-first**: No external services, works offline
- **Lightweight**: Minimal resource usage
- **Privacy-focused**: Your insights stay on your machine

## 🐍 Integration with uroboro

The extension executes these CLI commands:
```bash
uro capture "insight" --tags code-context
uro status --recent  
uro publish --type blog
```

Perfect for the fat-snake branch - full-featured integration without bloating main. 