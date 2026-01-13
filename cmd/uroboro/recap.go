package main

import (
	"flag"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/database"
)

type recapEvent struct {
	Timestamp time.Time
	Type      string // commit, decision, blocker, question, capture
	Content   string
	Hash      string // git commit hash if applicable
}

func handleRecap(args []string) {
	fs := flag.NewFlagSet("recap", flag.ExitOnError)
	days := fs.Int("days", 7, "Days to look back")
	project := fs.String("project", "", "Filter by project")
	branch := fs.String("branch", "", "Scope to git branch")
	brief := fs.Bool("brief", false, "One-line-per-item output")
	fs.Parse(args)

	proj := *project
	if proj == "" {
		proj = detectProject()
	}

	since := time.Now().AddDate(0, 0, -*days)

	// Collect events
	var events []recapEvent
	var decisions, blockers, questions []recapEvent

	// Get captures from database
	db, err := database.NewDB(getDBPath())
	if err != nil {
		fmt.Printf("Database error: %v\n", err)
		return
	}
	defer db.Close()

	captures, err := db.QueryCaptures(database.CaptureQuery{
		Days:    *days,
		Project: proj,
		Limit:   100,
	})
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		return
	}

	for _, c := range captures {
		eventType := "capture"
		tags := strings.Split(c.Tags, ",")
		for _, t := range tags {
			switch strings.TrimSpace(t) {
			case "decision":
				eventType = "decision"
			case "blocker":
				eventType = "blocker"
			case "question":
				eventType = "question"
			}
		}

		event := recapEvent{
			Timestamp: c.Timestamp,
			Type:      eventType,
			Content:   c.Content,
		}
		events = append(events, event)

		switch eventType {
		case "decision":
			decisions = append(decisions, event)
		case "blocker":
			blockers = append(blockers, event)
		case "question":
			questions = append(questions, event)
		}
	}

	// Get git commits
	commits := getGitCommits(since, *branch)
	events = append(events, commits...)

	// Get files changed
	files := getFilesChanged(since, *branch)

	// Sort events by time (newest first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	// Output
	if *brief {
		printBriefRecap(events)
	} else {
		printFullRecap(proj, *branch, since, events, decisions, blockers, questions, files)
	}
}

func getGitCommits(since time.Time, branch string) []recapEvent {
	args := []string{
		"log",
		"--since=" + since.Format("2006-01-02"),
		"--format=%H|%aI|%s",
		"--no-merges",
	}
	if branch != "" {
		args = append(args, branch)
	}

	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil
	}

	var events []recapEvent
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}

		ts, _ := time.Parse(time.RFC3339, parts[1])
		hash := parts[0]
		if len(hash) > 7 {
			hash = hash[:7]
		}

		events = append(events, recapEvent{
			Timestamp: ts,
			Type:      "commit",
			Content:   parts[2],
			Hash:      hash,
		})
	}

	return events
}

type fileChange struct {
	Path   string
	Status string
}

func getFilesChanged(since time.Time, branch string) []fileChange {
	var args []string

	if branch != "" {
		base := getBaseBranch()
		args = []string{"diff", "--name-status", base + "..." + branch}
	} else {
		args = []string{
			"log",
			"--since=" + since.Format("2006-01-02"),
			"--name-status",
			"--format=",
			"--no-merges",
		}
	}

	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil
	}

	seen := make(map[string]string)
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			status := statusWord(parts[0])
			path := parts[1]
			if _, exists := seen[path]; !exists {
				seen[path] = status
			}
		}
	}

	var files []fileChange
	for path, status := range seen {
		files = append(files, fileChange{Path: path, Status: status})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return files
}

func getBaseBranch() string {
	out, err := exec.Command("git", "rev-parse", "--verify", "main").Output()
	if err == nil && len(out) > 0 {
		return "main"
	}
	return "master"
}

func statusWord(s string) string {
	if len(s) == 0 {
		return "changed"
	}
	switch s[0] {
	case 'A':
		return "new"
	case 'M':
		return "modified"
	case 'D':
		return "deleted"
	case 'R':
		return "renamed"
	default:
		return "changed"
	}
}

func printBriefRecap(events []recapEvent) {
	for _, e := range events {
		content := e.Content
		if len(content) > 60 {
			content = content[:57] + "..."
		}
		fmt.Printf("%s  %-8s  %s\n",
			e.Timestamp.Format("Jan 02 15:04"),
			e.Type,
			content)
	}
}

func printFullRecap(project, branch string, since time.Time, events []recapEvent, decisions, blockers, questions []recapEvent, files []fileChange) {
	// Header
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("  RECAP: %s", project)
	if branch != "" {
		fmt.Printf(" (%s)", branch)
	}
	fmt.Println()
	fmt.Printf("  %s - %s\n", since.Format("Jan 2"), time.Now().Format("Jan 2, 2006"))
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println()

	// Timeline
	if len(events) > 0 {
		fmt.Println("## Timeline")
		fmt.Println()

		currentDay := ""
		for _, e := range events {
			day := formatDay(e.Timestamp)
			if day != currentDay {
				fmt.Printf("%s\n", day)
				currentDay = day
			}

			symbol := eventSymbol(e.Type)
			timeStr := e.Timestamp.Format("15:04")
			content := e.Content
			if len(content) > 55 {
				content = content[:52] + "..."
			}

			fmt.Printf("  %s  %s %s\n", timeStr, symbol, content)
		}
		fmt.Println()
	}

	// Decisions
	if len(decisions) > 0 {
		fmt.Printf("## Decisions (%d)\n", len(decisions))
		for i, d := range decisions {
			fmt.Printf("  %d. %s\n", i+1, d.Content)
		}
		fmt.Println()
	}

	// Blockers
	if len(blockers) > 0 {
		fmt.Printf("## Blockers (%d)\n", len(blockers))
		for _, b := range blockers {
			fmt.Printf("  - %s (%s)\n", b.Content, b.Timestamp.Format("Jan 2"))
		}
		fmt.Println()
	}

	// Questions
	if len(questions) > 0 {
		fmt.Println("## Open Questions")
		for _, q := range questions {
			fmt.Printf("  - %s\n", q.Content)
		}
		fmt.Println()
	}

	// Files
	if len(files) > 0 {
		fmt.Printf("## Files Changed (%d)\n", len(files))
		max := 10
		for i, f := range files {
			if i >= max {
				fmt.Printf("  ...%d more\n", len(files)-max)
				break
			}
			fmt.Printf("  %s (%s)\n", f.Path, f.Status)
		}
		fmt.Println()
	}

	if len(events) == 0 {
		fmt.Println("No recent activity found.")
	}
}

func formatDay(t time.Time) string {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	day := t.Format("2006-01-02")

	if day == today {
		return "Today"
	}
	if day == yesterday {
		return "Yesterday"
	}
	return t.Format("Jan 2 (Mon)")
}

func eventSymbol(t string) string {
	switch t {
	case "decision":
		return "[D]"
	case "blocker":
		return "[B]"
	case "question":
		return "[?]"
	case "commit":
		return "[C]"
	default:
		return "[.]"
	}
}
