package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/QRY91/uroboro/internal/capture"
	"github.com/QRY91/uroboro/internal/config"
	"github.com/QRY91/uroboro/internal/ecosystem"
	"github.com/QRY91/uroboro/internal/publish"
	"github.com/QRY91/uroboro/internal/status"
)

// Global ecosystem database instance
var edb *ecosystem.EcosystemDB

func main() {
	// Initialize ecosystem database first
	initializeEcosystemDatabase()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Support short flags for core commands
	switch command {
	case "capture", "-c":
		handleCapture(os.Args[2:])
	case "publish", "-p":
		handlePublish(os.Args[2:])
	case "status", "-s":
		handleStatus(os.Args[2:])
	case "process":
		handleProcess(os.Args[2:])
	case "sync":
		handleEcosystemSync(os.Args[2:])
	case "config":
		handleConfig(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func initializeEcosystemDatabase() {
	// Check for force-local flag
	forceLocal := false
	for _, arg := range os.Args {
		if arg == "--local" {
			forceLocal = true
			break
		}
	}

	// Get fallback path using existing config system
	fallbackPath, err := getOrCreateDefaultDBPath()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not determine fallback database path: %v\n", err)
		fallbackPath = "uroboro.sqlite" // Last resort
	}

	config := ecosystem.DatabaseConfig{
		ToolName:     ecosystem.ToolUroboro,
		FallbackPath: fallbackPath,
		ForceLocal:   forceLocal,
	}

	edb, err = ecosystem.NewEcosystemDB(config)
	if err != nil {
		fmt.Printf("⚠️  Failed to initialize ecosystem database: %v\n", err)
		fmt.Printf("    Continuing without ecosystem features...\n")
		return
	}

	if edb.IsShared() {
		fmt.Printf("🔗 Connected to QRY ecosystem database\n")
		fmt.Printf("    Database: %s\n", edb.DatabasePath())
		
		// Process any pending ecosystem messages
		if err := edb.ProcessUroboroCaptures(); err != nil {
			fmt.Printf("⚠️  Warning: Failed to process ecosystem messages: %v\n", err)
		}
	} else {
		fmt.Printf("📁 Using local uroboro database\n")
		fmt.Printf("    Database: %s\n", edb.DatabasePath())
		fmt.Printf("    Tip: Enable ecosystem features with shared database\n")
	}
}

func handleCapture(args []string) {
	// Parse flags with ecosystem awareness
	dbPath := ""
	useDefaultDB := false
	filteredArgs := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--db" {
			useDefaultDB = true
		} else if len(arg) > 5 && arg[:5] == "--db=" {
			dbPath = arg[5:]
		} else if arg == "--local" {
			// Skip --local flag (handled in initialization)
			continue
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	fs := flag.NewFlagSet("capture", flag.ExitOnError)
	project := fs.String("project", "", "Project name")
	tags := fs.String("tags", "", "Comma-separated tags")
	link := fs.Bool("link", true, "Link to recent context (ecosystem mode)")

	fs.Parse(filteredArgs)

	if len(fs.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "❌ No content provided for capture\n")
		fmt.Fprintf(os.Stderr, "Usage: uroboro capture \"content\" [options]\n")
		os.Exit(1)
	}

	content := fs.Args()[0]

	// Use ecosystem database if available, otherwise fall back to legacy system
	if edb != nil {
		err := captureWithEcosystem(content, *project, *tags, *link)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Capture failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Legacy capture logic (backward compatibility)
	fmt.Printf("📝 Using legacy capture mode\n")
	legacyCapture(content, *project, *tags, dbPath, useDefaultDB)
}

func captureWithEcosystem(content, project, tags string, linkContext bool) error {
	var capture *ecosystem.Capture
	var err error

	// Try to link to recent context if in ecosystem mode and linking enabled
	if edb.IsShared() && linkContext && project != "" {
		contexts, err := edb.GetRecentContexts(project, 1)
		if err == nil && len(contexts) > 0 {
			// Link capture to most recent context session
			capture, err = edb.InsertCaptureWithContext(content, project, tags, &contexts[0].ID)
			if err == nil {
				fmt.Printf("📝 Captured with context link: %s\n", contexts[0].SessionInfo)
			}
		}
	}

	// If context linking failed or wasn't attempted, create normal capture
	if capture == nil {
		capture, err = edb.InsertCapture(content, project, tags)
		if err != nil {
			return fmt.Errorf("failed to create capture: %w", err)
		}
		fmt.Printf("📝 Captured: %s\n", truncateString(content, 50))
	}

	// Send capture message to examinator for potential flashcard generation
	if edb.IsShared() {
		captureData := ecosystem.CaptureMessageData{
			Content: content,
			Project: project,
			Tags:    tags,
		}

		msg, err := ecosystem.NewToolMessage(
			ecosystem.ToolUroboro,
			ecosystem.ToolExaminator,
			ecosystem.MessageTypeCapture,
			captureData,
		)

		if err == nil {
			err = edb.SendToolMessage(msg.FromTool, msg.ToTool, msg.MessageType, msg.Data)
			if err != nil {
				fmt.Printf("⚠️  Warning: Failed to notify examinator: %v\n", err)
			} else {
				fmt.Printf("🎯 Notified examinator for potential flashcard generation\n")
			}
		}
	}

	// Track project activity in ecosystem
	if project != "" {
		err = edb.TrackProject(project, getCurrentDir(), ecosystem.ToolUroboro, hasGitRepo("."))
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to track project activity: %v\n", err)
		}
	}

	fmt.Printf("✅ Capture complete (ID: %d)\n", capture.ID)
	return nil
}

func handlePublish(args []string) {
	// Similar ecosystem-aware publish logic
	dbPath := ""
	useDefaultDB := false
	filteredArgs := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--db" {
			useDefaultDB = true
		} else if len(arg) > 5 && arg[:5] == "--db=" {
			dbPath = arg[5:]
		} else if arg == "--local" {
			continue
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	fs := flag.NewFlagSet("publish", flag.ExitOnError)
	days := fs.Int("days", 1, "Number of days to look back")
	blog := fs.Bool("blog", false, "Generate blog post")
	devlog := fs.Bool("devlog", false, "Generate devlog")
	title := fs.String("title", "", "Blog post title")
	preview := fs.Bool("preview", false, "Preview content without saving")
	format := fs.String("format", "markdown", "Output format: markdown, html, text")
	project := fs.String("project", "", "Project name")

	fs.Parse(filteredArgs)

	if edb != nil {
		err := publishWithEcosystem(*days, *blog, *devlog, *title, *preview, *format, *project)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Publish failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Legacy publish logic
	fmt.Printf("📝 Using legacy publish mode\n")
	legacyPublish(filteredArgs, dbPath, useDefaultDB)
}

func publishWithEcosystem(days int, blog, devlog bool, title, format, project string, preview bool) error {
	// Get recent captures using ecosystem database
	captures, err := edb.GetRecentCaptures(days, project)
	if err != nil {
		return fmt.Errorf("failed to get recent captures: %w", err)
	}

	if len(captures) == 0 {
		fmt.Printf("📝 No captures found in the last %d days", days)
		if project != "" {
			fmt.Printf(" for project '%s'", project)
		}
		fmt.Println()
		return nil
	}

	fmt.Printf("📚 Found %d captures to publish\n", len(captures))

	// Enhanced publishing with context awareness
	if edb.IsShared() && project != "" {
		// Get recent context to improve publication quality
		contexts, err := edb.GetRecentContexts(project, 1)
		if err == nil && len(contexts) > 0 {
			fmt.Printf("🔗 Using context: %s\n", contexts[0].SessionInfo)
		}
	}

	// Generate publication content
	content := generatePublicationContent(captures, blog, devlog, format)
	
	if preview {
		fmt.Println("\n" + content)
		return nil
	}

	// Save publication if not preview
	var captureIDs []int64
	for _, capture := range captures {
		captureIDs = append(captureIDs, capture.ID)
	}

	pubType := "devlog"
	if blog {
		pubType = "blog"
	}

	publication, err := edb.InsertPublication(title, content, format, pubType, project, "", captureIDs)
	if err != nil {
		return fmt.Errorf("failed to save publication: %w", err)
	}

	fmt.Printf("✅ Published %s (ID: %d)\n", pubType, publication.ID)
	return nil
}

func handleEcosystemSync(args []string) {
	if edb == nil {
		fmt.Printf("❌ Ecosystem database not available\n")
		os.Exit(1)
	}

	if !edb.IsShared() {
		fmt.Printf("📁 Local database mode - no ecosystem sync needed\n")
		return
	}

	fmt.Printf("🔄 Processing ecosystem messages...\n")
	
	err := edb.ProcessUroboroCaptures()
	if err != nil {
		fmt.Printf("❌ Ecosystem sync failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Ecosystem sync complete\n")
}

func handleStatus(args []string) {
	if edb != nil {
		fmt.Printf("🗄️  Database: %s\n", edb.DatabasePath())
		if edb.IsShared() {
			fmt.Printf("🔗 Ecosystem mode: ENABLED\n")
		} else {
			fmt.Printf("📁 Ecosystem mode: DISABLED (local database)\n")
		}
		fmt.Println()
	}

	// Use existing status logic or ecosystem-enhanced version
	status.ShowStatus()
}

func handleProcess(args []string) {
	if edb != nil && edb.IsShared() {
		fmt.Printf("🔄 Processing ecosystem messages...\n")
		err := edb.ProcessUroboroCaptures()
		if err != nil {
			fmt.Printf("❌ Message processing failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Legacy process logic for backward compatibility
	fmt.Printf("📝 Processing local messages...\n")
	// ... existing process logic
}

func handleConfig(args []string) {
	if len(args) == 0 {
		printConfigUsage()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "set-default-db":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "❌ Database path required\n")
			os.Exit(1)
		}
		if err := config.SetDefaultDBPath(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to set default database: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Default database set to: %s\n", args[1])
	case "enable-ecosystem":
		sharedPath, err := ecosystem.SharedDatabasePath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to get ecosystem database path: %v\n", err)
			os.Exit(1)
		}
		if err := config.SetDefaultDBPath(sharedPath); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to enable ecosystem mode: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("🔗 Ecosystem mode enabled\n")
		fmt.Printf("    Database: %s\n", sharedPath)
		fmt.Printf("    Restart uroboro to activate ecosystem features\n")
	case "disable-ecosystem":
		localPath, err := getOrCreateDefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to get local database path: %v\n", err)
			os.Exit(1)
		}
		if err := config.SetDefaultDBPath(localPath); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to disable ecosystem mode: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("📁 Ecosystem mode disabled\n")
		fmt.Printf("    Database: %s\n", localPath)
	default:
		printConfigUsage()
	}
}

// Utility functions

func legacyCapture(content, project, tags, dbPath string, useDefaultDB bool) {
	// Existing capture logic for backward compatibility
	var service *capture.CaptureService
	var err error

	finalDBPath := dbPath
	if useDefaultDB && finalDBPath == "" {
		defaultPath, err := getOrSetDefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Default database setup failed: %v\n", err)
			os.Exit(1)
		}
		finalDBPath = defaultPath
	} else if !useDefaultDB && finalDBPath == "" {
		configuredPath, err := config.LoadDefaultDBPath()
		if err == nil && configuredPath != "" {
			finalDBPath = configuredPath
		}
	}

	if finalDBPath != "" {
		service, err = capture.NewCaptureServiceWithDB(finalDBPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Database initialization failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		service = capture.NewCaptureService()
	}

	if err := service.Capture(content, project, tags); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Capture failed: %v\n", err)
		os.Exit(1)
	}
}

func legacyPublish(args []string, dbPath string, useDefaultDB bool) {
	// Existing publish logic for backward compatibility
	fs := flag.NewFlagSet("publish", flag.ExitOnError)
	days := fs.Int("days", 1, "Number of days to look back")
	blog := fs.Bool("blog", false, "Generate blog post")
	devlog := fs.Bool("devlog", false, "Generate devlog")
	title := fs.String("title", "", "Blog post title")
	preview := fs.Bool("preview", false, "Preview content without saving")
	format := fs.String("format", "markdown", "Output format: markdown, html, text")

	fs.Parse(args)

	var service *publish.PublishService
	var err error

	finalDBPath := dbPath
	if useDefaultDB && finalDBPath == "" {
		defaultPath, err := getOrSetDefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Default database setup failed: %v\n", err)
			os.Exit(1)
		}
		finalDBPath = defaultPath
	}

	if finalDBPath != "" {
		service, err = publish.NewPublishServiceWithDB(finalDBPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Database initialization failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		service = publish.NewPublishService()
	}

	if *blog {
		err = service.GenerateBlogPost(*days, *title, *preview, *format)
	} else if *devlog {
		err = service.GenerateDevlog(*days, *title, *preview, *format)
	} else {
		fmt.Fprintf(os.Stderr, "❌ Specify --blog or --devlog\n")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Publish failed: %v\n", err)
		os.Exit(1)
	}
}

func generatePublicationContent(captures []ecosystem.Capture, blog, devlog bool, format string) string {
	// Enhanced publication generation with ecosystem data
	content := ""
	
	if blog {
		content += "# Blog Post\n\n"
	} else if devlog {
		content += "# Development Log\n\n"
	}

	for _, capture := range captures {
		timestamp := capture.Timestamp.Format("2006-01-02 15:04")
		content += fmt.Sprintf("## %s\n", timestamp)
		
		if capture.Project != nil && *capture.Project != "" {
			content += fmt.Sprintf("**Project:** %s\n", *capture.Project)
		}
		
		if capture.Tags != nil && *capture.Tags != "" {
			content += fmt.Sprintf("**Tags:** %s\n", *capture.Tags)
		}
		
		content += fmt.Sprintf("\n%s\n\n", capture.Content)
	}

	return content
}

func getOrCreateDefaultDBPath() (string, error) {
	// Try to load existing config first
	if path, err := config.LoadDefaultDBPath(); err == nil && path != "" {
		return path, nil
	}

	// Fall back to creating default path
	return getOrSetDefaultDBPath()
}

func getOrSetDefaultDBPath() (string, error) {
	// Existing logic from original main.go
	defaultPath, err := config.GetDefaultDBPath()
	if err != nil {
		return "", err
	}

	if err := config.SetDefaultDBPath(defaultPath); err != nil {
		return "", err
	}

	return defaultPath, nil
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

func hasGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	_, err := os.Stat(gitPath)
	return err == nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func printUsage() {
	fmt.Println("uroboro - Content capture and publishing system")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  uroboro <command> [options]")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  capture, -c    Capture content")
	fmt.Println("  publish, -p    Publish captured content")
	fmt.Println("  status, -s     Show system status")
	fmt.Println("  process        Process ecosystem messages")
	fmt.Println("  sync           Sync with ecosystem database")
	fmt.Println("  config         Configuration management")
	fmt.Println()
	fmt.Println("ECOSYSTEM OPTIONS:")
	fmt.Println("  --local        Force local database mode")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  uroboro capture \"Fixed bug in authentication\"")
	fmt.Println("  uroboro capture \"Added new feature\" --project=myapp --tags=feature,auth")
	fmt.Println("  uroboro publish --blog --days=7 --title=\"Weekly Progress\"")
	fmt.Println("  uroboro sync")
	fmt.Println("  uroboro config enable-ecosystem")
}

func printConfigUsage() {
	fmt.Println("Configuration management")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  uroboro config <subcommand> [options]")
	fmt.Println()
	fmt.Println("SUBCOMMANDS:")
	fmt.Println("  set-default-db <path>    Set default database path")
	fmt.Println("  enable-ecosystem         Enable ecosystem mode")
	fmt.Println("  disable-ecosystem        Disable ecosystem mode")
}