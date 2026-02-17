# Style Analysis — Instructions for Claude Code

You are analyzing a developer's code changes and decision captures to distill their personal coding style into a structured style guide. You have access to a JSONL file containing two types of records extracted by `uroboro distill`.

## Input Format

Each line in the JSONL file is one of:

**Git extracts** (`source: "git"`) — commits that signal style/design work:
```json
{"source":"git","repo":"projectname","hash":"abc123","parent_hash":"fff000","message":"refactor: flatten nested auth checks","date":"2025-03-14T10:22:00Z","files":["src/auth/middleware.go"],"language":"go","diff":"--- a/...\n+++ b/...\n...","diff_stats":{"additions":12,"deletions":18,"files":1}}
```

**Uroboro extracts** (`source: "uroboro"`) — explicit developer decisions and context:
```json
{"source":"uroboro","type":"decision","content":"JWT over sessions — stateless scaling","project":"auth-service","tags":["architecture","auth"],"timestamp":"2025-03-14T10:15:00Z","correlated_git_hash":"abc123"}
```

When `correlated_git_hash` is present, the uro capture was made within 30 minutes of that git commit — they're about the same change. This pairing of **intent** (uro) with **implementation** (git diff) is the richest signal.

## Your Task

### Step 1: Read and Inventory

Read the JSONL file. Count records by `source`, `language`, and `repo`. Report the inventory as a brief summary before proceeding.

### Step 2: Analyze Git Extracts

For each git extract, examine the diff and commit message together. Look for:

- **Naming patterns**: Variable/function/type naming conventions. camelCase vs snake_case, abbreviation habits, naming length, prefix/suffix patterns.
- **Structure preferences**: Guard clauses vs nested conditionals, function length, file organization, when things get extracted into separate functions/files.
- **Error handling**: Wrapping patterns, sentinel errors, error message formatting, early returns on error.
- **Abstraction threshold**: When does inline code get extracted into a function or module? What triggers decomposition?
- **Simplification patterns**: What does the developer remove in "cleanup" commits? What do they consider noise?
- **Testing patterns**: Test organization, naming, table-driven tests, assertion style.

Group observations by language.

### Step 3: Analyze Uroboro Extracts

For each uro extract (especially `type: "decision"`):

- Parse the "X over Y — reason" pattern to extract:
  - The **preferred** approach (X)
  - The **rejected** approach (Y)
  - The **reasoning** (why)
- Group decisions by category: architecture, tooling, code style, dependencies, testing, workflow.
- Note recurring values in the reasoning (performance, simplicity, readability, consistency, DRY, explicit > implicit, etc.)

### Step 4: Cross-Reference Correlated Pairs

For each uro extract with a `correlated_git_hash`:
1. Find the matching git extract by hash
2. Read the uro content for **intent** and the git diff for **implementation**
3. These are your strongest evidence — a stated reason backed by actual code changes

### Step 5: Cluster and Rank

Group related observations into principles. Rank by:
- **Frequency**: How many times does this pattern appear?
- **Breadth**: Does it appear across multiple repos/languages?
- **Intent**: Is it backed by explicit uro decision captures?
- **Consistency**: Does the developer always do this, or only sometimes?

A principle with 8+ occurrences across 3 repos with uro decision backing is a hard rule.
A principle from 1 commit in 1 repo with no uro context is a candidate for review.

### Step 6: Produce Output

Generate three files:

---

#### `STYLE_GUIDE.md` — Full Reference

```markdown
# Code Style Guide
> Auto-generated from [N] commits across [repos], correlated with [M] uroboro captures.
> Date range: [earliest] — [latest]

## Philosophy
[3-5 sentences synthesized from the strongest cross-cutting themes.
What does this developer value most? What trade-offs do they consistently make?]

## Structure & Control Flow
### [Rule name — imperative mood]
- **Strength:** [Hard rule / Preference / Candidate] ([N] occurrences, [M] repos)
- **Rationale:** "[quoted uroboro capture if available]"
- **Before:**
  ```[lang]
  [actual code from a diff — the "before" state]
  ```
- **After:**
  ```[lang]
  [actual code from the same diff — the "after" state]
  ```

## Error Handling
### [Rule]
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

Keep under ~2000 words total for system prompt usability.

---

#### `STYLE_PROMPT.md` — System Prompt Fragment

```markdown
## Code Style Preferences

When writing or reviewing code for this developer, follow these conventions:

### Hard Rules (always follow)
1. [Concise directive]
2. [Concise directive]
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

This should be under 1500 tokens — concise enough to include in a CLAUDE.md or system prompt.

---

#### `style_rules.json` — Machine-Readable

```json
{
  "version": 1,
  "generated": "YYYY-MM-DD",
  "source_stats": {
    "git_extracts": N,
    "uro_extracts": M,
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

If asked to analyze per-project: filter records by the `repo` field before analysis. Note which rules are project-specific vs universal.

If asked for cross-project (default): analyze all records together. In the output, note which rules appear in multiple repos (universal) vs single repos (project-specific).

## Important Notes

- Use **actual code snippets from the diffs** for before/after examples. Don't invent synthetic examples.
- Before/after pairs come from the same diff: lines starting with `-` are "before", lines starting with `+` are "after".
- The developer's own words (from uro captures) are more authoritative than inferred patterns. Quote them when available.
- If a pattern only appears once, list it as a "candidate" not a "rule".
- The style guide should sound like it was written by the developer, not about them. Use "I prefer" or imperative mood, not "the developer tends to."
