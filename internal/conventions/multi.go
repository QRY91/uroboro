package conventions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/database"
	"github.com/QRY91/uroboro/internal/distill"
)

// DecisionLimit is the max number of decisions to include inline.
// At ~150 chars each, 300 decisions ≈ 45KB — manageable inline.
const DecisionLimit = 300

type Options struct {
	Repos        []string
	ScanDir      string   // Discover git repos in this directory
	Days         int
	Since        *time.Time
	Correlate    bool
	AuditDir     string
	OutDir       string
	Projects     []string // Explicit project filter for decisions (overrides auto-detect)
	AllDecisions bool     // Include decisions from all projects, not just matching repos
	AllCommits   bool     // Extract all commits, not just style-signal ones
	MaxPerRepo   int      // Cap commits per repo (0 = unlimited; default set in Run)
}

type Result struct {
	Manifest     Manifest
	JSONLPath    string
	ManifestPath string
	Prompt       string
}

func Run(opts Options, dbPath string) (*Result, error) {
	// Default output directory
	if opts.OutDir == "" {
		home, _ := os.UserHomeDir()
		opts.OutDir = filepath.Join(home, ".local", "share", "uroboro", "style-data")
	}
	if err := os.MkdirAll(opts.OutDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	home, _ := os.UserHomeDir()
	conventionsDir := filepath.Join(home, ".local", "share", "uroboro", "conventions")
	if err := os.MkdirAll(conventionsDir, 0755); err != nil {
		return nil, fmt.Errorf("create conventions dir: %w", err)
	}

	// --- Repo discovery ---
	var warnings []string
	repos := opts.Repos

	if opts.ScanDir != "" {
		discovered, err := discoverRepos(opts.ScanDir)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("scan %s: %v", opts.ScanDir, err))
		} else if len(discovered) == 0 {
			warnings = append(warnings, fmt.Sprintf("scan %s: no git repositories found", opts.ScanDir))
		}
		repos = append(repos, discovered...)
	}

	if len(repos) == 0 && opts.ScanDir == "" {
		return nil, fmt.Errorf("no repos specified (use positional args or --scan DIR)")
	}

	// --- Git extraction → JSONL (diffs only, no uro captures) ---
	var allGit []distill.GitExtract
	repoCounts := make(map[string]int)
	langCounts := make(map[string]int)

	for _, repo := range repos {
		absPath, err := filepath.Abs(repo)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("skip %s: %v", repo, err))
			continue
		}
		if _, err := os.Stat(filepath.Join(absPath, ".git")); err != nil {
			warnings = append(warnings, fmt.Sprintf("skip %s: not a git repository", absPath))
			continue
		}

		maxPerRepo := opts.MaxPerRepo
		if maxPerRepo == 0 {
			maxPerRepo = 50 // default cap to prevent large repos from dominating
		}
		ext := distill.NewGitExtractor(absPath, opts.Since)
		ext.AllCommits = opts.AllCommits
		ext.MaxPerRepo = maxPerRepo
		extracts, err := ext.Extract()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("skip %s: git extraction failed: %v", filepath.Base(absPath), err))
			continue
		}

		name := filepath.Base(absPath)
		repoCounts[name] = len(extracts)
		for _, g := range extracts {
			if g.Language != "" {
				langCounts[g.Language]++
			}
		}
		allGit = append(allGit, extracts...)
	}

	// Write git-only JSONL
	jsonlPath := filepath.Join(opts.OutDir, fmt.Sprintf("git-extracts-%s.jsonl", time.Now().Format("2006-01-02-150405")))
	if len(allGit) > 0 {
		f, err := os.Create(jsonlPath)
		if err != nil {
			return nil, fmt.Errorf("create JSONL: %w", err)
		}
		enc := json.NewEncoder(f)
		for _, g := range allGit {
			enc.Encode(g)
		}
		f.Close()
	} else if len(warnings) > 0 {
		return nil, fmt.Errorf("no git style signals found. warnings:\n  %s", strings.Join(warnings, "\n  "))
	}

	// --- Decision query → inline (DB, not JSONL) ---
	// Auto-scope: if not AllDecisions and no explicit Projects, use repo basenames
	projectFilter := opts.Projects
	if !opts.AllDecisions && len(projectFilter) == 0 {
		for name := range repoCounts {
			projectFilter = append(projectFilter, name)
		}
	}

	var decisions []database.Capture
	totalDecisions := 0

	db, err := database.NewDB(dbPath)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("uro db: %v (no decision context)", err))
	} else {
		defer db.Close()

		q := database.CaptureQuery{
			Tags:     []string{"decision"},
			Days:     opts.Days,
			Since:    opts.Since,
			Projects: projectFilter,
			Limit:    DecisionLimit,
		}
		if opts.Since != nil {
			q.Days = 0
		}
		decisions, err = db.QueryCaptures(q)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("query decisions: %v", err))
		}

		// Total count without limit (for "showing N of M" message)
		countQ := database.CaptureQuery{Tags: []string{"decision"}, Days: opts.Days, Since: opts.Since, Projects: projectFilter}
		if opts.Since != nil {
			countQ.Days = 0
		}
		all, _ := db.QueryCaptures(countQ)
		totalDecisions = len(all)
	}

	// Correlate git commits with decisions by timestamp proximity
	correlatedHashes := correlateByTime(allGit, decisions)

	// Handle audit context
	var auditPath string
	if opts.AuditDir != "" {
		ap, err := writeAuditContext(opts.AuditDir, opts.OutDir)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("audit: %v", err))
		} else {
			auditPath = ap
		}
	}

	// Build manifest
	repoNames := make([]string, 0, len(repoCounts))
	for name := range repoCounts {
		repoNames = append(repoNames, name)
	}
	sort.Strings(repoNames)

	manifest := Manifest{
		Date:            time.Now().Format("2006-01-02"),
		Repos:           repoNames,
		GitCount:        len(allGit),
		UroCount:        totalDecisions,
		CorrelatedCount: len(correlatedHashes),
		TotalRecords:    len(allGit) + len(decisions),
		Days:            opts.Days,
		OutputFile:      jsonlPath,
		AuditDir:        opts.AuditDir,
		ConventionsDir:  conventionsDir,
		LanguageCounts:  langCounts,
		RepoCounts:      repoCounts,
	}
	if opts.Since != nil {
		manifest.Since = opts.Since.Format("2006-01-02")
	}

	manifestPath := filepath.Join(opts.OutDir, "conventions-manifest.json")
	if err := WriteManifest(manifestPath, manifest); err != nil {
		return nil, fmt.Errorf("write manifest: %w", err)
	}

	prompt := buildPrompt(manifest, jsonlPath, auditPath, conventionsDir, decisions, totalDecisions, correlatedHashes, warnings, opts.AllCommits, opts.MaxPerRepo)

	return &Result{
		Manifest:     manifest,
		JSONLPath:    jsonlPath,
		ManifestPath: manifestPath,
		Prompt:       prompt,
	}, nil
}

