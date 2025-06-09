# Demo Recording Guide for uroboro Landing Page
## The Unified Development Assistant

## Mobile-Friendly Recording Specs

### Terminal Settings
- **Font:** JetBrains Mono or SF Mono (same as landing page)
- **Font Size:** 16-18px minimum (readable on mobile)
- **Terminal Width:** 80-100 characters max
- **Terminal Height:** 20-25 lines max
- **Background:** Dark theme (#1e1b2e or similar to match landing page)

### Recording Resolution
- **Target:** 1200x800px (16:10 ratio)
- **Mobile consideration:** Text should be readable when scaled to 300-400px width
- **Export:** GIF with optimized palette, <1MB per demo

### Command Examples to Record

#### Core Workflow Demo (`uroboro_demo_core.gif`)
```bash
# Show intelligent automation in action
$ uroboro capture "Implemented WebSocket real-time notifications"
🧠 Smart detection: Node.js project detected
🏷️ Auto-tagged: [feature, websocket, real-time]
✓ Captured with intelligent analysis

$ uro -s
📊 Recent Activity: 
   - 3 captures today
   - WebSocket implementation (auto-tagged: feature)
   - Auth optimization (auto-tagged: performance)
📋 Ripcord: Context copied to clipboard

$ uroboro publish --blog
🚀 Generated: "Building Real-Time Web Applications"
   → Enhanced with project context and smart categorization
```

#### Git Integration Demo (`uroboro_demo_git.gif`)
```bash
$ git commit -m "Add WebSocket connection pooling"
[main abc1234] Add WebSocket connection pooling

$ uroboro capture --auto-git
🧠 Project context: React + Node.js detected
🏷️ Commit analysis: Performance optimization
✓ Auto-captured with intelligent analysis: "WebSocket connection pooling implementation"

$ uro -s
📊 Git Integration:
   - 5 commits captured this week
   - 2 features documented with smart tags
   - Cross-project patterns detected
```

#### Complete Overview Demo (`uroboro_demo.gif`)
```bash
$ uroboro capture "Fixed memory leak in connection handler"
🧠 Smart detection: Performance issue identified
🏷️ Auto-tagged: [bugfix, performance, memory]
✓ Captured with intelligent analysis

$ uro -c "Added rate limiting to prevent abuse" 
🧠 Context linking: Security enhancement pattern
🏷️ Auto-tagged: [feature, security, rate-limiting]
✓ Enhanced capture with cross-reference

$ uroboro status
📊 Development Pipeline:
   - 2 captures today (auto-categorized)
   - Pattern detected: Performance + Security focus
   - Ready for enhanced publishing
📋 Ripcord: Full context ready

$ uro -p --blog
🚀 Generated blog post: "Performance and Security: A Dual Approach"
   → Content enriched with project context and intelligent linking
```

### Recording Tips

1. **Timing:** 2-3 seconds between commands
2. **Pauses:** 1-2 seconds to show intelligent automation output
3. **Loop:** 3-5 second pause before restarting
4. **Colors:** Use terminal colors that match landing page theme
5. **Commands:** Show both `uroboro` and `uro` formats
6. **Output:** Include smart features (detection, auto-tagging, ripcord)
7. **Intelligence:** Highlight zero-configuration and automatic enhancements
8. **Context:** Show cross-project pattern recognition and linking

### File Naming
- `uroboro_demo_core.gif` - Core workflow (capture, status, publish)
- `uroboro_demo_git.gif` - Git integration features
- `uroboro_demo.gif` - Complete overview demonstration

### Quality Check
- [ ] Text readable at 400px width
- [ ] Commands show both formats (uroboro/uro)
- [ ] Demonstrates intelligent automation features
- [ ] Shows smart detection, auto-tagging, and ripcord functionality
- [ ] Highlights zero-configuration intelligence
- [ ] File size under 1MB
- [ ] Loops smoothly
- [ ] Colors match landing page theme
- [ ] Intelligent automation clearly visible in output