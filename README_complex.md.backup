# uroboro 🐍

**The Unified Development Assistant**

*Formerly the "QRY Trinity" - wherewasi, examinator, and uroboro - now unified into one powerful tool*

Turn your development work into shareable content that gets you acknowledged for what you actually build. With smart project detection, content-based auto-tagging, and instant context extraction.

## 🤖 AI Collaboration Transparency

This project documentation and development has been enhanced through systematic AI collaboration following QRY Labs methodology:

- **Human-Centered Development**: All core functionality, architecture decisions, and strategic direction remain human-controlled
- **AI-Enhanced Documentation**: AI assistants help improve documentation quality and systematic presentation
- **Transparent Attribution**: AI collaboration is acknowledged openly as part of QRY's commitment to ethical technology use
- **Local-First Philosophy**: Uroboro itself uses local AI (Ollama) to maintain privacy and user control
- **Systematic Methodology**: AI collaboration follows structured procedures documented in `/ai/` directory

**Core Principle**: AI enhances human capability rather than replacing human judgment. The systematic approach to AI collaboration exemplifies uroboro's own philosophy of making quality work visible and professionally presentable.

## 🐍 Trinity Integration Complete

**uroboro** has successfully absorbed the QRY Trinity tools, unifying wherewasi's context detection, examinator's tagging intelligence, and uroboro's publishing power into one seamless experience.

**Trinity Features Now Built-In:**
- **Smart Project Detection** (from wherewasi): Auto-detects projects from git repos, package files, and directory names - zero configuration required
- **Content-Based Auto-Tagging** (from examinator): Analyzes capture content for patterns, detects action types (bugfix, feature, etc.), and identifies technology domains automatically  
- **Ripcord Functionality**: Instant context extraction to clipboard across all three commands - works cross-platform
- **Intelligent Context Linking**: Automatically connects related captures and provides rich development context

**The Philosophy Maintained:**
```bash
# Still just 3 commands - but now they're smarter
uroboro capture   # Query (with smart detection)
uroboro publish   # Refine (with intelligent tagging) 
uroboro status    # Yield (with rich context)
```

**Trinity Benefits:**
- **Reduced Complexity**: One codebase instead of three separate tools
- **Enhanced Intelligence**: Each command now leverages insights from all three original tools
- **Zero Configuration**: Smart detection eliminates setup requirements
- **Maintained Simplicity**: "3 commands beats 17" philosophy preserved while adding power

**Legacy Note**: Original trinity tools (wherewasi, examinator) are now legacy - uroboro contains all their functionality plus unified intelligence.

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

# Canvas timeline visualization
uro publish --journey
# Opens interactive timeline at http://localhost:8080

# Check status
uro status --days 7
```

## 🎯 North Star Workflow

1. **Capture** (10 seconds): Document insights as you work
2. **Publish** (2 minutes): Generate shareable content 
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

- **🎯 3 Core Commands** - capture, publish, status. Trinity absorbed, complexity eliminated.
- **🧠 Smart Project Detection** - Auto-discovers from git repos, package files, directory names
- **🏷️ Content-Based Auto-Tagging** - Analyzes patterns, detects action types, identifies tech domains
- **📋 Ripcord Functionality** - Instant context extraction to clipboard, cross-platform
- **⚡ 10-Second Capture** - Zero flow state interruption, now with intelligent context
- **📝 2-Minute Publish** - Content generation enhanced with trinity intelligence
- **🏠 100% Local AI** - Ollama, no external APIs
- **💰 $0/month costs** - No subscription, no usage fees
- **🎬 Canvas Timeline Visualization** - Interactive journey replay with `--journey` flag
- **📊 PostHog Analytics Integration** - Personal development insights (privacy-first)
- **🔒 Private by design** - Your code stays on your machine
- **🔗 Git integration** - When you need it, not when we think you should

## 🎬 Truth in Action

| Demo | What it shows | Why it matters |
|------|---------------|----------------|
| [**⚡ Core Workflow**](assets/uroboro_demo_core.gif) | capture → status → publish | The actual working product |
| [**🔗 Git Integration**](assets/uroboro_demo_git.gif) | Auto-capture commits | Optional but functional |
| [**🎬 Canvas Timeline**](http://localhost:8080) | `--journey` visualization | "Version control for your head" |
| [**🎬 Complete Overview**](assets/uroboro_demo.gif) | Full workflow | Tool that documents itself |

## 🧹 The Great CLI Cleanup

We started with **17 commands** and **1,558 lines** of bloated complexity. We removed 14 commands, kept 3 essential ones, and found that **focus beats features**: improve core functionality instead of adding new complexity.

### What We Kept (Core Commands + Trinity Intelligence)
- **`uroboro capture`** - 10-second insight logging with smart project detection and auto-tagging
- **`uroboro publish`** - Generate blog posts, social content, dev logs with enhanced context
- **`uroboro status`** - Complete pipeline overview with ripcord functionality

### What We Absorbed (Trinity Integration)
- ✅ **wherewasi context detection** - Now built into all commands
- ✅ **examinator tagging intelligence** - Automatic content analysis
- ✅ **Cross-tool communication** - Unified experience without ecosystem complexity
- ✅ **Smart fallbacks** - Zero-configuration project detection

### What We Cleaned Out (Feature Bloat)
- ❌ 14 unnecessary commands (kept the essential trinity)
- ❌ tamagoro egg system fantasy
- ❌ sensei learning complexity
- ❌ interview system bloat
- ❌ Q&A preparation overhead
- ❌ academic mode distraction
- ❌ 6 different project templates
- ❌ Voice training complexity
- ❌ **Multiple tool maintenance** - One codebase now handles everything

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

# Bonus: Visualize your journey
uroboro publish --journey
# Opens Canvas timeline at http://localhost:8080
```

