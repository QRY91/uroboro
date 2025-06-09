package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/capture"
	"github.com/QRY91/uroboro/internal/config"
	"github.com/QRY91/uroboro/internal/database"
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
		
		// Check for existing personal database and offer migration
		if err := checkAndOfferMigration(); err != nil {
			fmt.Printf("⚠️  Migration check failed: %v\n", err)
		}
		
		// Process any pending ecosystem messages
		if err := edb.ProcessUroboroCaptures(); err != nil {
			fmt.Printf("⚠️  Warning: Failed to process ecosystem messages: %v\n", err)
		}
	} else {
		fmt.Printf("📁 Using local uroboro database\n")
	}
}

func handleCapture(args []string) {
	// Resolve database path using unified approach
	dbResult, cleanedArgs, err := resolveDBPath(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database resolution failed: %v\n", err)
		os.Exit(1)
	}

	if len(cleanedArgs) == 0 {
		fmt.Fprintf(os.Stderr, "❌ No content provided for capture\n")
		fmt.Fprintf(os.Stderr, "Usage: uroboro capture \"content\" [options]\n")
		os.Exit(1)
	}

	// Separate content from flags - content is first argument, rest are flags
	content := cleanedArgs[0]
	flagArgs := cleanedArgs[1:]

	fs := flag.NewFlagSet("capture", flag.ExitOnError)
	project := fs.String("project", "", "Project name")
	tags := fs.String("tags", "", "Comma-separated tags")
	link := fs.Bool("link", true, "Link to recent context (ecosystem mode)")

	fs.Parse(flagArgs)

	// Use ecosystem database if available
	if dbResult.Source == "ecosystem" {
		err := captureWithEcosystem(content, *project, *tags, *link)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Capture failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Use database or fall back to flat files based on resolution
	if dbResult.UseDB {
		fmt.Printf("📝 Using database: %s (%s)\n", dbResult.DBPath, dbResult.Source)
		legacyCapture(content, *project, *tags, dbResult.DBPath, true)
	} else {
		fmt.Printf("📝 Using flat file storage\n")
		legacyCapture(content, *project, *tags, "", false)
	}
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
				fmt.Printf("📝 Captured with context link: %s\n", truncateString(contexts[0].SessionInfo, 40))
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

func handlePublish(args []string) {
	// Resolve database path using unified approach
	dbResult, cleanedArgs, err := resolveDBPath(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database resolution failed: %v\n", err)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("publish", flag.ExitOnError)
	days := fs.Int("days", 1, "Number of days to look back")
	blog := fs.Bool("blog", false, "Generate blog post")
	devlog := fs.Bool("devlog", false, "Generate devlog")
	title := fs.String("title", "", "Blog post title")
	preview := fs.Bool("preview", false, "Preview content without saving")
	format := fs.String("format", "markdown", "Output format: markdown, html, text")
	project := fs.String("project", "", "Project name")

	fs.Parse(cleanedArgs)

	// Use ecosystem database if available
	if dbResult.Source == "ecosystem" {
		err := publishWithEcosystem(*days, *blog, *devlog, *title, *format, *project, *preview)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Publish failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Use database or fall back to flat files based on resolution
	if dbResult.UseDB {
		fmt.Printf("📝 Using database: %s (%s)\n", dbResult.DBPath, dbResult.Source)
		legacyPublish(cleanedArgs, dbResult.DBPath, true)
	} else {
		fmt.Printf("📝 Using flat file storage\n")
		legacyPublish(cleanedArgs, "", false)
	}
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
			fmt.Printf("🔗 Using context: %s\n", truncateString(contexts[0].SessionInfo, 40))
		}
	}

	// Generate AI-enhanced publication content
	service := publish.NewPublishService()
	var content string
	
	if devlog {
		// Convert ecosystem captures to activity strings for AI generation
		var activity []string
		for _, capture := range captures {
			activity = append(activity, capture.Content)
		}
		
		content, err = service.GenerateDevlogFromActivity(activity, format)
		if err != nil {
			return fmt.Errorf("failed to generate AI devlog: %w", err)
		}
		
		fmt.Println("--- GENERATED DEVLOG ---")
		fmt.Println(content)
		fmt.Println("--- END DEVLOG ---")
	} else {
		// Fallback to simple content generation for blog and other types
		content = generatePublicationContent(captures, blog, devlog, format)
	}
	
	if preview {
		fmt.Println("\n" + content)
		return nil
	}

	// Save to file if devlog
	if devlog {
		filename := fmt.Sprintf("devlog-%s.md", time.Now().Format("2006-01-02"))
		outputPath := filepath.Join("output", "posts", filename)
		
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		
		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to save devlog: %w", err)
		}
		
		fmt.Printf("📄 Saved devlog to: %s\n", outputPath)
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
		err = service.GenerateBlog(*days, *title, *preview, *format)
	} else if *devlog {
		err = service.GenerateDevlog(*days)
	} else {
		fmt.Fprintf(os.Stderr, "❌ Specify --blog or --devlog\n")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Publish failed: %v\n", err)
		os.Exit(1)
	}
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
	// Resolve database path using unified approach
	dbResult, _, err := resolveDBPath(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Database resolution failed: %v\n", err)
		os.Exit(1)
	}

	// Show database information
	if dbResult.Source == "ecosystem" {
		fmt.Printf("🗄️  Database: %s\n", dbResult.DBPath)
		if edb.IsShared() {
			fmt.Printf("🔗 Ecosystem mode: ENABLED\n")
		} else {
			fmt.Printf("📁 Ecosystem mode: DISABLED (local database)\n")
		}
		fmt.Println()
	} else if dbResult.UseDB {
		fmt.Printf("🗄️  Database: %s (%s)\n", dbResult.DBPath, dbResult.Source)
		fmt.Println()
	} else {
		fmt.Printf("📁 Storage: Flat files (no database configured)\n")
		fmt.Println()
	}

	// Use status logic with resolved database path
	service := status.NewStatusService()
	dbPath := ""
	if dbResult.UseDB {
		dbPath = dbResult.DBPath
	}
	
	if err := service.ShowStatus(7, dbPath); err != nil {
		fmt.Printf("⚠️  Status check failed: %v\n", err)
	}
}
func printUsage() {
	// Detect command name from how binary was called
	binaryName := filepath.Base(os.Args[0])

	fmt.Println("uroboro - The Self-Documenting Content Pipeline")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s capture \"your insight here\" [--project name] [--tags tag1,tag2] [--link]\n", binaryName)
	fmt.Printf("  %s publish --blog [--title \"Title\"] [--days N] [--preview] [--project name]\n", binaryName)
	fmt.Printf("  %s publish --devlog [--days N] [--project name]\n", binaryName)
	fmt.Printf("  %s status\n", binaryName)
	fmt.Printf("  %s process [--db]\n", binaryName)
	fmt.Printf("  %s sync           # Process ecosystem messages\n", binaryName)
	fmt.Printf("  %s config [--set-default-db]\n", binaryName)
	fmt.Println()
	fmt.Println("Ecosystem Options:")
	fmt.Printf("  --local          Force local database mode\n")
	fmt.Println()
	fmt.Println("Short flags:")
	fmt.Printf("  %s -c \"your insight here\"    # capture\n", binaryName)
	fmt.Printf("  %s -p --blog                  # publish blog\n", binaryName)
	fmt.Printf("  %s -s                         # status\n", binaryName)
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s capture \"Fixed auth timeout - cut query time from 3s to 200ms\"\n", binaryName)
	fmt.Printf("  %s capture \"Implemented OAuth2 with JWT tokens\" --project=myapp\n", binaryName)
	fmt.Printf("  %s publish --blog --title \"This Week's Fixes\" --project=myapp\n", binaryName)
	fmt.Printf("  %s sync\n", binaryName)
	fmt.Printf("  %s status\n", binaryName)
	fmt.Printf("  %s -s\n", binaryName)
	fmt.Println()
	fmt.Println("Ecosystem Integration:")
	fmt.Printf("  Captures automatically link to wherewasi context when available\n")
	fmt.Printf("  Notifications sent to examinator for flashcard generation\n")
	fmt.Printf("  Use 'sync' command to process messages from other QRY tools\n")
}

