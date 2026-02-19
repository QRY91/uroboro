package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func handleInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	tags := fs.String("tags", "", "Comma-separated project tags to pre-populate")
	format := fs.String("format", "concise", "Capture format hint (e.g., concise, detailed, quotable)")
	force := fs.Bool("force", false, "Overwrite existing .claude/uroboro.tags")
	fs.Parse(args)

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "init: cannot determine working directory: %v\n", err)
		os.Exit(1)
	}

	project := filepath.Base(cwd)
	tagsFile := filepath.Join(cwd, ".claude", "uroboro.tags")

	// Check if already exists
	if !*force {
		if _, err := os.Stat(tagsFile); err == nil {
			fmt.Fprintf(os.Stderr, "init: %s already exists (use --force to overwrite)\n", tagsFile)
			os.Exit(1)
		}
	}

	// Ensure .claude directory exists
	if err := os.MkdirAll(filepath.Dir(tagsFile), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "init: failed to create .claude directory: %v\n", err)
		os.Exit(1)
	}

	content := generateTagsFile(project, *tags, *format)
	if err := os.WriteFile(tagsFile, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "init: failed to write %s: %v\n", tagsFile, err)
		os.Exit(1)
	}

	fmt.Printf("Created %s\n", tagsFile)
	fmt.Println("Edit this file to define your project's capture conventions.")
	fmt.Println("The SessionStart hook will inject these into every Claude Code session.")
}

func generateTagsFile(project, tagsCSV, format string) string {
	var b strings.Builder
	b.WriteString("# .claude/uroboro.tags\n")
	b.WriteString("# Project-specific uroboro capture convention\n")
	b.WriteString("#\n")
	b.WriteString("# The SessionStart hook reads this file and injects it into agent context.\n")
	b.WriteString("# Define tags, capture triggers, and format preferences below.\n")
	b.WriteString("# Install hooks first: uro hooks install\n\n")

	b.WriteString(fmt.Sprintf("project: %s\n\n", project))

	b.WriteString("tags:\n")
	if tagsCSV != "" {
		for _, tag := range strings.Split(tagsCSV, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				b.WriteString(fmt.Sprintf("  %-14s — [describe when to use this tag]\n", tag))
			}
		}
	} else {
		b.WriteString("  build-log     — general build/development captures\n")
		b.WriteString("  decision      — architectural or design decisions\n")
		b.WriteString("  discovery     — unexpected findings or insights\n")
	}

	b.WriteString("\ncapture-triggers:\n")
	b.WriteString("  - [describe situations that should trigger a capture]\n")
	b.WriteString("  - [e.g., \"any time a library is chosen over alternatives\"]\n")

	b.WriteString(fmt.Sprintf("\nformat: %s\n", format))

	return b.String()
}