// discoverRepos finds immediate subdirectories of dir that contain a .git folder.
func discoverRepos(dir string) ([]string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}
	var repos []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		candidate := filepath.Join(abs, e.Name())
		if _, err := os.Stat(filepath.Join(candidate, ".git")); err == nil {
			repos = append(repos, candidate)
		}
	}
	return repos, nil
}

// correlateByTime finds git commits that have a decision within ±30 minutes.
// Returns a set of git hashes that have correlated decisions.
func correlateByTime(gitExtracts []distill.GitExtract, decisions []database.Capture) map[string]struct{} {
	correlated := make(map[string]struct{})
	window := 30 * time.Minute
	for _, g := range gitExtracts {
		for _, d := range decisions {
			diff := g.Date.Sub(d.Timestamp)
			if diff < 0 {
				diff = -diff
			}
			if diff <= window {
				correlated[g.Hash] = struct{}{}
				break
			}
		}
	}
	return correlated
}

func writeAuditContext(auditDir, outDir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(auditDir, "*.md"))
	if err != nil {
		return "", fmt.Errorf("glob audit dir: %w", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no .md files in %s", auditDir)
	}

	sort.Strings(matches)
	outPath := filepath.Join(outDir, "audit-context.md")
	out, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	fmt.Fprintf(out, "# Workspace Audit Context\n\n")
	fmt.Fprintf(out, "Supplementary tech stack and architecture context for %d projects.\n\n", len(matches))
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		fmt.Fprintf(out, "---\n\n## %s\n\n%s\n\n", strings.TrimSuffix(filepath.Base(path), ".md"), string(data))
	}
	return outPath, nil
}

