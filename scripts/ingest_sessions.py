#!/usr/bin/env python3
"""uroboro ingest — Extract captures from Claude Code conversation logs.

Scans ~/.claude/projects/ for JSONL session logs without uroboro captures,
extracts decisions/blockers/questions via Claude API, and outputs candidate
captures for review or direct import.

Usage:
  # Scan and show what's available
  python scripts/ingest_sessions.py scan

  # Extract captures from uncaptured sessions (dry run)
  python scripts/ingest_sessions.py extract --dry-run

  # Extract and write to candidates file
  python scripts/ingest_sessions.py extract --out candidates.jsonl

  # Extract only for a specific project
  python scripts/ingest_sessions.py extract --project qallo --out candidates.jsonl

  # Import candidates into uroboro
  python scripts/ingest_sessions.py import candidates.jsonl

  # Import with review (interactive)
  python scripts/ingest_sessions.py import candidates.jsonl --interactive

Requires: pip install anthropic
"""

import argparse
import json
import os
import subprocess
import sys
from datetime import datetime
from pathlib import Path


CLAUDE_PROJECTS_DIR = Path.home() / ".claude" / "projects"
PROJECTS_DIR_OVERRIDE: Path | None = None  # Set via --projects-dir
MIN_SESSION_SIZE = 10_000  # 10KB — skip trivially small sessions
MAX_CONVERSATION_CHARS = 80_000  # ~20K tokens, fits in context with prompt
EXTRACTION_MODEL = "claude-haiku-4-5-20251001"

EXTRACTION_PROMPT = """\
You are analyzing a Claude Code conversation log to extract technical decisions, \
blockers, and open questions for a decision trail system called uroboro.

Extract ONLY substantive items:

**Decisions** — Active choices between alternatives.
  Format: "X over Y — reason"
  Examples: "JWT over sessions — stateless, scales horizontally"
            "Orphan branch over fork — clean history, no inherited commits"

**Blockers** — Work that could not proceed due to external dependencies.
  Format: "Blocked on X — waiting on Y"

**Questions** — Open questions that were deferred or need future investigation.
  Format: "How should we handle X?"

**Key context** — Significant architectural, design, or project context worth preserving.
  Only if it doesn't fit the above categories.

Do NOT extract:
- Routine code edits without meaningful choices
- Debugging steps or troubleshooting
- Greetings, small talk, status updates
- Information obvious from the code itself
- Trivial decisions (variable names, formatting, import ordering)
- Tool usage mechanics (file reads, searches)
- Already-captured uroboro entries (if you see uro_decision calls, those are already recorded)

For each extraction, return a JSON object with:
- type: "decision" | "blocker" | "question" | "capture"
- content: concise description (for decisions, use "X over Y — reason" format)
- tags: relevant comma-separated tags (e.g., "architecture,auth" or "deployment,docker")
- timestamp: ISO timestamp from conversation context (use the nearest message timestamp)
- confidence: "high" | "medium" | "low"

Return a JSON array. If nothing is worth capturing, return [].
Be conservative — better to miss captures than create noise.\
"""


def discover_sessions(
    project_filter: str | None = None,
    since: str | None = None,
    min_size: int = MIN_SESSION_SIZE,
) -> list[dict]:
    """Find all JSONL session logs, with metadata."""
    sessions = []

    projects_dir = PROJECTS_DIR_OVERRIDE or CLAUDE_PROJECTS_DIR
    if not projects_dir.exists():
        print(f"Claude projects dir not found: {projects_dir}", file=sys.stderr)
        return sessions

    for project_dir in sorted(projects_dir.iterdir()):
        if not project_dir.is_dir():
            continue

        project_name = derive_project_name(project_dir.name)
        if project_filter and project_filter.lower() not in project_name.lower():
            continue

        for jsonl_file in sorted(project_dir.glob("*.jsonl")):
            size = jsonl_file.stat().st_size
            if size < min_size:
                continue

            mtime = datetime.fromtimestamp(jsonl_file.stat().st_mtime)
            if since:
                since_dt = datetime.fromisoformat(since)
                if mtime < since_dt:
                    continue

            sessions.append({
                "path": jsonl_file,
                "project": project_name,
                "project_dir": project_dir.name,
                "size": size,
                "mtime": mtime,
                "session_id": jsonl_file.stem,
            })

    return sessions


