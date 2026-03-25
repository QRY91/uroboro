package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/QRY91/uroboro/internal/conventions"
)

func handleConventions(args []string) {
	fs := flag.NewFlagSet("conventions", flag.ExitOnError)
	days := fs.Int("days", 180, "Limit to last N days (default: 180)")
	since := fs.String("since", "", "Limit to after date (2006-01-02)")
	correlate := fs.Bool("correlate", true, "Join git↔uro captures by ±30min window")
	scanDir := fs.String("scan", "", "Discover git repos in this directory (combines with positional args)")
	auditDir := fs.String("audit-dir", "", "Workspace audit directory for supplementary context")
	outDir := fs.String("out", "", "Output directory (default: ~/.local/share/uroboro/style-data/)")
	allDecisions := fs.Bool("all-decisions", false, "Include decisions from all projects (default: auto-scope to repo names)")
	allCommits := fs.Bool("all-commits", false, "Extract all commits, not just style-signal ones (recommended for convention analysis)")
	maxPerRepo := fs.Int("max-per-repo", 50, "Cap commits per repo (0 = unlimited)")
	fs.Parse(args)

	repos := fs.Args()
	if len(repos) == 0 && *scanDir == "" {
		fmt.Fprintln(os.Stderr, "Usage: uroboro conventions [OPTIONS] [REPO1 REPO2 ...]")
		fmt.Fprintln(os.Stderr, "\nExtract coding conventions from git history across multiple repos.")
		fmt.Fprintln(os.Stderr, "Returns an analysis prompt for Claude to produce STYLE_GUIDE.md, STYLE_PROMPT.md, and style_rules.json.")
		fmt.Fprintln(os.Stderr, "\nTip: use --all-commits for richer signal (captures feature + fix commits, not just refactors).")
		fmt.Fprintln(os.Stderr, "\nOptions:")
		fs.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nExamples:")
		fmt.Fprintln(os.Stderr, "  uro conventions --all-commits ~/projects/myapp")
		fmt.Fprintln(os.Stderr, "  uro conventions --all-commits --scan ~/projects/ --max-per-repo 50")
		fmt.Fprintln(os.Stderr, "  uro conventions --all-decisions ~/projects/uroboro ~/projects/qallo")
		os.Exit(1)
	}

	// Resolve explicit repo paths
	for i, r := range repos {
		abs, err := filepath.Abs(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "conventions: resolve %s: %v\n", r, err)
			os.Exit(1)
		}
		repos[i] = abs
	}

	// Parse --since
	var sinceTime *time.Time
	if *since != "" {
		t, err := parseTimestamp(*since)
		if err != nil {
			fmt.Fprintf(os.Stderr, "conventions: invalid --since: %v\n", err)
			os.Exit(1)
		}
		sinceTime = &t
	}
	if *days > 0 && sinceTime == nil {
		t := time.Now().AddDate(0, 0, -*days)
		sinceTime = &t
	}

	opts := conventions.Options{
		Repos:        repos,
		ScanDir:      *scanDir,
		Days:         *days,
		Since:        sinceTime,
		Correlate:    *correlate,
		AuditDir:     *auditDir,
		OutDir:       *outDir,
		AllDecisions: *allDecisions,
		AllCommits:   *allCommits,
		MaxPerRepo:   *maxPerRepo,
	}

	result, err := conventions.Run(opts, getDBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "conventions: %v\n", err)
		os.Exit(1)
	}

	// Summary to stderr
	projectScope := "all projects"
	if !*allDecisions {
		projectScope = "auto-scoped to repo names"
	}
	fmt.Fprintf(os.Stderr, "conventions: %d git commits from %d repos, %d decisions (%s)\n",
		result.Manifest.GitCount, len(result.Manifest.Repos), result.Manifest.UroCount, projectScope)
	if result.Manifest.CorrelatedCount > 0 {
		fmt.Fprintf(os.Stderr, "conventions: %d correlated pairs\n", result.Manifest.CorrelatedCount)
	}
	fmt.Fprintf(os.Stderr, "conventions: git JSONL → %s\n", result.JSONLPath)

	fmt.Print(result.Prompt)
}
