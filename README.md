# uroboro 🐍

**The Self-Documenting Content Pipeline**

Transform your development insights into professional content that gets you acknowledged for your actual work.

## 🚀 Quick Start (Go CLI - Primary)

```bash
# Core workflow - 3 commands, that's it
uroboro capture "Fixed auth timeout - cut query time from 3s to 200ms"
# or: uro capture "Fixed auth timeout - cut query time from 3s to 200ms"

uroboro publish --blog --format html
# or: uro publish --blog --format html

uroboro status
# or: uro status
```

### Format Support
- `--format markdown` (default)
- `--format html` (with styling)  
- `--format text` (clean output)

### Examples
```bash
# Capture insights
uroboro capture "Implemented OAuth2 with JWT tokens"
uro capture "Reduced bundle size by 40% using tree shaking" --project frontend

# Generate content  
uro publish --blog --preview
uroboro publish --blog --format html --title "This Week's Wins"
uro publish --devlog

# Check status
uro status --days 7
```

## 🎯 North Star Workflow

1. **Capture** (10 seconds): Document insights as you work
2. **Publish** (2 minutes): Generate professional content 
3. **Get acknowledged**: Share your expertise and get noticed

No feature creep. No complexity. Just results.

## 📦 Installation

### Option 1: Go Install (Recommended)
```bash
go install github.com/QRY91/uroboro@latest
# Binary available as 'uroboro' in your $GOPATH/bin
```

### Option 2: Clone & Build
```bash
git clone https://github.com/qry91/uroboro
cd uroboro
# Use the pre-built binary (recommended for development)
./uroboro capture "test"

# Or build from source
go build -o uro ./cmd/uroboro
```

### Option 3: Direct Download
Download the latest binary from [Releases](https://github.com/QRY91/uroboro/releases).

## 🔧 Development

### Go Implementation
- **Fast**: Sub-second startup
- **Clean**: No dependencies beyond Go stdlib + Ollama
- **Complete**: Full format support, capture, publish, status
- **Tested**: Comprehensive unit test coverage for all core functionality

### Quality Assurance
```bash
# Run all tests
go test ./internal/...

# Run tests with coverage
go test -race -coverprofile=coverage.out ./internal/...

# Build both binaries
go build -o uroboro ./cmd/uroboro && ln -sf uroboro uro
```

### CI Pipeline
- **Automated testing** on every push/PR
- **Unit tests** for all 3 core commands
- **Integration tests** with XDG compliance verification  
- **Build verification** for both `uroboro` and `uro` binaries
- **Quality gates** preventing regressions
- **Coverage reporting** to maintain code quality

The pipeline ensures no "amateur hour" regressions - all 3 commands must work, every release.

### Historical Reference
Python reference implementation available in `archive/python-reference/` for archaeological purposes.

## 🎮 VSCode Extension

The extension automatically finds and uses the Go binary. Install from `vscode-extension/`.

## 🧭 Philosophy

> "I solve practical problems. Use a gun. If that don't work, use more gun." - TF2 Engineer

- **Practical over philosophical** - tools that work, not theories
- **More gun principle** - improve core features, don't add new ones  
- **Ship working solutions** - real impact over impressive demos

Three commands. That's it. 🎯

## 🎯 The Core Workflow

**Three commands beats seventeen.** We discovered this by removing 2,497 lines of outdated code (60.8% of the codebase), cutting 14 commands, and keeping only what delivers acknowledgment for your actual work.

```bash
# 1. Capture your work (10 seconds)
uroboro capture "Fixed auth timeout - cut query time from 3s to 200ms"
# or: uro capture "Fixed auth timeout - cut query time from 3s to 200ms"

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
- **`uroboro capture`** - 10-second insight logging
- **`uroboro publish`** - Generate blog posts, social content, dev logs  
- **`uroboro status`** - Complete pipeline overview

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
# Option 1: Go install (creates 'uroboro' command)
go install github.com/QRY91/uroboro@latest

# Option 2: Clone and use (creates './uroboro' command)
git clone https://github.com/qry91/uroboro && cd uroboro
```

### 2. Capture Your Work
```bash
# If installed via go install:
uroboro capture "Your development insight here"

# If using cloned repo:
./uroboro capture "Your development insight here"
```

### 3. Publish & Get Acknowledged
```bash
# If installed via go install:
uroboro publish --blog

# If using cloned repo:
./uroboro publish --blog
```

## 📋 Core Commands Reference

### `uroboro capture` / `uro capture`
**Purpose**: Lightning-fast insight logging during development

```bash
# Basic capture
uroboro capture "Fixed memory leak in connection pool"
# or: uro capture "Fixed memory leak in connection pool"

# With context (optional)
uro capture "Implemented WebSocket reconnection" --project my-app
```

**Features that actually work**:
- 10-second workflow
- Auto-git integration (when wanted)
- Project organization (when needed)
- Zero flow state interruption

### `uroboro publish` / `uro publish`
**Purpose**: Transform captures into professional content

```bash
# Generate blog post
uro publish --blog

# Generate dev log
uroboro publish --devlog

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

### `uroboro status` / `uro status`
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
- **Ollama** (for local AI)
- **Git** (for git integration)
- **8GB+ RAM** (recommended for AI models)

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