def derive_project_name(dir_name: str) -> str:
    """Derive a human-readable project name from Claude's encoded directory path.

    e.g. '-home-qry-work-bttmline-qallo' → 'qallo'
         '-home-qry-projects-claude-in-factorio' → 'claude-in-factorio'
    """
    # Remove the home prefix and split
    parts = dir_name.strip("-").split("-")

    # Find the meaningful suffix — skip home/user/work/projects/bttmline prefixes
    skip_prefixes = {"home", "qry", "work", "projects", "bttmline"}
    meaningful = []
    found_meaningful = False
    for part in parts:
        if found_meaningful or part.lower() not in skip_prefixes:
            found_meaningful = True
            meaningful.append(part)

    if not meaningful:
        return dir_name

    return "-".join(meaningful)


def has_uroboro_captures(path: Path) -> bool:
    """Check if a session already has uroboro MCP calls."""
    try:
        with open(path, "r", encoding="utf-8", errors="replace") as f:
            # Read in chunks to avoid loading huge files
            while chunk := f.read(1_000_000):
                if "uro_" in chunk:
                    return True
        return False
    except Exception:
        return False


def parse_session(path: Path) -> dict:
    """Parse a JSONL session into structured conversation data."""
    messages = []
    session_start = None
    session_end = None
    cwd = None
    git_branch = None

    try:
        with open(path, "r", encoding="utf-8", errors="replace") as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    entry = json.loads(line)
                except json.JSONDecodeError:
                    continue

                entry_type = entry.get("type")
                timestamp = entry.get("timestamp")

                # Track session metadata
                if entry_type == "user" and not cwd:
                    cwd = entry.get("cwd")
                    git_branch = entry.get("gitBranch")
                if timestamp:
                    ts = parse_iso_timestamp(timestamp)
                    if ts:
                        if session_start is None or ts < session_start:
                            session_start = ts
                        if session_end is None or ts > session_end:
                            session_end = ts

                # Extract conversation messages
                if entry_type not in ("user", "assistant"):
                    continue

                message = entry.get("message", {})
                role = message.get("role")
                content = message.get("content", [])

                if role not in ("user", "assistant"):
                    continue

                text_parts = []
                if isinstance(content, str):
                    text_parts.append(content)
                elif isinstance(content, list):
                    for block in content:
                        if isinstance(block, dict):
                            if block.get("type") == "text" and block.get("text"):
                                text_parts.append(block["text"])
                            elif block.get("type") == "tool_use":
                                # Include tool name for context
                                name = block.get("name", "")
                                if name and not name.startswith("mcp__uroboro"):
                                    text_parts.append(f"[tool: {name}]")
                            elif block.get("type") == "tool_result":
                                # Include short results for context
                                result_content = block.get("content", "")
                                if isinstance(result_content, str) and len(result_content) < 500:
                                    text_parts.append(f"[result: {result_content[:200]}]")

                if text_parts:
                    text = "\n".join(text_parts)
                    messages.append({
                        "role": role,
                        "text": text,
                        "timestamp": timestamp,
                    })

    except Exception as e:
        print(f"  Error parsing {path.name}: {e}", file=sys.stderr)

    return {
        "messages": messages,
        "session_start": session_start,
        "session_end": session_end,
        "cwd": cwd,
        "git_branch": git_branch,
    }


def parse_iso_timestamp(s: str) -> datetime | None:
    """Parse ISO timestamp string."""
    if not s:
        return None
    try:
        # Handle Z suffix and various formats
        s = s.replace("Z", "+00:00")
        return datetime.fromisoformat(s)
    except (ValueError, TypeError):
        return None


def format_conversation_for_extraction(parsed: dict, project: str) -> str:
    """Format parsed conversation into a compact string for LLM extraction."""
    lines = []
    lines.append(f"Project: {project}")
    if parsed.get("cwd"):
        lines.append(f"Working directory: {parsed['cwd']}")
    if parsed.get("git_branch"):
        lines.append(f"Git branch: {parsed['git_branch']}")
    if parsed.get("session_start"):
        lines.append(f"Session: {parsed['session_start'].isoformat()}")
    lines.append("---")

    for msg in parsed["messages"]:
        role = "USER" if msg["role"] == "user" else "ASSISTANT"
        ts = msg.get("timestamp", "")
        if ts:
            ts = f" [{ts[:19]}]"
        text = msg["text"]
        # Truncate very long messages (tool outputs, code dumps)
        if len(text) > 2000:
            text = text[:1000] + "\n[...truncated...]\n" + text[-500:]
        lines.append(f"{role}{ts}: {text}")
        lines.append("")

    result = "\n".join(lines)

    # Truncate total if needed
    if len(result) > MAX_CONVERSATION_CHARS:
        # Keep first 60% and last 30%, with a gap marker
        first = int(MAX_CONVERSATION_CHARS * 0.6)
        last = int(MAX_CONVERSATION_CHARS * 0.3)
        result = (
            result[:first]
            + "\n\n[... middle of conversation truncated ...]\n\n"
            + result[-last:]
        )

    return result


