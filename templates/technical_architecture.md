# Technical Architecture

## Built for Developers, By Developers

### System Overview
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Project A     │    │                  │    │                 │
│   .devlog/      │───▶│                  │    │                 │
│   - captures    │    │  ContentAggregator   │    │   ContentGenerator   │
└─────────────────┘    │                  │    │                 │
                       │  - Multi-project │    │  - Local LLM    │
┌─────────────────┐    │  - Git analysis  │───▶│  - Style system │
│   Project B     │    │  - Notes mining  │    │  - Templates    │
│   .devlog/      │───▶│                  │    │                 │
│   - context     │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                    │                      │
┌─────────────────┐                 │                      ▼
│   Daily Notes   │                 │            ┌─────────────────┐
│   ~/notes/      │────────────────▶│            │   Output        │
│   - insights    │                 │            │   - Blog posts  │
└─────────────────┘                 │            │   - Social      │
                                    ▼            │   - Devlogs     │
                         ┌──────────────────┐    └─────────────────┘
                         │   Knowledge      │
                         │   Mining         │
                         │   - Themes       │
                         │   - Patterns     │
                         └──────────────────┘
```

### Core Components

#### ContentAggregator
```python
class ContentAggregator:
    def collect_recent_activity(self, days=1):
        # Scans configured projects for:
        # - .devlog captures
        # - Git commit summaries  
        # - Modified files
        # - Cross-project patterns
        
    def mine_knowledge_base(self, deep_analysis=False):
        # Archaeological dig through notes
        # - Forgotten insights
        # - Technical evolution
        # - Learning patterns
```

#### ContentGenerator  
```python
class ContentGenerator:
    def __init__(self, voice="professional_conversational"):
        # 5 built-in writing styles:
        # - professional_conversational
        # - technical_deep
        # - storytelling  
        # - minimalist
        # - thought_leadership
        
    def generate_blog_post(self, activity, voice=None):
        # Local LLM processing with:
        # - Custom style instructions
        # - Brand voice preferences
        # - Project-specific context
```

### Local-First Philosophy

#### Why Not Cloud APIs?
| Aspect | Cloud APIs | Local LLMs |
|--------|------------|-------------|
| **Privacy** | Code sent to 3rd parties | Never leaves your machine |
| **Cost** | $50-200/month | $0/month |
| **Latency** | 2-5 seconds | 30-60 seconds |
| **Customization** | Limited prompts | Full model control |
| **Reliability** | Rate limits, outages | Offline capable |
| **Data retention** | Unknown policies | You own everything |

#### Performance Characteristics
- **Model size**: 4-7GB (fits in modern RAM)
- **Generation time**: 30-60 seconds for blog posts
- **Hardware requirements**: 16GB+ RAM recommended
- **Disk usage**: ~5GB for models + your content

### Privacy by Design

#### What Stays Local
✅ All source code and comments  
✅ Personal development notes  
✅ Project architecture details  
✅ Business logic and algorithms  
✅ Client information (if any)  

#### What Gets Processed
- Only the content you explicitly capture
- No automatic code scanning
- No telemetry or analytics
- No external network calls

### Extensibility

#### Adding New Projects
```json
{
  "projects": {
    "my-new-project": {
      "path": "~/projects/my-app",
      "type": "web",
      "active": true,
      "context": "React app with Node.js backend"
    }
  }
}
```

#### Custom Writing Voices
```json
{
  "voices": {
    "technical_influencer": {
      "description": "Technical content with industry insights",
      "prompt_additions": "Connect technical decisions to broader industry trends...",
      "example_phrases": ["Industry implications", "This approach"]
    }
  }
}
```

#### Output Channels
- **Blog platforms**: MDX, Markdown, HTML
- **Social networks**: Twitter threads, LinkedIn posts
- **Documentation**: Technical wikis, README updates
- **Analytics**: Development velocity tracking

---

**Open source, hackable, and designed to grow with your workflow.** 