# Demo Recording Guide for uroboro Landing Page

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
# Show both command formats
$ uroboro capture "Implemented WebSocket real-time notifications"
✓ Captured development insight

$ uro -s
📊 Recent Activity: 
   - 3 captures today
   - WebSocket implementation
   - Auth optimization

$ uroboro publish --blog
🚀 Generated: "Building Real-Time Web Applications"
   → Ready for publishing
```

#### Git Integration Demo (`uroboro_demo_git.gif`)
```bash
$ git commit -m "Add WebSocket connection pooling"
[main abc1234] Add WebSocket connection pooling

$ uroboro capture --auto-git
✓ Auto-captured from latest commit: "WebSocket connection pooling implementation"

$ uro -s
📊 Git Integration:
   - 5 commits captured this week
   - 2 features documented
```

#### Complete Overview Demo (`uroboro_demo.gif`)
```bash
$ uroboro capture "Fixed memory leak in connection handler"
✓ Captured: Memory optimization insight

$ uro -c "Added rate limiting to prevent abuse" 
✓ Captured: Security enhancement

$ uroboro status
📊 Development Pipeline:
   - 2 captures today
   - Ready for publishing

$ uro -p --blog
🚀 Blog post: "Performance and Security: Two Critical Fixes"
   → Professional content ready
```

### Recording Tips

1. **Timing:** 2-3 seconds between commands
2. **Pauses:** 1-2 seconds to show output
3. **Loop:** 3-5 second pause before restarting
4. **Colors:** Use terminal colors that match landing page theme
5. **Commands:** Show both `uroboro` and `uro` formats
6. **Output:** Include realistic success messages and status displays

### File Naming
- `uroboro_demo_core.gif` - Core workflow (capture, status, publish)
- `uroboro_demo_git.gif` - Git integration features
- `uroboro_demo.gif` - Complete overview demonstration

### Quality Check
- [ ] Text readable at 400px width
- [ ] Commands show both formats (uroboro/uro)
- [ ] Demonstrates current features
- [ ] File size under 1MB
- [ ] Loops smoothly
- [ ] Colors match landing page theme 