def extract_captures_via_sdk(
    conversation_text: str,
    model: str = EXTRACTION_MODEL,
) -> list[dict]:
    """Extract captures using the Anthropic Python SDK (requires ANTHROPIC_API_KEY)."""
    import anthropic

    client = anthropic.Anthropic()  # Uses ANTHROPIC_API_KEY env var
    response = client.messages.create(
        model=model,
        max_tokens=4096,
        system=EXTRACTION_PROMPT,
        messages=[
            {
                "role": "user",
                "content": f"Extract captures from this conversation:\n\n{conversation_text}",
            }
        ],
    )
    return response.content[0].text.strip()


def extract_captures_via_cli(
    conversation_text: str,
    model: str = "haiku",
) -> list[dict]:
    """Extract captures using the `claude` CLI (uses Claude Code's auth)."""
    prompt = f"{EXTRACTION_PROMPT}\n\nExtract captures from this conversation:\n\n{conversation_text}"

    result = subprocess.run(
        [
            "claude", "-p",
            "--model", model,
            "--output-format", "text",
            "--no-session-persistence",
            "--allowedTools", "",
        ],
        input=prompt,
        capture_output=True,
        text=True,
        timeout=120,
    )

    if result.returncode != 0:
        raise RuntimeError(f"claude CLI error: {result.stderr[:500]}")

    return result.stdout.strip()


def extract_captures_from_session(
    conversation_text: str,
    model: str = EXTRACTION_MODEL,
    backend: str = "auto",
) -> list[dict]:
    """Extract captures from conversation text using available backend.

    Backends:
      - "sdk": Use Anthropic Python SDK (needs ANTHROPIC_API_KEY)
      - "cli": Use `claude -p` CLI (uses Claude Code's auth)
      - "auto": Try SDK first, fall back to CLI
    """
    text = None

    if backend in ("sdk", "auto"):
        try:
            import anthropic  # noqa: F401
            if os.environ.get("ANTHROPIC_API_KEY"):
                text = extract_captures_via_sdk(conversation_text, model)
        except Exception as e:
            if backend == "sdk":
                print(f"  SDK error: {e}", file=sys.stderr)
                return []
            # Fall through to CLI

    if text is None and backend in ("cli", "auto"):
        try:
            # Map full model names to CLI aliases for convenience
            cli_model = model
            if "haiku" in model:
                cli_model = "haiku"
            elif "sonnet" in model:
                cli_model = "sonnet"
            elif "opus" in model:
                cli_model = "opus"
            text = extract_captures_via_cli(conversation_text, cli_model)
        except Exception as e:
            print(f"  CLI error: {e}", file=sys.stderr)
            return []

    if text is None:
        print("  No extraction backend available. Set ANTHROPIC_API_KEY or install claude CLI.", file=sys.stderr)
        return []

    try:
        # Handle markdown code blocks
        if text.startswith("```"):
            text = text.split("\n", 1)[1]  # Remove opening ```json
            text = text.rsplit("```", 1)[0]  # Remove closing ```
            text = text.strip()

        captures = json.loads(text)
        if not isinstance(captures, list):
            captures = [captures]

        return captures

    except json.JSONDecodeError as e:
        print(f"  Failed to parse LLM response as JSON: {e}", file=sys.stderr)
        print(f"  Raw response: {text[:500]}", file=sys.stderr)
        return []


def cmd_scan(args):
    """Scan and report on uncaptured sessions."""
    sessions = discover_sessions(
        project_filter=args.project,
        since=args.since,
        min_size=args.min_size,
    )

    # Check which have uroboro captures
    uncaptured = []
    captured_count = 0

    for s in sessions:
        if has_uroboro_captures(s["path"]):
            captured_count += 1
        else:
            uncaptured.append(s)

    # Group by project
    by_project: dict[str, list[dict]] = {}
    for s in uncaptured:
        by_project.setdefault(s["project"], []).append(s)

    total = len(sessions)
    print(f"Total sessions found: {total}")
    print(f"Already captured:     {captured_count}")
    print(f"Uncaptured:           {len(uncaptured)}")
    print(f"Projects with gaps:   {len(by_project)}")
    print()

    if not uncaptured:
        print("All sessions have uroboro captures!")
        return

    # Sort projects by number of uncaptured sessions
    sorted_projects = sorted(by_project.items(), key=lambda x: -len(x[1]))

    print(f"{'Project':<40} {'Sessions':>8} {'Total Size':>12}")
    print("-" * 64)
    for project, project_sessions in sorted_projects:
        total_size = sum(s["size"] for s in project_sessions)
        size_str = format_size(total_size)
        print(f"{project:<40} {len(project_sessions):>8} {size_str:>12}")

    if args.verbose:
        print()
        print("Detailed session list:")
        print()
        for project, project_sessions in sorted_projects:
            print(f"  {project}:")
            for s in sorted(project_sessions, key=lambda x: x["mtime"]):
                size_str = format_size(s["size"])
                print(f"    {s['mtime'].strftime('%Y-%m-%d')}  {size_str:>10}  {s['session_id'][:12]}...")
            print()


