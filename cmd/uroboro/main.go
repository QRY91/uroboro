package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/capture"
	"github.com/QRY91/uroboro/internal/common"
	"github.com/QRY91/uroboro/internal/database"
	"github.com/QRY91/uroboro/internal/journey"
	"github.com/QRY91/uroboro/internal/report"
	"github.com/QRY91/uroboro/internal/status"
	"github.com/QRY91/uroboro/internal/tui"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "capture", "-c":
		handleCapture(os.Args[2:])
	case "d", "decide":
		handleCapture(prependTags(os.Args[2:], "decision"))
	case "b", "block", "blocker":
		handleCapture(prependTags(os.Args[2:], "blocker"))
	case "q", "question":
		handleCapture(prependTags(os.Args[2:], "question"))
	case "search":
		handleSearch(os.Args[2:])
	case "mcp":
		handleMCP()
	case "recap":
		handleRecap(os.Args[2:])
	case "timeline", "-t":
		handleTimeline(os.Args[2:])
	case "web", "-w":
		handleWeb(os.Args[2:])
	case "graph", "-g":
		handleGraph(os.Args[2:])
	case "status", "-s":
		handleStatus(os.Args[2:])
	case "report", "-r":
		handleReport(os.Args[2:])
	case "distill":
		handleDistill(os.Args[2:])
	case "prompt-profile":
		handlePromptProfile(os.Args[2:])
	case "backup":
		handleBackup(os.Args[2:])
	case "hooks":
		handleHooks(os.Args[2:])
	case "init":
		handleInit(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func handleCapture(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: uroboro capture \"content\" [--project NAME] [--tags LIST] [--time TIME]")
		os.Exit(1)
	}

	fs := flag.NewFlagSet("capture", flag.ExitOnError)
	project := fs.String("project", "", "Project name")
	tags := fs.String("tags", "", "Comma-separated tags")
	timeStr := fs.String("time", "", "Timestamp (e.g., '2024-01-15 14:30' or '2024-01-15T14:30:00')")
	fs.Parse(args[1:])

	content := args[0]

	var timestamp *time.Time
	if *timeStr != "" {
		t, err := parseTimestamp(*timeStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid time format: %v\n", err)
			fmt.Fprintln(os.Stderr, "Supported formats: '2024-01-15 14:30', '2024-01-15T14:30:00', '2024-01-15'")
			os.Exit(1)
		}
		timestamp = &t
	}

	svc, err := capture.NewService(getDBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(1)
	}
	defer svc.Close()

	if err := svc.Capture(content, *project, *tags, timestamp); err != nil {
		fmt.Fprintf(os.Stderr, "Capture failed: %v\n", err)
		os.Exit(1)
	}
}

func handleSearch(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: uroboro search \"keyword\" [--project NAME] [--days N] [--since DATE] [--until DATE] [--tags TAGS] [--limit N]")
		os.Exit(1)
	}

	fs := flag.NewFlagSet("search", flag.ExitOnError)
	project := fs.String("project", "", "Filter by project")
	days := fs.Int("days", 0, "Limit to last N days")
	since := fs.String("since", "", "Start date (YYYY-MM-DD)")
	until := fs.String("until", "", "End date (YYYY-MM-DD)")
	tags := fs.String("tags", "", "Comma-separated tag filter")
	limit := fs.Int("limit", 50, "Maximum results")
	fs.Parse(args[1:])

	keyword := args[0]

	db, err := database.NewDB(getDBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	q := database.CaptureQuery{
		Keyword: keyword,
		Project: *project,
		Days:    *days,
		Limit:   *limit,
	}

	if *since != "" {
		t, err := time.ParseInLocation("2006-01-02", *since, time.Local)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid --since date: %v\n", err)
			os.Exit(1)
		}
		q.Since = &t
	}
	if *until != "" {
		t, err := time.ParseInLocation("2006-01-02", *until, time.Local)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid --until date: %v\n", err)
			os.Exit(1)
		}
		q.Until = &t
	}
	if *tags != "" {
		for _, tag := range strings.Split(*tags, ",") {
			if t := strings.TrimSpace(tag); t != "" {
				q.Tags = append(q.Tags, t)
			}
		}
	}

	captures, err := db.QueryCaptures(q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Search error: %v\n", err)
		os.Exit(1)
	}

	if len(captures) == 0 {
		fmt.Println("No captures found.")
		return
	}

	for _, c := range captures {
		proj := c.Project
		if proj == "" {
			proj = "-"
		}
		tagsInfo := ""
		if c.Tags != "" {
			tagsInfo = "  {" + c.Tags + "}"
		}
		fmt.Printf("%s  [%s]%s  %s\n",
			c.Timestamp.Format("2006-01-02 15:04"),
			proj,
			tagsInfo,
			c.Content,
		)
	}
}

func parseTimestamp(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse '%s'", s)
}