## 🎬 Canvas Timeline Visualization

**"Version control for your head"** - Interactive timeline showing your development journey.

```bash
# Launch Canvas timeline visualization
uroboro publish --journey

# Customize timeline
uroboro publish --journey --days 7 --port 8080
```

**Features:**
- **Project Lanes** - Visual organization by project with consistent colors
- **Event Flow** - Chronological visualization of captures, commits, and milestones  
- **Smooth Performance** - Canvas-based rendering handles hundreds of events
- **Interactive** - Hover for details, click for event information
- **Professional Quality** - Pixel-perfect visualization ready for sharing

**Perfect for:**
- Retrospective reviews of development progress
- Understanding work patterns and context switches
- Creating compelling portfolio demonstrations
- Neurodivergent developers who benefit from visual pattern recognition

## 📊 PostHog Analytics Integration

**Personal development insights with privacy-first analytics.**

```bash
# Analytics happen automatically when enabled
# View insights at your PostHog dashboard
```

**Features:**
- **Development Pattern Tracking** - Capture frequency, session duration, productivity metrics
- **Privacy-First Design** - Enhanced privacy mode, configurable data sharing
- **Flow State Analysis** - Understand your most productive development patterns
- **Goal Correlation** - Connect work patterns with project milestones

**Configuration:**
- Set up through `uroboro config` command
- Optional feature - works great without it
- Respects privacy preferences and local-first philosophy

## 📋 Core Commands Reference

### `uroboro capture` / `uro capture`
**Purpose**: Lightning-fast insight logging during development

```bash
# Basic capture (file storage)
uroboro capture "Fixed memory leak in connection pool"
# or: uro capture "Fixed memory leak in connection pool"

# With context (optional)
uro capture "Implemented WebSocket reconnection" --project my-app

# Database storage (opt-in)
uro capture --db "Fixed auth timeout"  # Uses default XDG path
uro capture --db=myproject.sqlite --project backend "Optimized queries"
```

**Storage Options**:
- **File storage** (default): Daily markdown files in `~/.local/share/uroboro/daily/`
- **Database storage** (opt-in): SQLite database for cross-tool communication and querying

**Features that actually work**:
- 10-second workflow
- Auto-git integration (when wanted)
- Project organization (when needed)
- Zero flow state interruption
- Cross-tool communication via SQLite (when using `--db`)

### `uroboro publish` / `uro publish`
**Purpose**: Transform captures into professional content

```bash
# Generate blog post (from file storage)
uro publish --blog

# Generate dev log
uroboro publish --devlog

# Generate social content
uro publish --social

# Canvas timeline visualization
uro publish --journey

# Custom timeframe
uro publish --blog --days 7

# From database storage
uro publish --blog --db  # Uses default XDG path
uro publish --devlog --db=myproject.sqlite --days 14
```

**Storage Sources**:
- **File storage** (default): Reads from daily markdown files in `~/.local/share/uroboro/daily/`
- **Database storage** (opt-in): Reads from SQLite database using `--db` flag

**Features that actually work**:
- Blog posts from your dev work
- Social media content
- Technical dev logs
- Canvas timeline visualization (--journey flag)
- Voice matching (working on making it sound like you)
- 2-minute generation time
- Cross-tool data access via SQLite queries

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

**Three commands beats seventeen. One tool beats three.**

This isn't just philosophy - it's what we learned twice. First, we took a bloated 17-command CLI with 1,558 lines and surgical-strike removed everything that didn't directly contribute to developer acknowledgment. Then, we took three separate trinity tools and unified them into one powerful assistant without adding complexity.

**The Trinity Transformation:**
- **Before**: wherewasi + examinator + uroboro = 3 tools, 3 codebases, ecosystem complexity
- **After**: uroboro = 1 tool, 3 commands, trinity intelligence built-in
- **Philosophy Maintained**: Query-Refine-Yield trifecta maps perfectly to capture-publish-status

**Result**: A tool that actually gets used because it respects your time and workflow, now with the intelligence of three tools unified into one seamless experience.

## 🤖 AI-Collaborative Design

**Real testimonial from uroboro's developer:**

> *"I don't even type `uro capture` anymore. I just remind my AI assistant to capture work at natural breakpoints - context switches, completed logic blocks, solved problems. The AI reads the captures, understands the context, and helps generate much better summaries than I'd write alone. It's seamless collaboration through simple CLI design."*

**Why this works:**
- **Universal accessibility**: Any AI that can use command line can help
- **Natural workflow**: Capture at logical breakpoints, not arbitrary timers  
- **Collaborative improvement**: Human domain expertise + AI writing skills
- **Future-proof**: Works with any LLM, not tied to specific AI services

The CLI design makes uroboro naturally compatible with AI assistance, creating a feedback loop where better captures lead to better content generation.

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