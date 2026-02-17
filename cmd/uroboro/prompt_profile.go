package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/QRY91/uroboro/internal/promptprofile"
)

func handlePromptProfile(args []string) {
	fs := flag.NewFlagSet("prompt-profile", flag.ExitOnError)
	project := fs.String("project", "", "Filter by project name")
	days := fs.Int("days", 0, "Only include sessions modified within last N days")
	stats := fs.Bool("stats", false, "Print statistics summary")
	extract := fs.Bool("extract", false, "Output JSONL extract of user prompts")
	outFile := fs.String("out", "", "Output file (default: stdout)")
	claudeDir := fs.String("claude-dir", "", "Claude Code projects directory (default: ~/.claude/projects)")
	fs.Parse(args)

	// Default to stats if neither mode specified
	if !*stats && !*extract {
		*stats = true
	}

	// Resolve output writer
	var w io.Writer = os.Stdout
	if *outFile != "" {
		f, err := os.Create(*outFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "prompt-profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}

	// Discover projects
	projects, err := promptprofile.DiscoverProjects(*claudeDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prompt-profile: discover projects: %v\n", err)
		os.Exit(1)
	}

	// Filter by project name if specified
	if *project != "" {
		lower := strings.ToLower(*project)
		var filtered []promptprofile.ProjectInfo
		for _, p := range projects {
			if strings.ToLower(p.ProjectName) == lower {
				filtered = append(filtered, p)
			}
		}
		projects = filtered
	}

	if len(projects) == 0 {
		fmt.Fprintln(os.Stderr, "prompt-profile: no matching projects found")
		os.Exit(1)
	}

	// Extract prompts from all matching projects
	var allPrompts []promptprofile.UserPrompt
	for _, p := range projects {
		prompts, err := promptprofile.ExtractFromProject(p, *days)
		if err != nil {
			fmt.Fprintf(os.Stderr, "prompt-profile: extract %s: %v\n", p.ProjectName, err)
			continue
		}
		allPrompts = append(allPrompts, prompts...)
	}

	fmt.Fprintf(os.Stderr, "prompt-profile: %d prompts from %d projects\n", len(allPrompts), len(projects))

	if *extract {
		enc := json.NewEncoder(w)
		for _, p := range allPrompts {
			enc.Encode(p)
		}
	}

	if *stats {
		s := promptprofile.Analyze(allPrompts)
		fmt.Fprint(w, promptprofile.FormatStats(s))
	}
}