// getOrSetDefaultDBPath tries to get default database path from config,
// or prompts user to set one if not configured
func getOrSetDefaultDBPath() (string, error) {
	// Try to load from existing config first
	configuredPath, err := config.LoadDefaultDBPath()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	// If we have a configured path, use it
	if configuredPath != "" {
		fmt.Printf("🗄️  Using configured database: %s\n", configuredPath)
		return configuredPath, nil
	}

	// Otherwise, use the config system to get/set default (may prompt user)
	return config.GetDefaultDBPath()
}

func handleProcess(args []string) {
	// Parse --db flag
	dbPath := ""
	useDefaultDB := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--db" {
			useDefaultDB = true
		} else if len(arg) > 5 && arg[:5] == "--db=" {
			dbPath = arg[5:]
		}
	}

	var err error
	finalDBPath := dbPath
	if useDefaultDB && finalDBPath == "" {
		finalDBPath, err = getOrSetDefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Default database setup failed: %v\n", err)
			os.Exit(1)
		}
	}

	if finalDBPath != "" {
		captureService, err := capture.NewCaptureServiceWithDB(finalDBPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Database initialization failed: %v\n", err)
			os.Exit(1)
		}

		if err := captureService.ProcessToolMessages(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to process tool messages: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Tool messages processed")
	} else {
		fmt.Println("📝 Processing tool messages requires database mode")
		fmt.Println("Use: uroboro process --db")
	}
}