def cmd_extract(args):
    """Extract captures from uncaptured sessions."""
    sessions = discover_sessions(
        project_filter=args.project,
        since=args.since,
        min_size=args.min_size,
    )

    uncaptured = [s for s in sessions if not has_uroboro_captures(s["path"])]

    if not uncaptured:
        print("No uncaptured sessions found.", file=sys.stderr)
        return

    # Process largest sessions first (most likely to have substance)
    uncaptured.sort(key=lambda x: -x["size"])

    if args.offset:
        uncaptured = uncaptured[args.offset:]

    if args.limit:
        uncaptured = uncaptured[: args.limit]

    print(f"Processing {len(uncaptured)} sessions...", file=sys.stderr)

    all_candidates = []

    for i, session in enumerate(uncaptured):
        project = session["project"]
        size_str = format_size(session["size"])
        print(
            f"  [{i + 1}/{len(uncaptured)}] {project} — {session['mtime'].strftime('%Y-%m-%d')} "
            f"({size_str}) ...",
            file=sys.stderr,
            end="",
            flush=True,
        )

        # Parse conversation
        parsed = parse_session(session["path"])

        if len(parsed["messages"]) < 3:
            print(" skipped (too few messages)", file=sys.stderr)
            continue

        # Format for extraction
        conversation_text = format_conversation_for_extraction(parsed, project)

        if args.dry_run:
            print(f" {len(parsed['messages'])} messages, {len(conversation_text)} chars", file=sys.stderr)
            continue

        # Extract via LLM
        captures = extract_captures_from_session(
            conversation_text, model=args.model, backend=args.backend,
        )
        print(f" → {len(captures)} captures", file=sys.stderr)

        # Enrich with session metadata
        for cap in captures:
            cap["project"] = project
            cap["source_session"] = session["session_id"]
            cap["source_dir"] = session["project_dir"]
            cap["session_date"] = session["mtime"].isoformat()
            # Ensure timestamp exists
            if not cap.get("timestamp"):
                cap["timestamp"] = session["mtime"].strftime("%Y-%m-%dT%H:%M:%S")

        all_candidates.extend(captures)

    if args.dry_run:
        print(f"\nDry run complete. {len(uncaptured)} sessions would be processed.", file=sys.stderr)
        return

    # Output candidates
    if not all_candidates:
        print("\nNo captures extracted.", file=sys.stderr)
        return

    output = sys.stdout
    if args.out:
        output = open(args.out, "w")

    for candidate in all_candidates:
        output.write(json.dumps(candidate) + "\n")

    if args.out:
        output.close()
        print(f"\n{len(all_candidates)} candidates written to {args.out}", file=sys.stderr)
    else:
        print(f"\n{len(all_candidates)} candidates extracted.", file=sys.stderr)


