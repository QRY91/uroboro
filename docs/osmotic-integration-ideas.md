# Osmotic-Uroboro Integration: Interview Mode

**Status**: Conceptual Design  
**Date**: June 2025  
**Purpose**: Turn world awareness into documented insights

## 🧠 The Vision

**Core Flow**: World briefing → Structured reflection → Captured insights

```bash
# Uroboro-initiated (preferred approach)
uroboro interview --source=osmotic --template=world-briefing

# What happens:
# 1. Calls osmotic status internally for briefing 
# 2. Shows secret-agent style 30-second briefing
# 3. Times out automatically (anti-engagement)
# 4. Asks contextual reflection questions
# 5. Captures insights with world-state metadata
```

## 🔄 Integration Architecture

### Osmotic Side (Minimal)
- Provide structured JSON output for briefings
- Keep focus on anti-engagement "get to work" energy
- No interview complexity in osmotic itself

### Uroboro Side (Interview Engine)
- `uroboro interview --source=X` capability
- Context-aware question generation
- Integration with capture system
- Template system for different briefing types

## 🎯 Smart Interview Questions

**Pattern-Contextual:**
- "Military tensions detected in Europe - how might this affect your supply chain work?"
- "Economic uncertainty trending upward - any implications for your project timelines?"
- "Tech regulation increasing - should this influence your architecture decisions?"

**Generic Reflection:**
- "What surprised you in this briefing?"
- "Any patterns that connect to problems you're solving?"
- "World context that might change your priorities this week?"

## 🛠️ Technical Implementation

### Data Flow
```
Osmotic JSON → Uroboro Interview → User Reflection → Capture System
```

### API Design
```bash
# Osmotic provides structured output
osmotic status --format=json --secret-agent
{
  "patterns": [...],
  "confidence": {...},
  "regions": [...],
  "timeout": 30,
  "briefing_id": "2025-06-06-morning"
}

# Uroboro consumes and extends
uroboro interview --source=osmotic --template=world-briefing
# → Internal call to osmotic
# → Display briefing with timeout
# → Generate contextual questions
# → Capture insights with metadata
```

## 🚀 Portfolio Value

**What This Demonstrates:**
- **Systems thinking** - Tools working together, not in isolation
- **Workflow intelligence** - Transform passive consumption into active insights
- **Anti-engagement design** - Even reflection is structured and time-bounded
- **Information architecture** - Turn data → patterns → insights → documented wisdom

## 🌊 Broader Ecosystem Integration

**The Full Pipeline:**
- **Osmotic** → World awareness (anti-engagement briefings)
- **Uroboro** → Insight capture and documentation 
- **WhereWasI** → Project context and handoffs
- **Examinator** → Documentation processing
- **Content Pipeline** → Transform insights into shareable content

**Vision**: Complete information manipulation and distribution system for focused builders.

## 🎭 Implementation Phases

### Phase 1: Basic Integration
- Osmotic JSON output format
- Uroboro interview command foundation
- Simple question templates

### Phase 2: Smart Questions  
- Pattern-aware question generation
- Context integration from other tools
- Historical insight tracking

### Phase 3: Full Pipeline
- Integration with documentation projects
- Content generation from insights
- Cross-tool context awareness

## 💡 Future Extensions

**Multi-Source Interviews:**
```bash
uroboro interview --source=osmotic,wherewasi --template=project-briefing
# → World state + Project context + Reflection questions
```

**Template System:**
- `world-briefing` - Global awareness reflection
- `project-handoff` - Context switching insights  
- `weekly-review` - Accumulated pattern analysis
- `decision-point` - Pre-decision context gathering

---

**Core Philosophy**: Turn information consumption into wisdom accumulation.  
**Anti-Engagement**: Even reflection is structured, time-bounded, and purpose-driven. 

## 🔄 Refined Implementation: Extend Capture Command

**Status**: Updated Design - Extend existing rather than add new  
**Rationale**: Stay true to 3-command philosophy, avoid feature bloat

### Extended Capture Command
```bash
# Normal capture (unchanged)
uroboro capture "Fixed auth timeout"

# Source-driven capture with integrated briefing + reflection
uroboro capture --source=osmotic --template=world-briefing
# → Calls osmotic internally
# → Shows 30-sec briefing  
# → Prompts contextual questions
# → Captures reflection insights
```

### Implementation Pattern
**Add to existing `CaptureService`**:
- `--source` flag triggers external tool integration
- `--template` controls question generation
- Still uses same `captureToDatabase()` / `captureToFile()` backend
- Metadata gets tagged with source context

**Data Flow**:
```go
if source == "osmotic" {
    briefingData := callOsmotic()
    displayBriefing(briefingData)
    insights := promptReflection(briefingData) 
    return c.Capture(insights, project, "osmotic,world-briefing")
}
```

### Why This Approach Wins
- **Philosophy preserved**: Still 3 commands, just extended capture functionality
- **Extensible**: `--source=wherewasi`, `--source=examinator` naturally fit later
- **Backward compatible**: Normal capture usage unchanged
- **Local-first**: Process communication, no APIs
- **Practical over philosophical**: Extend what works rather than add complexity

**Perfect fit with uroboro's "practical over philosophical" approach** - extend existing functionality rather than introduce new commands for one bespoke integration.