func handleConfig(args []string) {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	setDefaultDB := fs.Bool("set-default-db", false, "Set default database path")
	dbPath := fs.String("db-path", "", "Database path to set as default")
	autoYes := fs.Bool("yes", false, "Auto-accept default path without prompting")
	
	fs.Parse(args)

	if *setDefaultDB {
		var newPath string
		var err error

		if *dbPath != "" {
			// Use provided path
			newPath = *dbPath
			// Create directory if needed
			dbDir := filepath.Dir(newPath)
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to create database directory: %v\n", err)
				os.Exit(1)
			}
			// Save config
			cfg := &config.Config{DefaultDBPath: newPath}
			if err := config.SaveConfig(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to save config: %v\n", err)
				os.Exit(1)
			}
		} else if *autoYes {
			// Use default path without prompting
			defaultPath := filepath.Join(os.Getenv("HOME"), ".local", "share", "uroboro", "uroboro.sqlite")
			dbDir := filepath.Dir(defaultPath)
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to create database directory: %v\n", err)
				os.Exit(1)
			}
			cfg := &config.Config{DefaultDBPath: defaultPath}
			if err := config.SaveConfig(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to save config: %v\n", err)
				os.Exit(1)
			}
			newPath = defaultPath
		} else {
			// Interactive prompt
			newPath, err = config.PromptForDefaultDB()
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to set default database: %v\n", err)
				os.Exit(1)
			}
		}
		fmt.Printf("✅ Default database path set to: %s\n", newPath)
		return
	}

	// Show current configuration
	configuredPath, err := config.LoadDefaultDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🔧 UROBORO CONFIGURATION")
	fmt.Println("========================")
	
	if configuredPath != "" {
		fmt.Printf("Default database: %s\n", configuredPath)
	} else {
		fmt.Println("Default database: Not configured")
		fmt.Println("💡 Run 'uroboro config --set-default-db' to configure")
	}
	
	fmt.Printf("Config directory: %s\n", config.GetConfigPath())
}

// Utility functions for ecosystem integration

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
	return s[:maxLen] + "..."
}

// DBResolutionResult contains the resolved database path and metadata
type DBResolutionResult struct {
	DBPath     string
	Source     string // "ecosystem", "explicit", "default", "none"
	UseDB      bool   // true if should use database, false for flat files
}

// parseDBFlags extracts --db flag from args and returns cleaned args
func parseDBFlags(args []string) (dbPath string, explicitDB bool, cleanedArgs []string) {
	cleanedArgs = []string{}
	
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--db" {
			explicitDB = true
			// --db without value means use default
		} else if len(arg) > 5 && arg[:5] == "--db=" {
			dbPath = arg[5:]
			explicitDB = true
		} else if arg == "--local" {
			// Skip --local flag (handled in initialization)
			continue
		} else {
			cleanedArgs = append(cleanedArgs, arg)
		}
	}
	
	return dbPath, explicitDB, cleanedArgs
}

// resolveDBPath provides unified database path resolution across all commands
func resolveDBPath(args []string) (*DBResolutionResult, []string, error) {
	// Parse --db flags from arguments
	explicitDBPath, explicitDB, cleanedArgs := parseDBFlags(args)
	
	// Priority 1: Use explicit --db flag if provided (allows override of ecosystem)
	if explicitDB {
		finalDBPath := explicitDBPath
		
		// If --db flag used without value, get default
		if finalDBPath == "" {
			defaultPath, err := getOrSetDefaultDBPath()
			if err != nil {
				return nil, cleanedArgs, fmt.Errorf("default database setup failed: %w", err)
			}
			finalDBPath = defaultPath
		}
		
		return &DBResolutionResult{
			DBPath: finalDBPath,
			Source: "explicit",
			UseDB:  true,
		}, cleanedArgs, nil
	}
	
	// Priority 2: Use ecosystem database if available
	if edb != nil {
		return &DBResolutionResult{
			DBPath: edb.DatabasePath(),
			Source: "ecosystem",
			UseDB:  true,
		}, cleanedArgs, nil
	}
	
	// Priority 3: Use configured default database if available
	defaultPath, err := config.LoadDefaultDBPath()
	if err == nil && defaultPath != "" {
		return &DBResolutionResult{
			DBPath: defaultPath,
			Source: "default",
			UseDB:  true,
		}, cleanedArgs, nil
	}
	
	// Priority 4: Fall back to flat files
	return &DBResolutionResult{
		DBPath: "",
		Source: "none",
		UseDB:  false,
	}, cleanedArgs, nil
}

