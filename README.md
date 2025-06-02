# uroboro 🐍

**The Self-Documenting Content Pipeline**

Transform development insights into professional content that gets you noticed. **Three commands. That's it.** Local AI only, zero API costs.

![Demo](assets/uroboro_demo.gif)

## 🎯 The Core Workflow

**Three commands beats seventeen.** We discovered this by cutting uroboro from 1,558 lines to 270 lines, removing 14 commands, and keeping only what delivers acknowledgment for your actual work.

```bash
# 1. Capture your work (10 seconds)
uro capture "Fixed auth timeout - cut query time from 3s to 200ms"

# 2. Publish content (2 minutes)  
uro publish --blog

# 3. Check everything (complete overview)
uro status
```

## ✨ What Actually Works

- **🎯 3 Core Commands** - capture, publish, status. That's it.
- **⚡ 10-Second Capture** - Zero flow state interruption
- **📝 2-Minute Publish** - Professional content generation
- **🏠 100% Local AI** - Ollama, no external APIs
- **💰 $0/month costs** - No subscription, no usage fees
- **🔒 Private by design** - Your code stays on your machine
- **🔗 Git integration** - When you need it, not when we think you should

## 🎬 Truth in Action

| Demo | What it shows | Why it matters |
|------|---------------|----------------|
| [**⚡ Core Workflow**](assets/uroboro_demo_core.gif) | capture → status → publish | The actual working product |
| [**🔗 Git Integration**](assets/uroboro_demo_git.gif) | Auto-capture commits | Optional but functional |
| [**🎬 Complete Overview**](assets/uroboro_demo.gif) | Full workflow | Tool that documents itself |

## 🧹 The Great CLI Cleanup

We started with **17 commands** and **1,558 lines** of bloated complexity. We removed 14 commands, kept 3 essential ones, and found that **focus beats features**: improve core functionality instead of adding new complexity.

### What We Kept (Core Commands)
- **`uro capture`** - 10-second insight logging
- **`uro publish`** - Generate blog posts, social content, dev logs  
- **`uro status`** - Complete pipeline overview

### What We Cleaned Out (Feature Bloat)
- ❌ 14 unnecessary commands
- ❌ tamagoro egg system fantasy
- ❌ sensei learning complexity
- ❌ interview system bloat
- ❌ Q&A preparation overhead
- ❌ academic mode distraction
- ❌ 6 different project templates
- ❌ Voice training complexity

## 🚀 Get Started in 3 Minutes

### 1. Install
```bash
git clone https://github.com/qry91/uroboro
cd uroboro && pip install -e .
```

### 2. Capture Your Work
```bash
uro capture "Your development insight here"
```

### 3. Publish & Get Acknowledged
```bash
uro publish --blog
```

## 📋 Core Commands Reference

### `uro capture`
**Purpose**: Lightning-fast insight logging during development

```bash
# Basic capture
uro capture "Fixed memory leak in connection pool"

# With context (optional)
uro capture "Implemented WebSocket reconnection" --project my-app
```

**Features that actually work**:
- 10-second workflow
- Auto-git integration (when wanted)
- Project organization (when needed)
- Zero flow state interruption

### `uro publish`
**Purpose**: Transform captures into professional content

```bash
# Generate blog post
uro publish --blog

# Generate dev log
uro publish --devlog

# Generate social content
uro publish --social

# Custom timeframe
uro publish --blog --days 7
```

**Features that actually work**:
- Professional blog posts
- Social media content
- Technical dev logs
- Auto-voice detection (sounds like you, not AI)
- 2-minute generation time

### `uro status`
**Purpose**: Complete overview of your development pipeline

```bash
# See everything
uro status
```

**Shows you**:
- Recent captures across all projects
- Git integration status
- Content generation history
- Usage analytics (local only)
- System health

## 💡 Why Three Commands Work

### 🎯 Drunk User Test
If someone slightly impaired can't use your tool successfully in under 2 minutes, it's too complex. Three commands pass this test.

### ⚡ Focus Principle
When in doubt, improve what you have instead of adding new features. Every new feature is a chance to confuse users and break what works.

### 🔥 Acknowledgment Pipeline
One goal: help developers get acknowledged for their actual work. Everything else is distraction. We removed the distractions.

## 🏠 Local AI Setup

### Requirements
- **Python 3.8+**
- **Ollama** (for local AI)
- **Git** (for git integration)
- **16GB+ RAM** (recommended)

### Install Ollama
```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Pull a model (example)
ollama pull mistral:latest
```

uroboro automatically detects and uses your Ollama models. No configuration needed.

## 🔒 Privacy First

- **100% local processing** - No external API calls
- **Zero data collection** - Your insights stay on your machine  
- **Optional usage stats** - Local SQLite only, your control
- **No telemetry** - No phone home, no analytics
- **Full data ownership** - Export and analyze your own patterns

## 🚢 North Star Principle

**Three commands beats seventeen.** 

This isn't just philosophy - it's what we learned. We took a bloated 17-command CLI with 1,558 lines and surgical-strike removed everything that didn't directly contribute to developer acknowledgment.

**Result**: A tool that actually gets used because it respects your time and workflow.

## 🤝 Contributing

Found a bug in one of the core commands? Want to improve the core workflow? 

1. **Keep it simple** - If it adds complexity, we probably don't want it
2. **Drunk user test** - Can someone slightly impaired use it?
3. **Acknowledgment focus** - Does it help developers get noticed for their work?

## 📄 License

MIT License - Use it, modify it, ship it. Just get acknowledged for your work.

---

**uroboro**: The tool that documents itself while helping you document everything else 🐍

*Made by developers, for developers who deserve acknowledgment.*