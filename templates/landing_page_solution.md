# How uroboro Works

## From Development Work to Published Content in 3 Steps

### Step 1: Capture (5 seconds)
```bash
./capture.sh "Implemented real-time notifications using WebSockets"
```
- Quick terminal command during development
- No interruption to your flow state
- Captures context while it's fresh in your mind

### Step 2: Aggregate (Automatic)
uroboro collects your captures across all projects:
- Git commit summaries
- Daily development notes  
- Project-specific insights
- Cross-project patterns

### Step 3: Generate (AI-Powered)
Local AI transforms your raw captures into:
- **Blog posts** - Professional, engaging articles
- **Social hooks** - Twitter-ready content threads
- **Technical docs** - Architecture decisions explained
- **Progress summaries** - Weekly/monthly development reports

## The Architecture
```
Daily Dev Work → Quick Capture → Local AI → Published Content
     ↓               ↓               ↓               ↓
Terminal commands → .devlog files → Mistral → Blog/Social/Devlog
```

## Why Local AI?
- **Privacy** - Your code never leaves your machine
- **Cost** - $0/month after setup (no API fees)
- **Speed** - No network latency or rate limits
- **Control** - Customize the AI behavior completely
- **Reliability** - Works offline, no service dependencies

## Example Transformation

**Your capture:**
```
Fixed memory leak in the WebSocket connection pool. 
Turns out we weren't properly cleaning up event listeners 
when connections closed unexpectedly.
```

**Generated blog content:**
> "Today I discovered something interesting about WebSocket lifecycle management. While debugging what seemed like a simple memory issue, I uncovered a subtle but important pattern about event listener cleanup that every real-time application developer should know..."

---

**The magic happens while you sleep. Wake up to fresh content about yesterday's work.** 