// checkAndOfferMigration checks for existing personal databases and offers migration
func checkAndOfferMigration() error {
	// Only run if we have an ecosystem database
	if edb == nil || !edb.IsShared() {
		return nil
	}

	personalDBPath, err := config.LoadDefaultDBPath()
	if err != nil || personalDBPath == "" {
		// No configured personal database, nothing to migrate
		return nil
	}

	// Check if personal database exists and has data
	if !fileExists(personalDBPath) {
		return nil
	}

	captureCount, err := countCapturesInDB(personalDBPath)
	if err != nil || captureCount == 0 {
		// No captures to migrate
		return nil
	}

	// Check if we've already migrated (to avoid repeated prompts)
	ecosystemCount, err := countCapturesInEcosystem()
	if err == nil && ecosystemCount > 0 {
		// Ecosystem already has data, assume migration already happened
		return nil
	}

	// Offer migration
	return promptForMigration(personalDBPath, captureCount)
}

// promptForMigration prompts user to migrate personal database to ecosystem
func promptForMigration(personalDBPath string, captureCount int) error {
	fmt.Printf("🔄 Found %d existing captures in personal database\n", captureCount)
	fmt.Printf("   Copy to ecosystem storage for tool integration? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" || input == "y" || input == "yes" {
		return performMigration(personalDBPath, captureCount)
	}

	fmt.Printf("📝 No migration performed. Use --db to access previous captures.\n")
	return nil
}

// performMigration copies captures from personal DB to ecosystem DB
func performMigration(personalDBPath string, captureCount int) error {
	fmt.Printf("📋 Copying %d captures to ecosystem storage...\n", captureCount)

	// Open personal database
	personalDB, err := database.NewDB(personalDBPath)
	if err != nil {
		return fmt.Errorf("failed to open personal database: %w", err)
	}
	defer personalDB.Close()

	// Get all captures from personal database
	captures, err := personalDB.GetAllCaptures()
	if err != nil {
		return fmt.Errorf("failed to read captures: %w", err)
	}

	// Copy each capture to ecosystem database
	copied := 0
	for _, capture := range captures {
		project := ""
		if capture.Project.Valid {
			project = capture.Project.String
		}
		
		tags := ""
		if capture.Tags.Valid {
			tags = capture.Tags.String
		}

		if _, err := edb.InsertCaptureWithTimestamp(capture.Content, project, tags, capture.CreatedAt); err != nil {
			fmt.Printf("⚠️  Warning: Failed to copy capture: %v\n", err)
			continue
		}
		copied++
	}

	if copied == captureCount {
		fmt.Printf("✅ Successfully copied %d captures to ecosystem storage\n", copied)
	} else {
		fmt.Printf("⚠️  Copied %d of %d captures (some failures)\n", copied, captureCount)
	}

	// Ask about cleanup
	return promptForCleanup(personalDBPath)
}

// promptForCleanup asks user if they want to keep or remove the original database
func promptForCleanup(personalDBPath string) error {
	fmt.Printf("   Keep original database? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "n" || input == "no" {
		if err := os.Remove(personalDBPath); err != nil {
			fmt.Printf("⚠️  Warning: Could not remove original database: %v\n", err)
			fmt.Printf("   You can manually delete: %s\n", personalDBPath)
		} else {
			fmt.Printf("🗑️  Removed original database\n")
		}
		
		// Clear the config entry
		configObj := &config.Config{DefaultDBPath: ""}
		if err := config.SaveConfig(configObj); err != nil {
			fmt.Printf("⚠️  Warning: Could not clear config: %v\n", err)
		}
	} else {
		fmt.Printf("📁 Original database preserved at: %s\n", personalDBPath)
		fmt.Printf("   Use --db to access it when needed\n")
	}

	return nil
}

// countCapturesInDB counts captures in a specific database file
func countCapturesInDB(dbPath string) (int, error) {
	db, err := database.NewDB(dbPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	captures, err := db.GetAllCaptures()
	if err != nil {
		return 0, err
	}

	return len(captures), nil
}

// countCapturesInEcosystem counts captures in the ecosystem database
func countCapturesInEcosystem() (int, error) {
	if edb == nil {
		return 0, fmt.Errorf("no ecosystem database")
	}

	captures, err := edb.GetRecentCaptures(999999, "") // Get all captures
	if err != nil {
		return 0, err
	}

	return len(captures), nil
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
