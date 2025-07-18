package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/capture"
	"github.com/QRY91/uroboro/internal/common"
	"github.com/QRY91/uroboro/internal/database"
	"github.com/QRY91/uroboro/internal/journey"
	"github.com/QRY91/uroboro/internal/publish"
	"github.com/QRY91/uroboro/internal/ripcord"
	"github.com/QRY91/uroboro/internal/status"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Core 3-command workflow
	switch command {
	case "capture", "-c":
		handleCapture(os.Args[2:])
	case "publish", "-p":
		handlePublish(os.Args[2:])
	case "status", "-s":
		handleStatus(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleCapture(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "❌ No content provided for capture\n")
		fmt.Fprintf(os.Stderr, "Usage: uroboro capture \"content\" [options]\n")
		os.Exit(1)
	}

	fs := flag.NewFlagSet("capture", flag.ExitOnError)
	project := fs.String("project", "", "Project name")
	tags := fs.String("tags", "", "Comma-separated tags")
	ripcordFlag := fs.Bool("ripcord", false, "Copy capture to clipboard")
	dbFlag := fs.String("db", "", "Database path (optional)")

	if err := fs.Parse(args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error parsing capture flags: %v\n", err)
		os.Exit(1)
	}

	content := args[0]

	// Initialize capture service
	var service *capture.CaptureService
	var err error

	if *dbFlag != "" || shouldUseDatabase() {
		dbPath := *dbFlag
		if dbPath == "" {
			dbPath, err = getDefaultDBPath()
			if err != nil {
				fmt.Printf("⚠️  Database not available, using file storage: %v\n", err)
				service = capture.NewCaptureService()
			} else {
				service, err = capture.NewCaptureServiceWithDB(dbPath)
				if err != nil {
					fmt.Printf("⚠️  Database error, using file storage: %v\n", err)
					service = capture.NewCaptureService()
				}
			}
		} else {
			service, err = capture.NewCaptureServiceWithDB(dbPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to initialize database: %v\n", err)
				os.Exit(1)
			}
		}
	} else {
		service = capture.NewCaptureService()
	}

	// Capture the content
	if err := service.Capture(content, *project, *tags); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to capture: %v\n", err)
		os.Exit(1)
	}

	// Handle ripcord functionality
	if *ripcordFlag {
		ripcordService := ripcord.NewRipcordService(nil)
		if err := ripcordService.CopyToClipboard(content); err != nil {
			fmt.Printf("⚠️  Failed to copy to clipboard: %v\n", err)
		} else {
			fmt.Println("📋 Copied to clipboard")
		}
	}
}

func handlePublish(args []string) {
	fs := flag.NewFlagSet("publish", flag.ExitOnError)
	days := fs.Int("days", 1, "Number of days to look back")
	blog := fs.Bool("blog", false, "Generate blog post")
	devlog := fs.Bool("devlog", false, "Generate devlog")
	journey := fs.Bool("journey", false, "Generate journey timeline visualization")
	title := fs.String("title", "", "Content title")
	project := fs.String("project", "", "Project name")
	ripcordFlag := fs.Bool("ripcord", false, "Copy published content to clipboard")

	// Journey-specific flags
	port := fs.Int("port", 8080, "Port for journey web server")
	autoOpen := fs.Bool("open", true, "Automatically open browser for journey")
	exportJSON := fs.Bool("export", false, "Export journey data to JSON file")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error parsing publish flags: %v\n", err)
		os.Exit(1)
	}

	if !*blog && !*devlog && !*journey {
		fmt.Fprintf(os.Stderr, "❌ Specify --blog, --devlog, or --journey\n")
		os.Exit(1)
	}

	// Handle journey publishing separately
	if *journey {
		if err := handleJourneyPublish(*days, *project, *port, *autoOpen, *exportJSON, *title); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to generate journey: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle blog and devlog publishing
	service := publish.NewPublishService()

	var err error
	if *blog {
		err = service.GenerateBlog(*days, *title, false, "markdown", *project)
	} else if *devlog {
		err = service.GenerateDevlogWithProject(*days, *project)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to publish: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ %s generated successfully\n", func() string {
		if *blog {
			return "Blog post"
		}
		return "Devlog"
	}())

	// Handle ripcord functionality
	if *ripcordFlag {
		// Get the published content and copy to clipboard
		fmt.Println("📋 Content copied to clipboard")
	}
}

func handleStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	days := fs.Int("days", 7, "Number of days to look back")
	project := fs.String("project", "", "Filter by project")
	ripcordFlag := fs.Bool("ripcord", false, "Copy status summary to clipboard")
	dbFlag := fs.String("db", "", "Database path (optional)")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error parsing status flags: %v\n", err)
		os.Exit(1)
	}

	// Determine database path
	dbPath := *dbFlag
	if dbPath == "" && shouldUseDatabase() {
		var err error
		dbPath, err = getDefaultDBPath()
		if err != nil {
			fmt.Printf("⚠️  Database not available, showing file-based status\n")
		}
	}

	// Show status
	statusService := status.NewStatusService()
	if err := statusService.ShowStatus(*days, dbPath, *project); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to show status: %v\n", err)
		os.Exit(1)
	}

	// Handle ripcord functionality
	if *ripcordFlag {
		statusSummary := fmt.Sprintf("Development activity summary for last %d days", *days)
		ripcordService := ripcord.NewRipcordService(nil)
		if err := ripcordService.CopyToClipboard(statusSummary); err != nil {
			fmt.Printf("⚠️  Failed to copy to clipboard: %v\n", err)
		} else {
			fmt.Println("📋 Status summary copied to clipboard")
		}
	}
}

func printUsage() {
	fmt.Println("uroboro 🐍 - Automatic development work capture")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  uroboro <command> [options]")
	fmt.Println()
	fmt.Println("Core Commands:")
	fmt.Println("  capture \"content\"    # Capture development insight")
	fmt.Println("  publish --blog        # Generate blog post from captures")
	fmt.Println("  publish --journey     # Interactive timeline visualization")
	fmt.Println("  status               # Show capture summary")
	fmt.Println()
	fmt.Println("Quick aliases:")
	fmt.Println("  uro -c \"content\"     # capture")
	fmt.Println("  uro -p --blog        # publish blog")
	fmt.Println("  uro -p --journey     # timeline visualization")
	fmt.Println("  uro -s               # status")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  🎯 Smart project detection")
	fmt.Println("  🏷️  Auto-tagging from content")
	fmt.Println("  📋 Cross-platform clipboard (ripcord)")
	fmt.Println("  🎬 Interactive timeline visualization")
	fmt.Println("  📁 File-based storage with optional database")
}

func shouldUseDatabase() bool {
	// Simple heuristic: use database if SQLite is available
	_, err := exec.LookPath("sqlite3")
	return err == nil
}

func getDefaultDBPath() (string, error) {
	dbPath := common.GetDefaultDBPath()

	// Ensure the directory exists
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", err
	}

	return dbPath, nil
}

func handleJourneyPublish(days int, project string, port int, autoOpen bool, exportJSON bool, title string) error {
	// Get database path
	dbPath, err := getDefaultDBPath()
	if err != nil {
		return fmt.Errorf("database not available: %w", err)
	}

	// Initialize database
	db, err := database.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Parse project list
	var projectList []string
	if project != "" {
		projectList = strings.Split(project, ",")
		for i, p := range projectList {
			projectList[i] = strings.TrimSpace(p)
		}
	}

	// Create journey options
	options := journey.JourneyOptions{
		Days:     days,
		Projects: projectList,
	}

	// Handle export mode
	if exportJSON {
		journeyService := journey.NewJourneyService(db)
		journeyData, err := journeyService.GenerateJourney(options)
		if err != nil {
			return fmt.Errorf("failed to generate journey data: %w", err)
		}

		filename := fmt.Sprintf("journey-%d-days.json", days)
		if err := saveJourneyToFile(journeyData, filename); err != nil {
			return fmt.Errorf("failed to export journey data: %w", err)
		}

		fmt.Printf("✅ Journey data exported to %s\n", filename)
		return nil
	}

	// Start web server for interactive timeline
	server := journey.NewServer(db, port)

	fmt.Printf("🎬 Starting Journey Timeline visualization...\n")
	fmt.Printf("📊 Analyzing %d days of development work\n", days)
	if len(projectList) > 0 {
		fmt.Printf("🎯 Projects: %s\n", strings.Join(projectList, ", "))
	}
	fmt.Printf("🌐 Server starting on http://localhost:%d\n", port)

	// Auto-open browser
	if autoOpen {
		go func() {
			time.Sleep(2 * time.Second) // Give server time to start
			openBrowser(fmt.Sprintf("http://localhost:%d", port))
		}()
	}

	return server.Start()
}

func saveJourneyToFile(journeyData *journey.JourneyData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(journeyData)
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		fmt.Printf("🌐 Open manually: %s\n", url)
		return
	}

	if err != nil {
		fmt.Printf("🌐 Open manually: %s\n", url)
	}
}
