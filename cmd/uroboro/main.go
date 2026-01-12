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
	case "timeline", "-t":
		handleTimeline(os.Args[2:])
	case "status", "-s":
		handleStatus(os.Args[2:])
	case "report", "-r":
		handleReport(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func handleCapture(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: uroboro capture \"content\" [--project NAME] [--tags LIST]")
		os.Exit(1)
	}

	fs := flag.NewFlagSet("capture", flag.ExitOnError)
	project := fs.String("project", "", "Project name")
	tags := fs.String("tags", "", "Comma-separated tags")
	fs.Parse(args[1:])

	content := args[0]

	svc, err := capture.NewService(getDBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database error: %v\n", err)
		os.Exit(1)
	}
	defer svc.Close()

	if err := svc.Capture(content, *project, *tags); err != nil {
		fmt.Fprintf(os.Stderr, "Capture failed: %v\n", err)
		os.Exit(1)
	}
}

func handleTimeline(args []string) {
	fs := flag.NewFlagSet("timeline", flag.ExitOnError)
	days := fs.Int("days", 7, "Number of days to show")
	project := fs.String("project", "", "Filter by project")
	export := fs.Bool("export", false, "Export to JSON file")
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

func printUsage() {
	fmt.Println(`uroboro - Track what you did and when

Commands:
  capture "content"    Capture a work insight
  timeline             View interactive timeline (TUI)
  status               Show recent activity
  report               Generate time report for billing

Aliases:
  uro -c "content"     capture
  uro -t               timeline
  uro -s               status
  uro -r               report

Report options:
  --days N             Days to include (default: 7)
  --project NAME       Filter by project
  --format FORMAT      plain, markdown, csv (default: plain)
  --output FILE        Write to file instead of stdout
  --gap MINUTES        Session gap threshold (default: 30)

Examples:
  uro capture "Fixed auth bug in login flow"
  uro timeline --days 14
  uro report --days 7 --format markdown
  uro report --project myapp --format csv --output timesheet.csv`)
}

func getDBPath() string {
	dbPath := common.GetDefaultDBPath()
	os.MkdirAll(filepath.Dir(dbPath), 0755)
	return dbPath
}
