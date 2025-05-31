# uroboro

*uroborouroborouroboro...*

> **🌐 Visit [uroboro.dev](https://uroboro.dev)** for the full experience, examples, and interactive demo.

An AI-powered content aggregation and generation system that transforms your daily development work into blog posts, social media content, and structured development logs using **local LLMs only**. Zero API costs, maximum privacy.

The name "uroboro" captures the recursive nature of this tool - development work feeds content creation, which feeds development insights, which feeds more content... where does one end and the other begin?

## 🌟 Key Features

- 🏠 **100% Local AI** - No API costs, complete privacy
- 💰 **$0/month operation** - One-time setup, zero ongoing costs  
- 🎯 **Multi-project tracking** - Captures work across all your projects
- ✍️ **5 writing styles** - Professional, technical, storytelling, minimalist, thought leadership
- 🚀 **Zero-friction workflow** - Single command captures, automatic generation

## 🤔 What Does It Actually Do?

**In Simple Terms**: You type `./capture.sh "Fixed that tricky bug"` and it automatically becomes part of tomorrow's blog post about your development progress.

**The Full Picture**:
- **Captures** your daily development work with single terminal commands
- **Aggregates** context from multiple projects, understanding what each project is about
- **Generates** polished content using local AI that knows your development style
- **Publishes** directly to your blog in the right format (MDX for Next.js)
- **Creates** social media hooks ready for Twitter/Bluesky
- **Maintains** searchable development history across all your projects

## 🚀 Quick Start

**Want to see how it works first?** Visit **[uroboro.dev](https://uroboro.dev)** for an interactive demo and detailed examples.

**Ready to install?** Follow the 5-minute setup below or check out the [Getting Started Guide](https://uroboro.dev#how-it-works).

## 🏗️ How It Works

### The Architecture
```
Development Work → Quick Capture → AI Processing → Published Content
     ↓               ↓               ↓               ↓
Terminal commands → .devlog files → Local LLM → Blog/Social/Devlog
```

### The Components
- **ContentAggregator**: Collects activity from configured projects and notes
- **ContentGenerator**: Uses local LLM to transform raw captures into polished content  
- **Project Context**: Each project has a `.devlog/README.md` explaining its purpose to the AI
- **Multi-format Output**: Blog posts (MDX), social hooks, development summaries

### The Data Flow
1. You work on projects and capture insights via `./capture.sh`
2. Captures go to `.devlog/YYYY-MM-DD-capture.md` files in each project
3. `generate_content.py` reads recent captures + project context
4. Local LLM (Mistral via Ollama) generates structured content
5. Blog posts save to `../qryzone/content/blog/` in MDX format
6. Social hooks display in terminal for copy/paste

## 💰 Costs of Operation

**TL;DR: $0 ongoing costs**

- **Cloud AI APIs**: None (everything runs locally)
- **Storage**: ~820KB for the tool + your text captures
- **Compute**: Uses your local machine via Ollama
- **Dependencies**: Python 3.8+ + Ollama (both free)

**What You Need to Pay For**: Nothing. The most expensive part is the electricity to run Ollama.

## 🚀 What's Needed to Get Started

### Prerequisites
```bash
# 1. Python 3.8+ (probably already have)
python3 --version

# 2. Install Ollama (if not already)
curl -fsSL https://ollama.ai/install.sh | sh

# 3. Pull a language model
ollama pull mistral:latest
```

### 5-Minute Setup
```bash
# Clone this repo
git clone <this-repo>
cd uroboro

# Run once to create default config
python3 src/aggregator.py

# Edit config to point to your projects
nano config/settings.json

# Create devlog directories in your projects
mkdir ~/my-project/.devlog

# Test it works
./capture.sh "Testing the pipeline"
python3 generate_content.py --preview
```

## 🤖 Does the AI Work Locally?

**Yes, 100% local**. The tool uses:

- **Ollama** as the local LLM server
- **Mistral** as the default model (you can change this)
- **No internet required** once set up
- **No API keys** or external services
- **Complete privacy** - your code/notes never leave your machine

### Supported Models
```bash
# Recommended (fast, good quality)
ollama pull mistral:latest

# Alternatives
ollama pull llama2:7b-chat      # Meta's model
ollama pull deepseek-r1:7b      # Coding-focused
ollama pull codellama:7b        # Code-specialized
```

## 📁 Where Does the Data Go?

### Input Data (Your Captures)
```
~/your-project/.devlog/
├── 2024-05-30-capture.md      # Your daily captures
├── 2024-05-31-capture.md      # Organized by date
└── README.md                  # Project context for AI
```

### Generated Content
```
uroboro/output/
├── daily-runs/                 # Raw activity JSON
│   └── activity_2024-05-30_21-30-00.json
└── knowledge-mining/           # AI analysis dumps
    └── archaeology-notes-2024-05-30.md

../qryzone/content/blog/       # Published blog posts
└── 2024-05-30-development-progress.mdx
```

### Configuration
```
uroboro/config/
└── settings.json              # Project paths, AI model, etc.
```

**Data Privacy**: Everything stays on your machine. No telemetry, no cloud uploads, no external API calls.

## 📊 Current Project Configuration

The tool is currently configured to monitor these projects:

| Project | Type | Role | Status |
|---------|------|------|--------|
| quantum-dice | game | Development project | Active |
| qryzone | website | Output channel (blog) | Active |
| notes | knowledge | Knowledge base | Active |
| panopticron | tool | Development project | Active |  
| uroboro | meta | Self-documenting | Active |

## 📝 Daily Usage

### Morning: Generate Yesterday's Content
```bash
python3 generate_content.py --title "Development Progress" --tags daily-update
```

### During Development: Quick Captures
```bash
./capture.sh "Implemented user authentication" quantum-dice
./capture.sh "Found interesting React pattern"
./capture.sh "Fixed critical payment bug" --tags bugfix critical
```

### Evening: Review and Publish
```bash
# See what content would be generated
python3 generate_content.py --preview

# Generate everything (blog + social + devlog)
python3 generate_content.py

# Custom weekly summary
python3 generate_content.py --days 7 --title "Week in Development"
```

## 🔧 Project Structure

```
uroboro/           # 820KB total
├── src/
│   ├── aggregator.py          # Collects activity across projects
│   └── processors/
│       └── content_generator.py # AI content generation
├── config/
│   └── settings.json          # Project paths and configuration
├── templates/                 # Content generation templates
├── output/                   # Generated content and logs
├── capture.sh                # Quick capture script
├── generate_content.py       # Main content generation script
└── test_llm.py              # Test local LLM integration
```

## ⚙️ Configuration Deep Dive

Edit `config/settings.json` to customize:

```json
{
  "notes_root": "~/notes",
  "llm_model": "mistral:latest",
  "projects": {
    "my-project": {
      "path": "~/projects/my-project",
      "type": "web|game|tool|meta|knowledge",
      "active": true,
      "role": "development|output_channel|knowledge_base",
      "description": "AI context about this project"
    }
  },
  "output_channels": {
    "blog": "qryzone",
    "social": "general"
  }
}
```

## 🧠 AI Context System

Each project can include `.devlog/README.md` with context for the AI:

```markdown
# Project: Quantum Dice Game

## Purpose
A web-based dice game that uses quantum mechanics for true randomness.

## Current Focus  
- WebGL visualizations of quantum states
- Real-world physics integration
- User experience improvements

## Technical Stack
React, Three.js, quantum API integration

## AI Instructions
When generating content about this project, emphasize the technical innovation
and learning journey rather than game mechanics.
```

## 🚀 Advanced Features

### Knowledge Mining
```bash
# Analyze your entire notes directory for insights
python3 generate_content.py --output knowledge --notes-path ~/notes

# Deep archaeological dig through knowledge base
python3 generate_content.py --output knowledge --mega-mining
```

### Multi-Project Insights
The AI automatically finds connections between projects:
- Shared technical patterns
- Cross-project learnings
- Recurring challenges and solutions

### Zero-Friction Workflow
- **Terminal integration**: Works from any project directory
- **Cursor integration**: Capture directly from your editor
- **Background processing**: Generate content while you work
- **Minimal interruption**: Designed for flow state preservation

## 🔮 Scaling & Future

**Immediate Extensions:**
- Voice capture integration (`./capture.sh` via speech-to-text)
- Git commit analysis (automatic capture from commit messages)
- Code change summarization (what files changed, why)
- Template customization (different AI personalities)

**Advanced Possibilities:**
- Multi-agent AI workflows (specialist agents for different content types)
- Cross-team collaboration (shared capture across team members)
- Technical documentation auto-generation
- Integration with project management tools

## 🤝 Contributing

This is a personal tool that grew organically. The architecture is simple and hackable:
- Add new content types in `ContentGenerator`
- Add new capture sources in `ContentAggregator`  
- Modify AI prompts in the template files
- Create new output channels in the config

**Want to contribute?** Check out the [issues](https://github.com/qry91/uroboro/issues) or visit [uroboro.dev](https://uroboro.dev) to get in touch.

## 📈 Success Metrics

After building this tool, the measurable outcomes:
- **Daily writing**: From sporadic to automated daily blog posts
- **Social presence**: Consistent content hooks for social media
- **Development visibility**: Clear record of daily progress across projects
- **Knowledge retention**: Searchable history of development insights
- **Content quality**: AI transforms rough notes into polished articles

The meta-aspect is compelling: using the tool to document building the tool creates a recursive content loop that embodies the "open garage door" philosophy.

---

## 🔗 Links & Resources

- **🌐 Website**: [uroboro.dev](https://uroboro.dev) - Interactive demo and examples
- **📚 Documentation**: [Getting Started](https://uroboro.dev#how-it-works) - Step-by-step setup guide
- **💬 Contact**: [hello@uroboro.dev](mailto:hello@uroboro.dev) - Questions or setup help
- **🐦 Updates**: [@qrynx](https://twitter.com/qrynx) - Follow for updates and insights
- **📝 Blog**: [qry.zone](https://qry.zone) - Real examples of content generated by uroboro

**License**: MIT - Build upon it, modify it, make it your own. 