# uroboro

**Track what you did and when. Local-first development context capture.**

## Install

```bash
go install github.com/QRY91/uroboro/cmd/...@latest
```

This installs both `uroboro` and `uro` (shorter alias).

## Commands

```bash
# Capture work insights
uro capture "Fixed auth bug in login flow"

# Quick captures with auto-tagging
uro d "JWT over sessions — stateless, scales horizontally"  # decision
uro b "waiting on backend API"                               # blocker
uro q "token revocation strategy?"                           # question

# Search past captures
uro search "auth" --project myapp --days 30

# See recent context (decisions, blockers, commits)
uro recap --days 7

# Interactive timeline (TUI)
uro timeline

# Activity summary
uro status

# Time reports for billing
uro report --days 7 --format markdown
```

## Claude Code Integration

Uroboro includes an MCP server for automatic context capture with Claude Code.

### Setup

```bash
claude mcp add uroboro --scope user -- uroboro mcp
```

### How It Works

Once configured, Claude automatically captures:
- **Decisions** when choosing between alternatives
- **Blockers** when work is stuck on dependencies
- **Questions** when deferring open issues

No manual capture needed. Run `uro recap` to see what happened.

### Available MCP Tools

| Tool | Purpose |
|------|---------|
| `uro_decision` | Record technical decisions |
| `uro_blocker` | Record blockers |
| `uro_question` | Record open questions |
| `uro_capture` | General capture |
| `uro_recap` | Get recent context |
| `uro_search` | Search past captures |

## Philosophy

- **Local-first** — SQLite database in `~/.local/share/uroboro/`
- **No cloud** — Your data stays on your machine
- **Tool-agnostic** — Works standalone or with Claude Code
- **Minimal friction** — Quick captures, powerful retrieval

## Data

All captures stored locally:
```
~/.local/share/uroboro/uroboro.sqlite
```

Export via `uro timeline --export` (JSON) or `uro report --format csv`.

## License

MIT
