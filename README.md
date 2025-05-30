# Dev Content Pipeline

*uroborouroborouroboro...*

An AI-powered content aggregation and generation system for developers. Automatically transforms your daily development work into blog posts, social media content, and structured development logs using local LLMs.

The name "uroboro" captures the recursive nature of this tool - development work feeds content creation, which feeds development insights, which feeds more content... where does one end and the other begin?

## 🚀 What It Does

- **Captures development activity** across multiple projects with zero friction
- **Aggregates context** from project devlogs, daily notes, and git activity
- **Generates content** using local AI models (Ollama)
- **Publishes automatically** to your blog (qryzone format)
- **Creates social media hooks** ready for posting
- **Maintains development history** with searchable context

## ✨ Features

- **Multi-project aggregation**: Collect context from all your active projects
- **AI-powered content generation**: Local LLM integration with Mistral/Llama models
- **Zero-setup capture**: Quick terminal commands for logging development work
- **Structured output**: Blog posts, devlogs, social content in proper formats
- **Automatic publishing**: Direct integration with Next.js/MDX blog structure
- **Local-first**: Works entirely offline with local LLMs
- **Cross-project insights**: Discover connections between different projects

## 🛠 Installation & Setup

### Prerequisites
- Python 3.8+
- [Ollama](https://ollama.ai) installed with a model (mistral, llama2, etc.)

### Quick Start

1. **Clone and setup**:
   ```bash
   git clone <this-repo>
   cd dev-content-pipeline
   python3 src/aggregator.py  # Creates default config
   ```

2. **Configure your projects** in `config/settings.json`:
   ```json
   {
     "projects": {
       "my-project": {
         "path": "~/projects/my-project",
         "type": "web",
         "active": true
       }
     }
   }
   ```

3. **Create devlog folders** in your projects:
   ```bash
   mkdir ~/projects/my-project/.devlog
   ```

## 📝 Daily Usage

### Capture Development Work
```bash
# Project-specific captures
./capture.sh "Fixed authentication bug" my-project
./capture.sh "Redesigned user dashboard" my-project

# General development notes  
./capture.sh "Great insight about React state management"

# With tags
./capture.sh "Implemented caching layer" my-project --tags performance optimization
```

### Generate Content
```bash
# Generate everything (blog + social + devlog)
python3 generate_content.py

# Specific content types
python3 generate_content.py --output blog --title "Weekly Development Update"
python3 generate_content.py --output social
python3 generate_content.py --output devlog

# Custom options
python3 generate_content.py --days 7 --title "Week in Review" --tags development AI automation
```

### Content Output

**Blog posts** → Saved to `../qryzone/content/blog/` in MDX format  
**Social hooks** → Ready-to-post Twitter/social content  
**Devlog summaries** → Structured development progress  
**Raw activity** → JSON files in `output/daily-runs/`

## 🏗 Project Structure

```
dev-content-pipeline/
├── src/
│   ├── aggregator.py           # Core aggregation logic
│   └── processors/
│       └── content_generator.py # AI content generation
├── config/
│   └── settings.json           # Project configuration
├── templates/                  # Content templates
├── output/                     # Generated content and logs
├── capture.sh                  # Quick capture script
├── generate_content.py         # Main content generation
└── test_llm.py                # LLM integration testing
```

## ⚙️ Configuration

Edit `config/settings.json`:

```json
{
  "notes_root": "~/notes",
  "llm_model": "mistral:latest",
  "projects": {
    "project-name": {
      "path": "~/path/to/project",
      "type": "web|game|tool|etc",
      "active": true,
      "role": "output_channel|knowledge_base|etc"
    }
  },
  "output_channels": {
    "blog": "qryzone",
    "social": "general"
  }
}
```

## 🤖 AI Integration

Uses local LLMs via Ollama for:
- **Content summarization**: Extracting key insights from development activity
- **Blog post generation**: Creating engaging, well-structured articles
- **Social media hooks**: Generating tweet-ready content with hashtags
- **Cross-project analysis**: Finding connections between different projects

### Supported Models
- `mistral:latest` (recommended)
- `llama2:7b-chat`
- `deepseek-r1:7b`
- Any Ollama-compatible model

## 📊 Example Workflow

```bash
# Morning: Generate yesterday's content
python3 generate_content.py --title "Development Progress" --tags daily-update

# During development: Quick captures
./capture.sh "Implemented user authentication system" web-app
./capture.sh "Discovered interesting pattern in state management"
./capture.sh "Fixed critical bug in payment processing" web-app

# Evening: Weekly summary
python3 generate_content.py --days 7 --output blog --title "Week in Development"
```

## 🔧 Advanced Features

### Multi-Project Aggregation
Automatically collects and correlates activity across all your projects, finding technical connections and shared patterns.

### AI-Powered Insights
Local LLM analyzes your development patterns to:
- Identify recurring challenges
- Suggest next steps
- Extract technical insights
- Generate engaging content narratives

### Zero-Friction Capture
Designed for minimal interruption to your development flow. Quick terminal commands that take seconds.

## 🚀 Scaling & Future Features

**Immediate Extensions**:
- Voice capture integration
- Git commit analysis
- Code change summarization
- Project ignore patterns
- Template customization

**Advanced Possibilities**:
- Multi-agent AI workflows
- Automated social media posting
- Technical documentation generation
- Cross-team collaboration features
- Analytics and trend analysis

## 🤝 AI Assistant Integration

Each project can include a `.devlog/README.md` with context for AI assistants:

```markdown
# Project: My Web App
## Current Focus: User authentication and dashboard
## Tech Stack: React, Node.js, PostgreSQL
## Recent Challenges: State management complexity
## Capture Usage: ./capture.sh "description" my-web-app
```

## 📈 Benefits

- **Consistent documentation** of development progress
- **Automated content creation** for blogs and social media
- **Historical context** for project decisions
- **Cross-project learning** and pattern recognition
- **Reduced friction** for sharing development insights
- **Professional portfolio building** through regular content

## 🛡 Privacy & Local-First

- All processing happens locally using Ollama
- No data sent to external APIs (unless you choose to)
- Full control over your development data
- Works offline once models are downloaded

---

*uroborouroborouroboro...* Transform your daily development work into a content engine! 🚀 