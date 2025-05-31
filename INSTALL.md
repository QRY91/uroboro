# Installing uroboro CLI

Transform from scattered scripts to a unified `uroboro` command.

## Quick Install

```bash
# In your uroboro directory
pip install -e .

# Now you can use:
uroboro capture "Fixed WebSocket memory leak"
uroboro generate --blog --voice storytelling
uroboro status
```

## What Changes

**Before:**
```bash
./capture.sh "content"                    # Capture
python3 generate_content.py --blog        # Generate  
python3 voice_analyzer.py                 # Voice analysis
```

**After:**
```bash
uroboro capture "content"                 # Capture
uroboro generate --blog                   # Generate
uroboro voice                             # Voice analysis
uroboro mine --mega                       # Knowledge mining
uroboro status                            # Check what's happening
```

## All Commands

### Capture Development Insights
```bash
uroboro capture "Fixed authentication bug in login flow"
uroboro capture "Implemented real-time notifications" --project quantum-dice
uroboro capture "Discovered pattern in event handling" --tags architecture patterns
```

### Generate Content
```bash
uroboro generate                          # All content types from today
uroboro generate --blog --voice storytelling
uroboro generate --social --days 3       # Last 3 days for social posts
uroboro generate --devlog --preview      # Preview devlog without saving
```

### Voice & Style
```bash
uroboro voice                             # Analyze your writing style
uroboro generate --voice personal_excavated  # Use your trained voice
```

### Knowledge Mining
```bash
uroboro mine                              # Basic knowledge analysis
uroboro mine --mega                       # Deep archaeological expedition  
uroboro mine --path ~/notes/projects      # Mine specific directory
uroboro mine --preview                    # See analysis without saving
```

### Status & Monitoring
```bash
uroboro status                            # Quick overview
uroboro status --days 14 --verbose       # Detailed 2-week activity
```

## Migration

Your existing scripts still work! But the unified CLI is cleaner:

- ✅ Better help system (`uroboro generate --help`)
- ✅ Consistent flag naming across commands
- ✅ Easy to remember and share with others
- ✅ Professional CLI that feels like `git` or `docker`

## Backward Compatibility

Keep the old scripts around for a while:
- `capture.sh` → will call `uroboro capture` internally
- `generate_content.py` → still works for complex workflows

## Installation Methods

### Development Install (Recommended)
```bash
pip install -e .                         # Editable install for development
```

### Regular Install
```bash
pip install .                            # Regular install
```

### System Install
```bash
sudo pip install -e .                    # System-wide (use with caution)
```

### User Install
```bash
pip install --user -e .                  # User-local install
```

## Troubleshooting

**Command not found?**
```bash
# Check if installed
pip list | grep uroboro

# Reinstall
pip uninstall uroboro
pip install -e .
```

**Import errors?**
- Make sure you're in the uroboro directory
- Check that all dependencies are installed
- Run `python -c "from src.aggregator import ContentAggregator"` to test

**Old habits?**
- Keep using `./capture.sh` if you prefer
- Add `alias uro=uroboro` to your shell config for even shorter commands 