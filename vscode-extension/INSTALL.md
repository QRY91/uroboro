# uroboro VS Code Extension Installation

## 🚀 Quick Install (Any VS Code/Cursor Window)

### Method 1: Install Pre-built Package
```bash
# From the uroboro/vscode-extension directory:
cd vscode-extension
npm install
npx @vscode/vsce package
cursor --install-extension uroboro-vscode-0.1.0.vsix
```

### Method 2: Development Install
```bash
# From the uroboro/vscode-extension directory:
./setup-dev.sh
# Then in VS Code: File → Open Folder → select vscode-extension
# Press F5 to launch Extension Development Host
```

## ⚙️ Configuration

After installation, configure the Python path:
1. **Ctrl+Shift+P** → "Preferences: Open Settings"
2. Search for **"uroboro"**
3. Set **"Python Path"** to your Python executable (usually `python3` on Linux)

## 🎯 Usage

| Command | Keybinding | Description |
|---------|------------|-------------|
| **Quick Capture** | `Ctrl+Shift+U` | Git-style capture (save to commit) |
| **Alt Capture** | `Ctrl+Alt+C` | Alternative capture shortcut |
| **Code Context** | `Ctrl+Alt+X` | Capture with file/line context |
| **Quick Publish** | `Ctrl+Alt+P` | Generate blog post in editor |
| **Show Status** | `Ctrl+Shift+S` | View recent captures |

## 🔧 Requirements

- **uroboro installed**: `pip install -e .` from uroboro root
- **VS Code/Cursor**: 1.74.0 or newer
- **Python 3.8+**: Accessible via command line

## 🐛 Troubleshooting

### "Failed to capture insight"
- Check uroboro output panel: **View** → **Output** → select "uroboro"
- Verify Python path in settings
- Ensure uroboro is installed: `python3 -m src.cli --help`

### "Cannot open file"
- Extension runs commands from uroboro project root
- Check working directory in output panel
- Verify file paths are being resolved correctly

## 📋 Features

✅ **Git-style captures** - Temp file workflow, no dialogs  
✅ **Smart publishing** - Markdown/text choice, auto-open in editor  
✅ **Status integration** - Live activity in status bar  
✅ **Code context** - Right-click selection to capture with location  
✅ **Universal keybindings** - Work with/without vim extension  
✅ **Local-first** - Uses your existing uroboro installation  

## 🎪 Dogfooding Workflow

1. **Code normally** in Cursor/VS Code
2. **Hit insight** → `Ctrl+Shift+U` → type in temp file → Ctrl+S
3. **End of session** → `Ctrl+Alt+P` → choose format → edit in place
4. **Publish anywhere** → Blog, LinkedIn, dev.to

Perfect for experimental/fat-snake branch users who want IDE integration! 