func handleTimeline(args []string) {
	fs := flag.NewFlagSet("timeline", flag.ExitOnError)
	days := fs.Int("days", 7, "Number of days to show")
	project := fs.String("project", "", "Filter by project")
	export := fs.Bool("export", false, "Export to JSON file")
	exportHTML := fs.Bool("export-html", false, "Export to standalone HTML file")
	fs.Parse(args)

	db, err := database.NewDB(getDBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	var projects []string
	if *project != "" {
		projects = strings.Split(*project, ",")
	}

	opts := journey.Options{Days: *days, Projects: projects}

	svc := journey.NewService(db)
	data, err := svc.GenerateJourney(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generate error: %v\n", err)
		os.Exit(1)
	}

	if *export {
		filename := fmt.Sprintf("journey-%d-days.json", *days)
		f, err := os.Create(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		enc.Encode(data)
		fmt.Printf("Exported to %s\n", filename)
		return
	}

	if *exportHTML {
		html, err := GenerateStandaloneHTML(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "HTML generation error: %v\n", err)
			os.Exit(1)
		}
		filename := fmt.Sprintf("timeline-%d-days.html", *days)
		if err := os.WriteFile(filename, []byte(html), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Exported to %s\n", filename)
		return
	}

	// Run TUI
	if err := tui.Run(data); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}

func handleStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	days := fs.Int("days", 7, "Number of days to show")
	project := fs.String("project", "", "Filter by project")
	fs.Parse(args)

	svc := status.NewService()
	if err := svc.Show(*days, getDBPath(), *project); err != nil {
		fmt.Fprintf(os.Stderr, "Status error: %v\n", err)
		os.Exit(1)
	}
}

func handleReport(args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	days := fs.Int("days", 7, "Number of days to report")
	project := fs.String("project", "", "Filter by project")
	format := fs.String("format", "plain", "Output format: plain, markdown, csv")
	output := fs.String("output", "", "Output file (default: stdout)")
	sessionGap := fs.Int("gap", 30, "Minutes between events to start new session")
	fs.Parse(args)

	db, err := database.NewDB(getDBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	var projects []string
	if *project != "" {
		projects = strings.Split(*project, ",")
	}

	opts := journey.Options{Days: *days, Projects: projects}
	svc := journey.NewService(db)
	data, err := svc.GenerateJourney(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generate error: %v\n", err)
		os.Exit(1)
	}

	r := report.Generate(data, time.Duration(*sessionGap)*time.Minute)

	var out string
	switch *format {
	case "markdown", "md":
		out = r.FormatMarkdown()
	case "csv":
		out = r.FormatCSV()
	default:
		out = r.FormatPlain()
	}

	if *output != "" {
		if err := os.WriteFile(*output, []byte(out), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Report written to %s\n", *output)
	} else {
		fmt.Print(out)
	}
}

func handleBackup(args []string) {
	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	dest := fs.String("dest", "", "Destination directory (default: ~/.local/share/uroboro/backups)")
	keep := fs.Int("keep", 10, "Number of backups to keep (0 = keep all)")
	list := fs.Bool("list", false, "List existing backups")
	fs.Parse(args)

	backupDir := *dest
	if backupDir == "" {
		backupDir = common.GetBackupDir()
	}

	if *list {
		listBackups(backupDir)
		return
	}

	dbPath := getDBPath()

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Create backup dir: %v\n", err)
		os.Exit(1)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("uroboro-backup-%s.sqlite", timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	// Use VACUUM INTO for a clean, consistent backup
	db, err := database.NewDB(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(1)
	}

	if err := db.VacuumInto(backupPath); err != nil {
		db.Close()
		fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
		os.Exit(1)
	}

	// Verify backup by counting captures
	count, err := db.CaptureCount()
	db.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not verify source count: %v\n", err)
	}

	backupDB, err := database.NewDB(backupPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not verify backup: %v\n", err)
	} else {
		backupCount, err := backupDB.CaptureCount()
		backupDB.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not count backup captures: %v\n", err)
		} else if backupCount != count {
			fmt.Fprintf(os.Stderr, "Warning: backup has %d captures, source has %d\n", backupCount, count)
		}
	}

	info, _ := os.Stat(backupPath)
	size := "unknown"
	if info != nil {
		size = formatSize(info.Size())
	}

	fmt.Printf("Backup created: %s (%s, %d captures)\n", backupPath, size, count)

	// Rotate old backups
	if *keep > 0 {
		rotateBackups(backupDir, *keep)
	}
}

func listBackups(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No backups found.")
			return
		}
		fmt.Fprintf(os.Stderr, "Read backup dir: %v\n", err)
		os.Exit(1)
	}

	var backups []os.DirEntry
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "uroboro-backup-") && strings.HasSuffix(e.Name(), ".sqlite") {
			backups = append(backups, e)
		}
	}

	if len(backups) == 0 {
		fmt.Println("No backups found.")
		return
	}

	fmt.Printf("Backups in %s:\n", dir)
	for _, b := range backups {
		info, _ := b.Info()
		size := "?"
		if info != nil {
			size = formatSize(info.Size())
		}
		fmt.Printf("  %s  %s\n", b.Name(), size)
	}
	fmt.Printf("\n%d backup(s)\n", len(backups))
}

