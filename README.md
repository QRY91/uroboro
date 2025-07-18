# uroboro 🐍

**Automatic development work capture with beautiful timeline visualization**

Turn your development work into shareable content that gets you acknowledged for what you actually build.

## 🎯 The Core Problem

Developers do amazing work that goes unnoticed. You solve complex problems, make smart decisions, and build valuable features - but there's no record of the thinking behind it.

**uroboro captures your actual work automatically and generates professional content from it.**

## ⚡ 3-Command Workflow

```bash
# 1. Capture insights as you work (10 seconds)
uroboro capture "Fixed database timeout - cut query time from 3s to 200ms"

# 2. Generate content from your work (2 minutes)
uroboro publish --blog
uroboro publish --journey  # Interactive timeline visualization

# 3. Check your progress
uroboro status
```

## 🎬 Timeline Visualization (The Killer Feature)

**See your development work as a beautiful interactive timeline:**

```bash
uroboro publish --journey --days 30
# Opens http://localhost:8080 with your development timeline
```

- **Visual timeline** of your actual work
- **Smart milestones** detection
- **Project breakdowns** and patterns
- **Export options** for sharing

This is what makes uroboro different from productivity apps - it shows your **real development journey**.

## 🚀 Quick Start

### Install
```bash
go install github.com/QRY91/uroboro/cmd/uroboro@latest
```

### Use
```bash
# Start capturing your work
uroboro capture "Your development insight here"

# Generate a blog post from recent work
uroboro publish --blog --days 7

# See your timeline
uroboro publish --journey
```

## ✨ What Makes This Different

- **Automatic project detection** - no configuration needed
- **Smart tagging** - recognizes features, bugfixes, decisions
- **Timeline visualization** - see your work as a story
- **Zero friction** - designed for busy developers
- **Local-first** - your data stays on your machine

## 🧹 Simplified Architecture

**We removed the bloat:** 10,372 → 6,449 lines of code (38% reduction)

**Removed:**
- ❌ Analytics dashboards
- ❌ AI coaching systems  
- ❌ Complex workflow management
- ❌ Enterprise features

**Kept what works:**
- ✅ Timeline visualization
- ✅ Smart project detection
- ✅ Content generation
- ✅ 3-command simplicity

## 🎯 Philosophy

**"Invisible capture, flexible analysis"** - inspired by [arbtt](https://arbtt.nomeata.de/)

Like arbtt captures window activity silently, uroboro captures development insights with minimal friction. The magic happens when you analyze the data later.

## 📊 Example Output

**Timeline JSON:**
```json
{
  "events": [
    {
      "timestamp": "2025-01-15T10:30:00Z",
      "content": "Fixed database timeout - cut query time from 3s to 200ms",
      "project": "api-server",
      "eventType": "bugfix",
      "importance": 2
    }
  ],
  "projects": [
    {
      "name": "api-server",
      "eventCount": 15,
      "color": "#FF6B6B"
    }
  ]
}
```

**Generated Blog Post:**
```markdown
# This Week's Development Wins

## Performance Optimization
Fixed a critical database timeout issue that was affecting 10k+ users. 
The solution involved optimizing the query structure, reducing execution 
time from 3 seconds to 200ms - a 93% improvement.

## Technical Details
- Identified N+1 query pattern in user dashboard
- Implemented eager loading for related models
- Added database indexes for commonly filtered fields
```

## 🔒 Privacy & Data

- **Local-first**: Everything runs on your machine
- **No tracking**: No analytics, no telemetry
- **Your data**: SQLite database in `~/.local/share/uroboro/`
- **Simple export**: JSON format for portability

## 🤝 Contributing

Found a bug? Have an idea? 

**Before adding features:** Read the [North Star document](uroboro-north-star.md) - we're focused on the core mission and avoid feature creep.

## 🎪 Anti-Marketing

**What this ISN'T:**
- A productivity dashboard
- An AI-powered assistant  
- A team collaboration tool
- A project management system

**What this IS:**
- A tool that documents your actual work
- A timeline generator for your development journey
- A way to get acknowledged for what you build

---

*"The only tool that turns your actual development work into professional content"*

**Built by a developer, for developers. No enterprise BS, no revolutionary AI claims - just a useful tool that works.**