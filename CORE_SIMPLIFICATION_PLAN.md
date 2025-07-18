# uroboro Core Simplification Plan 🎯

**Goal**: Transform uroboro from a 10,372-line "revolutionary AI framework" into a focused, arbtt-inspired development work capture tool.

## 🚨 Current State Analysis

### The Complexity Problem
- **10,372 lines of Go code** across 28 files
- **600+ lines** just for PostHog analytics
- **927 lines** for AI collaboration framework  
- Complex web server, database layers, session management
- Feature creep has obscured the core mission

### What Actually Works (Keep These)
- ✅ **Journey timeline visualization** - loved and actually useful
- ✅ **Smart project detection** - auto-detects git repos, package files
- ✅ **Auto-tagging** - content-based pattern detection
- ✅ **3-command structure** - capture, publish, status
- ✅ **File-based storage fallback** - works without complexity

### Cloud/Enterprise Bloat (Remove These)
- ❌ **PostHog analytics** - belongs in cloud version
- ❌ **AI collaboration framework** - not core workflow
- ❌ **Complex session management** - unnecessary overhead
- ❌ **Database complexity** - simple file storage first
- ❌ **Web server complexity** - minimize serving needs

## 🐍 The arbtt-Inspired Vision

### Core Philosophy
**"Invisible capture, flexible analysis"**

Like arbtt captures window activity silently, uroboro should capture development activity automatically with minimal friction.

### Simplified Architecture

```
uroboro-capture (daemon)     # Silent background capture
├── Git commit detection     # Auto-capture commits
├── File change monitoring   # Track active work
├── Context extraction       # Smart project detection
└── Timeline storage         # Simple file-based logs

uroboro-timeline (analysis)  # Flexible timeline generation  
├── Journey visualization    # Interactive timeline (keep!)
├── Content generation       # Blog/devlog from captures
├── Milestone detection      # Smart event categorization
└── Export formats           # Multiple output options
```

### The New 3-Command Workflow

```bash
# 1. Start invisible capture (like arbtt-capture)
uroboro-capture &            # Runs in background, captures everything

# 2. Generate content from captured work
uroboro timeline --blog --last-week
uroboro timeline --journey --days 30

# 3. Check what's been captured
uroboro status --summary
```

## 📊 Target Architecture

### Core Components (Keep Simple)
1. **Capture Engine** (~200 lines)
   - Git commit monitoring
   - File change detection
   - Context extraction
   - Timeline logging

2. **Timeline Generator** (~400 lines)
   - Journey visualization (existing code)
   - Content generation
   - Export formats

3. **Analysis Engine** (~200 lines)
   - Smart tagging
   - Milestone detection
   - Project summaries

**Total Target: ~800 lines** (down from 10,372)

### Storage Strategy
- **Primary**: Simple file-based timeline logs
- **Format**: JSON lines for easy parsing
- **Location**: `~/.uroboro/timeline/`
- **Backup**: Git-based versioning

## 🎯 Implementation Phases

### Phase 1: Core Extraction (Week 1)
- [ ] Extract journey visualization code
- [ ] Extract smart project detection  
- [ ] Extract auto-tagging logic
- [ ] Create minimal file-based storage
- [ ] Remove analytics/AI/web complexity

### Phase 2: Daemon Architecture (Week 2)
- [ ] Create `uroboro-capture` daemon
- [ ] Implement git commit monitoring
- [ ] Add file change detection
- [ ] Automatic context extraction
- [ ] Background timeline building

### Phase 3: Analysis Refinement (Week 3)
- [ ] Enhance timeline generation
- [ ] Improve content generation
- [ ] Refine milestone detection
- [ ] Polish journey visualization

### Phase 4: Testing & Polish (Week 4)
- [ ] Test automatic capture
- [ ] Validate timeline accuracy
- [ ] Ensure zero-friction workflow
- [ ] Performance optimization

## 🚀 Success Metrics

### The Drunk User Test
*Can someone slightly impaired successfully use uroboro in under 2 minutes?*

Target workflow:
```bash
# Install and forget
uroboro-capture &

# Weekly content generation
uroboro timeline --blog
```

### Code Quality Metrics
- **Lines of code**: <1,000 (down from 10,372)
- **Dependencies**: Minimal (remove PostHog, AI libs)
- **Startup time**: <100ms
- **Memory usage**: <10MB

### User Experience Metrics
- **Setup time**: <30 seconds
- **Daily interaction**: 0 seconds (automatic)
- **Content generation**: <2 minutes
- **Timeline accuracy**: >90%

## 🔒 What We're NOT Building

### Explicit Anti-Features
- ❌ **No AI coaching** - not our problem
- ❌ **No productivity analytics** - arbtt does this better
- ❌ **No team collaboration** - focus on individual workflow
- ❌ **No plugin architecture** - keep it simple
- ❌ **No cloud integration** - local-first always

### Cloud Version Separation
PostHog analytics, AI collaboration, and team features belong in a separate "uroboro-cloud" that interoperates with core uroboro.

## 🎭 The Honest Marketing Position

### Before (Current)
*"Revolutionary AI-powered development assistant with enterprise-grade analytics"*

### After (Target)
*"Automatic development work capture with beautiful timeline visualization"*

### The Positioning
- **Honest about scope**: "Solo developer tool, not enterprise platform"
- **Clear about automation**: "Works in background, generates content on demand"
- **Transparent about simplicity**: "Three commands, zero configuration"

## 🧭 Migration Strategy

### For Existing Users
1. **Data preservation**: Export existing captures to new format
2. **Feature parity**: Ensure journey visualization remains intact
3. **Gradual transition**: Keep old commands as aliases initially
4. **Clear communication**: Explain the focus shift

### For New Users
1. **Instant value**: Timeline visualization works immediately
2. **No setup friction**: Auto-detects projects and context
3. **Immediate content**: Can generate blog posts from day 1

## 📝 Implementation Notes

### Technical Decisions
- **Language**: Keep Go for performance and single-binary distribution
- **Storage**: JSON lines for human readability and tool compatibility
- **Web UI**: Minimal embedded server for timeline visualization only
- **Dependencies**: Audit and minimize external libraries

### Quality Assurance
- **Testing**: Focus on capture accuracy and timeline generation
- **Documentation**: Simple README with 3-command workflow
- **Distribution**: Single binary with embedded web assets

---

**Remember**: We're in a blue ocean. The opportunity is real. Stay focused on the core mission: *"Help developers get acknowledged for their actual work."*

The timeline visualization is the killer feature that differentiates us from generic productivity tools. Everything else should support that core experience.