# uroboro

**Track what you did and when. Local-first development context capture.**

## Install

```bash
go install github.com/QRY91/uroboro/cmd/...@latest
```

This installs both `uroboro` (full) and `uro` (shorter alias with core commands).

## Commands

```bash
# Capture work insights
uro capture "Fixed auth bug in login flow"

# Quick captures with auto-tagging
uro d "JWT over sessions — stateless, scales horizontally"  # decision
uro b "waiting on backend API"                               # blocker
uro q "token revocation strategy?"                           # question

# Retroactive capture (backdate entries)
uro capture "Shipped v2 auth" --time "2024-01-15 14:30"

# Search past captures
uro search "auth" --project myapp --days 30

# See recent context (decisions, blockers, commits)
uro recap --days 7
uro recap --days 7 --branch feature/auth --brief

# Interactive timeline (TUI)
uro timeline

# Web timeline (scrollable, detailed)
uro web --port 8080

# Graph view (auto-scaled overview, fits screen)
uro graph --days 90

# Activity summary
uro status

# Time reports for billing
uro report --days 7 --format markdown

# Extract style signals from git + uroboro captures (JSONL)
uro distill --days 180 --correlate --out style-data.jsonl

# Analyze Claude Code prompting patterns
uro prompt-profile --days 30
uro prompt-profile --extract --out prompts.jsonl
```

## Multi-device

Captures record which machine they came from. Move data between machines with export/import:

```bash
# Export captures as JSONL
uroboro export --since 2024-01-01 --out captures.jsonl
uroboro export --project myapp --days 30  # to stdout

# Import from another machine (deduplicates by content+project)
uroboro import captures.jsonl --machine cambrian
uroboro import --dry-run captures.jsonl   # preview first

# Pull live from a remote machine (requires uroboro installed there)
ssh cambrian uroboro export --since 2024-01-01 | uroboro import --machine cambrian
```

Override the detected hostname with `UROBORO_MACHINE=myhost` (useful in containers/CI).

For retroactive ingestion from Claude Code session logs (no uroboro install needed on the remote):

```bash
# Rsync remote sessions, then extract decisions via LLM
rsync -av remote:~/.claude/projects/ ~/.claude/remote-projects/
python scripts/ingest_sessions.py --projects-dir ~/.claude/remote-projects scan
python scripts/ingest_sessions.py --projects-dir ~/.claude/remote-projects extract --out candidates.jsonl
uroboro import candidates.jsonl --machine remote
```

## Backup

```bash
# Create a backup
uroboro backup

# List existing backups
uroboro backup --list

# Custom destination, keep last 5
uroboro backup --dest /mnt/backup/uroboro --keep 5
```

Backups use SQLite's `VACUUM INTO` for a clean, consistent copy. Set up a systemd timer for automatic daily backups:

```bash
# ~/.config/systemd/user/uroboro-backup.service
[Unit]
Description=Uroboro database backup

[Service]
Type=oneshot
ExecStart=uroboro backup --keep 10
```

```bash
# ~/.config/systemd/user/uroboro-backup.timer
[Unit]
Description=Daily uroboro backup

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
```

```bash
systemctl --user enable --now uroboro-backup.timer
```

`Persistent=true` ensures missed backups run on next boot.

## Visualization

**Timeline** (`uro timeline`) — Terminal UI for browsing events with filters.

**Web** (`uro web`) — Browser timeline with list, horizontal, and project views. Compact mode, presentation mode, diff mode, narrative pane, keyboard navigation (`j`/`k`, `1`/`2`/`3` views, `p` present, `n` summary, `N` narrative), and standalone HTML export.

**Graph** (`uro graph`) — Canvas scatter plot showing all activity at a glance. Projects on Y-axis, time on X-axis. Scales to thousands of days.

## Claude Code Integration

Uroboro includes an MCP server for automatic context capture with Claude Code.

### Setup

```bash
claude mcp add uroboro --scope user -- uroboro mcp
uroboro hooks install
uroboro init
```

Add the MCP server, install enforcement hooks, and create per-project capture conventions (`.claude/uroboro.tags`). Then in `~/.claude/CLAUDE.md`, tell Claude to use the tools automatically.

### How It Works

Once configured, Claude automatically captures:
- **Decisions** when choosing between alternatives
- **Blockers** when work is stuck on dependencies
- **Questions** when deferring open issues

Enforcement hooks audit capture compliance at session end and nudge during long sessions. Run `uro recap` to see what happened.

### Available MCP Tools

| Tool | Purpose |
|------|---------|
| `uro_decision` | Record technical decisions |
| `uro_blocker` | Record blockers |
| `uro_question` | Record open questions |
| `uro_capture` | General capture |
| `uro_recap` | Get recent context |
| `uro_search` | Search past captures |
| `uro_stats` | Aggregate statistics (tags, activity, projects) |
| `uro_distill` | Extract style signals to JSONL file |
| `uro_prompt_profile` | Analyze prompting patterns |
| `uro_enforcement` | Configure capture compliance hooks |

## Style Distillation

Extract personal coding style from your actual work history — git commits and uroboro captures — to generate evidence-backed style guides.

```bash
# Extract style signals from one repo
uro distill --days 180 --correlate --out style.jsonl

# Multi-repo extraction
./scripts/distill-multi.sh --days 180 --correlate \
  --out ~/.local/share/uroboro/style-data/ \
  ~/projects/repo1 ~/projects/repo2

# Analyze prompting patterns from Claude Code sessions
uro prompt-profile --days 30
```

The `distill` command extracts style-relevant git commits (refactors, cleanups, structural changes) and uroboro decision captures as JSONL. Use `--correlate` to link commits with captures that occurred within 30 minutes of each other.

Use `scripts/style-analysis-prompt.md` as a Claude Code prompt to analyze the JSONL and produce a personalized style guide, system prompt fragment, and machine-readable rules.

## Philosophy

- **Local-first** — SQLite database in `~/.local/share/uroboro/`
- **No cloud** — Your data stays on your machine(s)
- **Tool-agnostic** — Works standalone or with Claude Code
- **Minimal friction** — Quick captures, powerful retrieval
- **Multi-device ready** — Export/import JSONL to sync across machines; machine field tracks capture origin

## Data

All captures stored locally:
```
~/.local/share/uroboro/uroboro.sqlite
```

Export via `uro timeline --export` (JSON), `uro timeline --export-html` (standalone HTML), or `uro report --format csv`.

## License

MIT
