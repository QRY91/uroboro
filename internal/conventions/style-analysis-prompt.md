# Style Analysis — Instructions for Claude Code

You are analyzing a developer's code changes and decision captures to distill their personal coding style into a structured style guide. You have two data sources.

## Data Sources

**Git extracts** (JSONL file, path provided above) — commits that signal style/design work:
```json
{"source":"git","repo":"projectname","hash":"abc123","parent_hash":"fff000","message":"refactor: flatten nested auth checks","date":"2025-03-14T10:22:00Z","files":["src/auth/middleware.go"],"language":"go","diff":"--- a/...\n+++ b/...\n...","diff_stats":{"additions":12,"deletions":18,"files":1}}
```

**Inline decisions** (provided above the separator) — explicit developer decisions:
```
2025-03-14 [auth-service] JWT over sessions — stateless, scales horizontally
```

These decisions follow an "X over Y — reason" pattern. When a decision was made within 30 minutes of a git commit, they're correlated — that pairing of **intent** (decision) with **implementation** (git diff) is the richest signal.

**`uro_search` tool** — use this to explore decisions by topic:
```
uro_search(query="architecture")   → decisions about architecture choices
uro_search(query="testing")        → decisions about test strategy
uro_search(query="error handling") → decisions about error patterns
uro_search(query="naming")         → decisions about naming conventions
```
The inline decisions are the most recent N. Use `uro_search` when you want to probe a specific pattern or need more context on a topic.

## Your Task

### Step 1: Inventory

Report briefly:
- Number of git extracts, by language and repo
- Number of inline decisions, by project
- Which git commits are correlated with decisions (±30min timestamp match)

### Step 2: Analyze Git Extracts

Read the git JSONL file. For each commit, examine the diff and message together. Look for:

- **Naming patterns**: Variable/function/type naming conventions. camelCase vs snake_case, abbreviation habits, naming length, prefix/suffix patterns.
- **Structure preferences**: Guard clauses vs nested conditionals, function length, file organization, when things get extracted into separate functions/files.
- **Error handling**: Wrapping patterns, sentinel errors, error message formatting, early returns on error.
- **Abstraction threshold**: When does inline code get extracted into a function or module? What triggers decomposition?
- **Simplification patterns**: What does the developer remove in "cleanup" commits? What do they consider noise?
- **Testing patterns**: Test organization, naming, table-driven tests, assertion style.

Group observations by language.

### Step 3: Analyze Decisions

For the inline decisions (and any additional ones from `uro_search`):

- Parse the "X over Y — reason" pattern to extract:
  - The **preferred** approach (X)
  - The **rejected** approach (Y)
  - The **reasoning** (why)
- Group decisions by category: architecture, tooling, code style, dependencies, testing, workflow.
- Note recurring values in the reasoning (performance, simplicity, readability, consistency, DRY, explicit > implicit, etc.)
- Use `uro_search` if you want to probe a specific topic.

### Step 4: Cross-Reference Correlated Pairs

For each git commit that has a temporally-correlated decision:
1. Find the matching decision by timestamp proximity
2. Read the decision for **intent** and the git diff for **implementation**
3. These are your strongest evidence — a stated reason backed by actual code changes

### Step 5: Cluster and Rank

Group related observations into principles. Rank by:
- **Frequency**: How many times does this pattern appear?
- **Breadth**: Does it appear across multiple repos/languages?
- **Intent**: Is it backed by explicit decision captures?
- **Consistency**: Does the developer always do this, or only sometimes?

A principle with 8+ occurrences across 3 repos with decision backing is a hard rule.
A principle from 1 commit in 1 repo with no decision context is a candidate for review.

### Step 6: Produce Output

Generate three files:

---

#### `STYLE_GUIDE.md` — Full Reference

```markdown
# Code Style Guide
> Auto-generated from [N] commits across [repos], correlated with [M] decisions.
> Date range: [earliest] — [latest]

## Philosophy
[3-5 sentences synthesized from the strongest cross-cutting themes.
What does this developer value most? What trade-offs do they consistently make?]

## Structure & Control Flow
### [Rule name — imperative mood]
- **Strength:** [Hard rule / Preference / Candidate] ([N] occurrences, [M] repos)
- **Rationale:** "[quoted decision if available]"
- **Before:**
  ```[lang]
  [actual code from a diff — the "before" state]
  ```
- **After:**
  ```[lang]
  [actual code from the same diff — the "after" state]
  ```

## Error Handling
...

## Naming
...

## [Language-Specific: Go]
...

## [Language-Specific: TypeScript]
...

## Anti-Patterns
### Never: [pattern the developer actively refactors away from]
- **Evidence:** [what they removed/replaced]
```

Keep under ~2000 words total.

---

#### `STYLE_PROMPT.md` — System Prompt Fragment

```markdown
## Code Style Preferences

When writing or reviewing code for this developer, follow these conventions:

### Hard Rules (always follow)
1. [Concise directive]
...

### Preferences (follow unless there's a good reason not to)
1. [Concise directive]
...

### Anti-Patterns (flag if you see these)
1. [What to avoid]
...

### Language-Specific
**Go:** [key conventions]
**TypeScript:** [key conventions]
```

Under 1500 tokens — concise enough to include in a CLAUDE.md or system prompt.

---

#### `style_rules.json` — Machine-Readable

```json
{
  "version": 1,
  "generated": "YYYY-MM-DD",
  "source_stats": {
    "git_extracts": N,
    "decisions": M,
    "repos": ["repo1", "repo2"],
    "date_range": ["2025-01-01", "2026-02-17"]
  },
  "rules": [
    {
      "id": "structure-001",
      "rule": "Prefer guard clauses with early returns over nested conditionals",
      "category": "structure",
      "severity": "hard_rule",
      "languages": ["go", "typescript"],
      "frequency": 12,
      "repos": 3,
      "example_hash": "abc1234",
      "rationale": "Flatten early — depth is the enemy of readability"
    }
  ],
  "anti_patterns": [
    {
      "id": "anti-001",
      "pattern": "Silently swallowing errors",
      "category": "error_handling",
      "evidence_count": 5
    }
  ]
}
```

## Per-Project vs Cross-Project

If asked to analyze per-project: filter decisions by project name. Note which rules are project-specific vs universal.

If asked for cross-project (default): analyze all records together, noting which rules appear in multiple repos.

## Important Notes

- Use **actual code snippets from the diffs** for before/after examples. Don't invent synthetic examples.
- Before/after pairs come from the same diff: lines starting with `-` are "before", lines starting with `+` are "after".
- The developer's own words (from decisions) are more authoritative than inferred patterns. Quote them when available.
- If a pattern only appears once, list it as a "candidate" not a "rule".
- The style guide should sound like it was written by the developer, not about them. Use "I prefer" or imperative mood.