func rotateBackups(dir string, keep int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var backups []os.DirEntry
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "uroboro-backup-") && strings.HasSuffix(e.Name(), ".sqlite") {
			backups = append(backups, e)
		}
	}

	if len(backups) <= keep {
		return
	}

	// DirEntry is sorted by name, and our timestamp format sorts chronologically
	toRemove := backups[:len(backups)-keep]
	for _, b := range toRemove {
		path := filepath.Join(dir, b.Name())
		if err := os.Remove(path); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not remove old backup %s: %v\n", b.Name(), err)
		} else {
			fmt.Printf("Rotated: %s\n", b.Name())
		}
	}
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func printUsage() {
	fmt.Println(`uroboro - Track what you did and when

Commands:
  capture "content"    Capture a work insight
  d "X over Y"         Quick decision capture
  b "waiting on X"     Quick blocker capture
  q "open question"    Quick question capture
  search "keyword"     Search past captures
  recap                Show recent decisions, blockers, commits
  timeline             View interactive timeline (TUI)
  web                  View scrollable timeline in browser
  graph                View auto-scaled overview graph (fits screen)
  status               Show recent activity
  report               Generate time report for billing
  backup               Back up the capture database
  distill              Extract style signal data from git + uroboro
  prompt-profile       Analyze Claude Code prompting patterns
  hooks                Install/uninstall enforcement hooks for Claude Code
  init                 Create .claude/uroboro.tags for current project
  mcp                  Start MCP server (for Claude Code)

Aliases:
  uro -c "content"     capture
  uro -t               timeline
  uro -w               web
  uro -g               graph
  uro -s               status
  uro -r               report

Capture options:
  --project NAME       Project name (auto-detected if in git repo)
  --tags LIST          Comma-separated tags
  --time TIME          Timestamp for retroactive logging (default: now)
                       Formats: '2024-01-15 14:30', '2024-01-15T14:30:00', '2024-01-15'

Search options:
  --project NAME       Filter by project
  --days N             Limit to last N days
  --limit N            Maximum results (default: 20)

Timeline options:
  --days N             Days to show (default: 7)
  --project NAME       Filter by project
  --export             Export to JSON file
  --export-html        Export to standalone HTML file

Web options:
  --days N             Days to show (default: 7)
  --port N             HTTP port (default: 8080)
  --project NAME       Filter by project

Graph options:
  --days N             Days to show (default: 30)
  --port N             HTTP port (default: 8080)
  --project NAME       Filter by project

Recap options:
  --days N             Days to look back (default: 7)
  --project NAME       Filter by project
  --branch NAME        Scope to git branch

Report options:
  --days N             Days to include (default: 7)
  --project NAME       Filter by project
  --format FORMAT      plain, markdown, csv (default: plain)
  --output FILE        Write to file instead of stdout
  --gap MINUTES        Session gap threshold (default: 30)

Prompt profile options:
  --project NAME       Filter by project name
  --days N             Only recent sessions (last N days)
  --stats              Print statistics summary (default)
  --extract            Output JSONL extract of user prompts
  --out FILE           Output file (default: stdout)
  --claude-dir PATH    Claude Code projects dir (default: ~/.claude/projects)

Distill options:
  --source SOURCE      git, uro, or all (default: all)
  --repo PATH          Git repository path (default: current directory)
  --out FILE           Output file (default: stdout)
  --project NAME       Filter uroboro captures by project
  --days N             Limit to last N days
  --since DATE         Limit to after date (2006-01-02)
  --correlate          Join git↔uro captures by ±30min window

Backup options:
  --dest DIR           Destination directory (default: ~/.local/share/uroboro/backups)
  --keep N             Number of backups to keep, 0 = all (default: 10)
  --list               List existing backups

Hooks options:
  install              Install enforcement hooks into ~/.claude/
  uninstall            Remove enforcement hooks
  status               Check installation status

Init options:
  --tags LIST          Comma-separated project tags to pre-populate
  --format FORMAT      Capture format hint (default: concise)
  --force              Overwrite existing .claude/uroboro.tags

Examples:
  uro d "JWT over sessions - stateless scaling"
  uro b "waiting on backend API"
  uro q "token revocation strategy?"
  uro recap --days 14
  uro search "auth" --project myapp
  uro timeline --days 14
  uro web --port 3000
  uro graph --days 60
  uro timeline --export-html --days 30

Setup:
  uroboro hooks install                          Install enforcement hooks
  uroboro hooks status                           Check hook status
  uroboro init                                   Create capture conventions for current project
  uroboro init --tags "frontend,perf,a11y"       Pre-populate with custom tags`)
}

func getDBPath() string {
	dbPath := common.GetDefaultDBPath()
	os.MkdirAll(filepath.Dir(dbPath), 0755)
	return dbPath
}

func prependTags(args []string, tag string) []string {
	if len(args) == 0 {
		return args
	}
	// Keep content first, then add --tags flag
	return append([]string{args[0], "--tags", tag}, args[1:]...)
}
