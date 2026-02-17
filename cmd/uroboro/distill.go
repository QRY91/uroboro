package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/QRY91/uroboro/internal/database"
	"github.com/QRY91/uroboro/internal/distill"
)

func handleDistill(args []string) {
	fs := flag.NewFlagSet("distill", flag.ExitOnError)
	source := fs.String("source", "all", "What to extract: git, uro, all")
	repo := fs.String("repo", ".", "Path to git repository")
	outFile := fs.String("out", "", "Output file (default: stdout)")
	project := fs.String("project", "", "Filter uroboro captures by project")
	days := fs.Int("days", 0, "Limit to last N days")
	since := fs.String("since", "", "Limit to after this date (2006-01-02)")
	correlate := fs.Bool("correlate", false, "Join git commits to uro captures by ±30min window")
	fs.Parse(args)

	// Resolve output writer
	var w io.Writer = os.Stdout
	if *outFile != "" {
		f, err := os.Create(*outFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "distill: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}

	// Parse --since
	var sinceTime *time.Time
	if *since != "" {
		t, err := parseTimestamp(*since)
		if err != nil {
			fmt.Fprintf(os.Stderr, "distill: invalid --since: %v\n", err)
			os.Exit(1)
		}
		sinceTime = &t
	}

	// Resolve --days to a since time if no explicit --since
	if *days > 0 && sinceTime == nil {
		t := time.Now().AddDate(0, 0, -*days)
		sinceTime = &t
	}

	// Resolve repo path
	repoPath, err := filepath.Abs(*repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "distill: %v\n", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(w)

	switch *source {
	case "git":
		extracts, err := extractGit(repoPath, sinceTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "distill: git extraction: %v\n", err)
			os.Exit(1)
		}
		for _, g := range extracts {
			enc.Encode(g)
		}
		fmt.Fprintf(os.Stderr, "distill: %d git extracts\n", len(extracts))

	case "uro":
		extracts, err := extractUro(*project, *days, sinceTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "distill: uro extraction: %v\n", err)
			os.Exit(1)
		}
		for _, u := range extracts {
			enc.Encode(u)
		}
		fmt.Fprintf(os.Stderr, "distill: %d uro extracts\n", len(extracts))

	case "all":
		gitExtracts, err := extractGit(repoPath, sinceTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "distill: git extraction: %v\n", err)
			os.Exit(1)
		}
		uroExtracts, err := extractUro(*project, *days, sinceTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "distill: uro extraction: %v\n", err)
			os.Exit(1)
		}
		if *correlate {
			distill.Correlate(gitExtracts, uroExtracts)
		}
		for _, g := range gitExtracts {
			enc.Encode(g)
		}
		for _, u := range uroExtracts {
			enc.Encode(u)
		}
		correlated := 0
		if *correlate {
			for _, u := range uroExtracts {
				if u.CorrelatedGitHash != "" {
					correlated++
				}
			}
		}
		fmt.Fprintf(os.Stderr, "distill: %d git + %d uro extracts", len(gitExtracts), len(uroExtracts))
		if *correlate {
			fmt.Fprintf(os.Stderr, " (%d correlated)", correlated)
		}
		fmt.Fprintln(os.Stderr)

	default:
		fmt.Fprintf(os.Stderr, "distill: unknown --source %q (want: git, uro, all)\n", *source)
		os.Exit(1)
	}
}

func extractGit(repoPath string, since *time.Time) ([]distill.GitExtract, error) {
	ext := distill.NewGitExtractor(repoPath, since)
	return ext.Extract()
}

func extractUro(project string, days int, since *time.Time) ([]distill.UroExtract, error) {
	db, err := database.NewDB(getDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ext := distill.NewUroExtractor(db, project, days, since)
	return ext.Extract()
}