func buildPrompt(m Manifest, jsonlPath, auditPath, conventionsDir string, decisions []database.Capture, totalDecisions int, correlatedHashes map[string]struct{}, warnings []string, allCommits bool, maxPerRepo int) string {
	var b strings.Builder

	b.WriteString("## Convention Analysis — Ready to Run\n\n")

	filterDesc := "style-signal filter"
	if allCommits {
		if maxPerRepo > 0 {
			filterDesc = fmt.Sprintf("all commits (capped at %d/repo)", maxPerRepo)
		} else {
			filterDesc = "all commits"
		}
	}
	fmt.Fprintf(&b, "**Git commits:** %d commits from %d repos (%s)\n", m.GitCount, len(m.Repos), filterDesc)
	fmt.Fprintf(&b, "**Decisions:** %d inline", len(decisions))
	if totalDecisions > len(decisions) {
		fmt.Fprintf(&b, " (of %d total — use `uro_search` for more)", totalDecisions)
	}
	b.WriteString("\n")

	if len(correlatedHashes) > 0 {
		fmt.Fprintf(&b, "**Correlated:** %d commits have a decision within ±30min\n", len(correlatedHashes))
	}

	if len(m.LanguageCounts) > 0 {
		langs := make([]string, 0, len(m.LanguageCounts))
		for lang, count := range m.LanguageCounts {
			langs = append(langs, fmt.Sprintf("%s (%d)", lang, count))
		}
		sort.Strings(langs)
		fmt.Fprintf(&b, "**Languages:** %s\n", strings.Join(langs, ", "))
	}

	fmt.Fprintf(&b, "**Git JSONL:** `%s`\n", jsonlPath)
	if auditPath != "" {
		fmt.Fprintf(&b, "**Workspace context:** `%s`\n", auditPath)
	}

	if len(warnings) > 0 {
		b.WriteString("\n**Warnings:**\n")
		for _, w := range warnings {
			fmt.Fprintf(&b, "- %s\n", w)
		}
	}

	// Instructions
	b.WriteString("\n### Instructions\n\n")
	if m.GitCount > 0 {
		fmt.Fprintf(&b, "1. Read the git JSONL at `%s` — contains %d commits with diffs.\n", jsonlPath, m.GitCount)
		b.WriteString("   The file may be large; read it in chunks using `offset` and `limit` parameters if needed.\n")
		b.WriteString("   Each line is one JSON commit record. Read ~20 lines at a time to stay within context.\n")
	}
	if auditPath != "" {
		fmt.Fprintf(&b, "2. Read workspace context at `%s` for tech stack per project\n", auditPath)
	}
	b.WriteString("3. Review the inline decisions below (most recent first)\n")
	b.WriteString("4. Use `uro_search` to explore specific topics deeper (e.g. `uro_search(query=\"architecture\")`, `uro_search(query=\"testing\")`)\n")
	fmt.Fprintf(&b, "5. Write the three output files to `%s`:\n", conventionsDir)
	b.WriteString("   - `STYLE_GUIDE.md` — Full reference with before/after examples from actual diffs\n")
	b.WriteString("   - `STYLE_PROMPT.md` — Compact fragment for CLAUDE.md (under 1500 tokens)\n")
	b.WriteString("   - `style_rules.json` — Machine-readable rules\n")

	// Inline decisions
	if len(decisions) > 0 {
		b.WriteString("\n---\n\n## Inline Decisions\n\n")
		if totalDecisions > len(decisions) {
			fmt.Fprintf(&b, "_Showing %d of %d decisions (most recent). Use `uro_search` to query by topic._\n\n", len(decisions), totalDecisions)
		}
		for _, d := range decisions {
			proj := d.Project
			if proj == "" {
				proj = "-"
			}
			fmt.Fprintf(&b, "%s [%s] %s\n",
				d.Timestamp.Format("2006-01-02"),
				proj,
				d.Content,
			)
		}
	}

	b.WriteString("\n---\n\n")
	b.WriteString(AnalysisPrompt)

	return b.String()
}
