# Dev Content Pipeline

A centralized content aggregation and processing system for developers. Collects context from multiple projects, daily notes, and development activities to generate blog posts, social media content, and development logs.

## Features

- **Multi-project aggregation**: Collect context from all your active projects
- **Quick capture**: Rapid note-taking from terminal or editor
- **Structured output**: Generate devlogs, blog posts, and social content
- **Local-first**: Designed to work with local LLMs for cost-effective processing

## Quick Start

1. **Test basic functionality**:
   ```bash
   python src/aggregator.py
   ```

2. **Quick capture**:
   ```bash
   ./capture.sh "Fixed animation bug in quantum dice"
   ```

3. **Collect recent activity**:
   ```bash
   python src/aggregator.py collect
   ```

## Configuration

The system will auto-create `config/settings.json` on first run. Edit this file to:
- Set your notes directory
- Configure project paths
- Adjust content types

## Project Structure

```
dev-content-pipeline/
├── src/                 # Core Python modules
├── config/              # Configuration files
├── templates/           # Content templates
├── output/              # Generated content
└── examples/            # Usage examples
```

## Next Steps

- [ ] Add local LLM integration for content generation
- [ ] Implement git activity tracking
- [ ] Create blog post templates
- [ ] Add social media formatting 