def cmd_import(args):
    """Import candidates from JSONL file into uroboro."""
    candidates_path = Path(args.file)
    if not candidates_path.exists():
        print(f"File not found: {candidates_path}", file=sys.stderr)
        sys.exit(1)

    candidates = []
    with open(candidates_path) as f:
        for line in f:
            line = line.strip()
            if line:
                candidates.append(json.loads(line))

    if not candidates:
        print("No candidates found in file.", file=sys.stderr)
        return

    print(f"Loaded {len(candidates)} candidates")

    # Filter by confidence
    if args.min_confidence:
        confidence_order = {"high": 3, "medium": 2, "low": 1}
        min_level = confidence_order.get(args.min_confidence, 0)
        candidates = [
            c for c in candidates
            if confidence_order.get(c.get("confidence", "low"), 0) >= min_level
        ]
        print(f"After confidence filter ({args.min_confidence}+): {len(candidates)}")

    imported = 0
    skipped = 0

    for i, cap in enumerate(candidates):
        cap_type = cap.get("type", "capture")
        content = cap.get("content", "")
        tags = cap.get("tags", "")
        project = cap.get("project", "")
        timestamp = cap.get("timestamp", "")

        if not content:
            skipped += 1
            continue

        # Ensure type tag is included
        if cap_type in ("decision", "blocker", "question") and cap_type not in tags:
            tags = f"{cap_type},{tags}" if tags else cap_type

        # Display for review
        confidence = cap.get("confidence", "?")
        print(f"\n[{i + 1}/{len(candidates)}] ({confidence}) {cap_type}: {content}")
        if tags:
            print(f"  tags: {tags}")
        print(f"  project: {project}  time: {timestamp}")

        if args.interactive:
            response = input("  Import? [y/N/q] ").strip().lower()
            if response == "q":
                break
            if response != "y":
                skipped += 1
                continue

        if args.dry_run:
            print("  → would import")
            continue

        # Call uroboro CLI
        cmd = ["uroboro", "capture", content, "--tags", tags]
        if project:
            cmd.extend(["--project", project])
        if timestamp:
            # Strip trailing Z (UTC marker) — uroboro CLI expects bare ISO timestamps
            cmd.extend(["--time", timestamp.rstrip("Z")])

        try:
            subprocess.run(cmd, check=True, capture_output=True, text=True, errors="replace")
            imported += 1
        except subprocess.CalledProcessError as e:
            print(f"  Error: {e.stderr}", file=sys.stderr)
            skipped += 1
        except FileNotFoundError:
            print("Error: 'uroboro' binary not found in PATH", file=sys.stderr)
            sys.exit(1)

    print(f"\nDone: {imported} imported, {skipped} skipped")


def format_size(bytes: int) -> str:
    if bytes >= 1 << 20:
        return f"{bytes / (1 << 20):.1f} MB"
    if bytes >= 1 << 10:
        return f"{bytes / (1 << 10):.1f} KB"
    return f"{bytes} B"


def main():
    parser = argparse.ArgumentParser(
        description="uroboro ingest — backfill captures from Claude Code conversation logs"
    )
    parser.add_argument(
        "--projects-dir",
        metavar="DIR",
        help="Override Claude projects directory (default: ~/.claude/projects/)",
    )
    subparsers = parser.add_subparsers(dest="command", required=True)

    # scan
    scan_parser = subparsers.add_parser("scan", help="Scan and report uncaptured sessions")
    scan_parser.add_argument("--project", help="Filter by project name (substring match)")
    scan_parser.add_argument("--since", help="Only sessions after date (YYYY-MM-DD)")
    scan_parser.add_argument("--min-size", type=int, default=MIN_SESSION_SIZE, help="Min session size in bytes")
    scan_parser.add_argument("-v", "--verbose", action="store_true", help="Show individual sessions")

    # extract
    extract_parser = subparsers.add_parser("extract", help="Extract captures from uncaptured sessions")
    extract_parser.add_argument("--project", help="Filter by project name (substring match)")
    extract_parser.add_argument("--since", help="Only sessions after date (YYYY-MM-DD)")
    extract_parser.add_argument("--min-size", type=int, default=MIN_SESSION_SIZE, help="Min session size in bytes")
    extract_parser.add_argument("--limit", type=int, help="Max sessions to process")
    extract_parser.add_argument("--offset", type=int, default=0, help="Skip first N sessions (sorted by size desc)")
    extract_parser.add_argument("--model", default=EXTRACTION_MODEL, help="Claude model for extraction")
    extract_parser.add_argument("--backend", choices=["auto", "sdk", "cli"], default="auto",
                                help="Extraction backend: sdk (ANTHROPIC_API_KEY), cli (claude -p), auto (try both)")
    extract_parser.add_argument("--out", help="Output JSONL file (default: stdout)")
    extract_parser.add_argument("--dry-run", action="store_true", help="Parse sessions but don't call API")

    # import
    import_parser = subparsers.add_parser("import", help="Import candidates into uroboro")
    import_parser.add_argument("file", help="Candidates JSONL file")
    import_parser.add_argument("--interactive", action="store_true", help="Review each capture before importing")
    import_parser.add_argument("--min-confidence", choices=["low", "medium", "high"], help="Minimum confidence level")
    import_parser.add_argument("--dry-run", action="store_true", help="Show what would be imported")

    args = parser.parse_args()

    global PROJECTS_DIR_OVERRIDE
    if args.projects_dir:
        PROJECTS_DIR_OVERRIDE = Path(args.projects_dir).expanduser()

    if args.command == "scan":
        cmd_scan(args)
    elif args.command == "extract":
        cmd_extract(args)
    elif args.command == "import":
        cmd_import(args)


if __name__ == "__main__":
    main()
