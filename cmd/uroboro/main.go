package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/QRY91/uroboro/internal/analytics"
	"github.com/QRY91/uroboro/internal/capture"
	"github.com/QRY91/uroboro/internal/config"
	"github.com/QRY91/uroboro/internal/context"
	"github.com/QRY91/uroboro/internal/database"
	"github.com/QRY91/uroboro/internal/publish"
	"github.com/QRY91/uroboro/internal/ripcord"
	"github.com/QRY91/uroboro/internal/status"
	"github.com/QRY91/uroboro/internal/tagging"
)

func main() {
	// Initialize PostHog analytics
	analytics.Initialize()
	defer analytics.Get().Close()

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
	case "config":
		handleConfig(os.Args[2:])
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

	content := args[0]

	fs := flag.NewFlagSet("capture", flag.ExitOnError)
	project := fs.String("project", "", "Project name")
	tags := fs.String("tags", "", "Comma-separated tags")
	ripcordFlag := fs.Bool("ripcord", false, "Copy enriched context to clipboard after capture")
	dbFlag := fs.String("db", "", "Database path (optional)")

	if err := fs.Parse(args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error parsing capture flags: %v\n", err)
		os.Exit(1)
	}

	// Smart project detection if no project provided
	if *project == "" {
		detector := context.NewProjectDetector()
		if detectedProject := detector.DetectProject(); detectedProject != "" {
			*project = detectedProject
			fmt.Printf("🔍 Auto-detected project: %s\n", *project)
		}
	}

	// Smart auto-tagging
	if *tags == "" {
		analyzer := tagging.NewTagAnalyzer()
		if autoTags := analyzer.AnalyzeTags(content); len(autoTags) > 0 {
			*tags = strings.Join(autoTags, ",")
			fmt.Printf("🏷️  Auto-detected tags: %s\n", *tags)
		}
	}

	// Try database first, fall back to files
	var err error
	if *dbFlag != "" || shouldUseDatabase() {
		dbPath := *dbFlag
		if dbPath == "" {
			dbPath, err = getDefaultDBPath()
			if err != nil {
				fmt.Printf("⚠️  Database setup failed, using file storage: %v\n", err)
				err = captureToFile(content, *project, *tags)
			} else {
				fmt.Printf("🗄️  Using configured database: %s\n", dbPath)
				err = captureToDatabase(content, *project, *tags, dbPath)
			}
		} else {
			fmt.Printf("🗄️  Using database: %s\n", dbPath)
			err = captureToDatabase(content, *project, *tags, dbPath)
		}
	} else {
		fmt.Printf("📁 Using file storage\n")
		err = captureToFile(content, *project, *tags)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Capture failed: %v\n", err)
		os.Exit(1)
	}

	// Track successful capture
	tagsList := strings.Split(*tags, ",")
	if *tags == "" {
		tagsList = []string{}
	}
	analytics.Get().TrackCaptureSimple(content, *project, tagsList)

	// Handle ripcord functionality
	if *ripcordFlag {
		var ripcordService *ripcord.RipcordService
		if *dbFlag != "" {
			if db, err := database.NewDB(*dbFlag); err == nil {
				ripcordService = ripcord.NewRipcordService(db)
			}
		} else {
			ripcordService = ripcord.NewRipcordService(nil)
		}

		if ripcordService != nil {
			err := ripcordService.QuickRipcord()
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Ripcord failed: %v\n", err)
			}
		}
	}
}

func handlePublish(args []string) {
	fs := flag.NewFlagSet("publish", flag.ExitOnError)
	days := fs.Int("days", 1, "Number of days to look back")
	blog := fs.Bool("blog", false, "Generate blog post")
	devlog := fs.Bool("devlog", false, "Generate devlog")
	title := fs.String("title", "", "Blog post title")
	preview := fs.Bool("preview", false, "Preview content without saving")
	format := fs.String("format", "markdown", "Output format: markdown, html, text")
	project := fs.String("project", "", "Project name")
	ripcordFlag := fs.Bool("ripcord", false, "Copy published content to clipboard")
	dbFlag := fs.String("db", "", "Database path (optional)")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error parsing publish flags: %v\n", err)
		os.Exit(1)
	}

	if !*blog && !*devlog {
		fmt.Fprintf(os.Stderr, "❌ Specify --blog or --devlog\n")
		os.Exit(1)
	}

	// Try database first, fall back to files
	var service *publish.PublishService
	var err error

	if *dbFlag != "" || shouldUseDatabase() {
		dbPath := *dbFlag
		if dbPath == "" {
			dbPath, err = getDefaultDBPath()
			if err != nil {
				fmt.Printf("⚠️  Database not available, using file storage: %v\n", err)
				service = publish.NewPublishService()
			} else {
				fmt.Printf("🗄️  Using configured database: %s\n", dbPath)
				service, err = publish.NewPublishServiceWithDB(dbPath)
			}
		} else {
			fmt.Printf("🗄️  Using database: %s\n", dbPath)
			service, err = publish.NewPublishServiceWithDB(dbPath)
		}
	} else {
		fmt.Printf("📁 Using file storage\n")
		service = publish.NewPublishService()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Service initialization failed: %v\n", err)
		os.Exit(1)
	}

	if *blog {
		err = service.GenerateBlog(*days, *title, *preview, *format, *project)
	} else if *devlog {
		err = service.GenerateDevlogWithProject(*days, *project)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Publish failed: %v\n", err)
		analytics.Get().TrackPublishSimple(*format, false, 0)
		os.Exit(1)
	}

	// Track successful publish
	analytics.Get().TrackPublishSimple(*format, true, 1000) // TODO: get actual word count

	// Handle ripcord functionality
	if *ripcordFlag {
		var ripcordService *ripcord.RipcordService
		if *dbFlag != "" {
			if db, err := database.NewDB(*dbFlag); err == nil {
				ripcordService = ripcord.NewRipcordService(db)
			}
		} else {
			ripcordService = ripcord.NewRipcordService(nil)
		}

		if ripcordService != nil {
			err := ripcordService.WorkRipcord(*days)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Ripcord failed: %v\n", err)
			}
		}
	}
}

func handleStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	days := fs.Int("days", 7, "Number of days to look back")
	ripcordFlag := fs.Bool("ripcord", false, "Copy status summary to clipboard")
	dbFlag := fs.String("db", "", "Database path (optional)")
	project := fs.String("project", "", "Project name")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error parsing status flags: %v\n", err)
		os.Exit(1)
	}

	// Try database first, fall back to files
	dbPath := ""
	if *dbFlag != "" || shouldUseDatabase() {
		if *dbFlag != "" {
			dbPath = *dbFlag
			fmt.Printf("🗄️  Database: %s\n", dbPath)
		} else {
			var err error
			dbPath, err = getDefaultDBPath()
			if err != nil {
				fmt.Printf("📁 Using file storage\n")
			} else {
				fmt.Printf("🗄️  Database: %s\n", dbPath)
			}
		}
	} else {
		fmt.Printf("📁 Using file storage\n")
	}

	fmt.Println()

	service := status.NewStatusService()
	err := service.ShowStatus(*days, dbPath, *project)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Status failed: %v\n", err)
		os.Exit(1)
	}

	// Track status check
	analytics.Get().TrackStatusCheck(0, 0, 1, 0, 0.0) // TODO: get actual metrics

	// Handle ripcord functionality
	if *ripcordFlag {
		var ripcordService *ripcord.RipcordService
		if dbPath != "" {
			if db, err := database.NewDB(dbPath); err == nil {
				ripcordService = ripcord.NewRipcordService(db)
			}
		} else {
			ripcordService = ripcord.NewRipcordService(nil)
		}

		if ripcordService != nil {
			err := ripcordService.WorkRipcord(*days)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Ripcord failed: %v\n", err)
			}
		}
	}
}

func handleConfig(args []string) {
	if len(args) == 0 {
		// Show current config
		dbPath, err := config.LoadDefaultDBPath()
		if err != nil {
			fmt.Printf("❌ Failed to load config: %v\n", err)
			os.Exit(1)
		}

		if dbPath == "" {
			fmt.Println("📁 No default database configured (using file storage)")
		} else {
			fmt.Printf("🗄️  Default database: %s\n", dbPath)
		}
		return
	}

	command := args[0]
	switch command {
	case "set-db":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "❌ Usage: uroboro config set-db <path>\n")
			os.Exit(1)
		}
		dbPath := args[1]

		// Expand relative paths to absolute
		if !filepath.IsAbs(dbPath) {
			abs, err := filepath.Abs(dbPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Invalid path: %v\n", err)
				os.Exit(1)
			}
			dbPath = abs
		}

		err := config.SaveDefaultDBPath(dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to save config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Default database set to: %s\n", dbPath)

	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown config command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Available commands: set-db\n")
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("uroboro - The Unified Development Assistant")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  uroboro capture \"content\" [flags]    # Capture development insights")
	fmt.Println("  uroboro publish [flags]               # Generate content from captures")
	fmt.Println("  uroboro status [flags]                # Show development pipeline status")
	fmt.Println("  uroboro config [command]              # Configure uroboro")
	fmt.Println()
	fmt.Println("Short aliases:")
	fmt.Println("  uro -c \"content\"    # capture")
	fmt.Println("  uro -p --devlog      # publish devlog")
	fmt.Println("  uro -s              # status")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  🧠 Smart project detection")
	fmt.Println("  🏷️  Auto-tagging")
	fmt.Println("  📋 Ripcord functionality")
	fmt.Println("  🗄️  Optional database storage")
	fmt.Println("  📁 File-based fallback")
}

// Helper functions

func captureToDatabase(content, project, tags, dbPath string) error {
	service, err := capture.NewCaptureServiceWithDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database service: %w", err)
	}

	err = service.CaptureWithMetadata(content, project, tags)
	if err != nil {
		return fmt.Errorf("database capture failed: %w", err)
	}

	fmt.Printf("✅ Captured to database\n")
	return nil
}

func captureToFile(content, project, tags string) error {
	service := capture.NewCaptureService()
	err := service.CaptureWithMetadata(content, project, tags)
	if err != nil {
		return fmt.Errorf("file capture failed: %w", err)
	}

	fmt.Printf("✅ Captured to file\n")
	return nil
}

func shouldUseDatabase() bool {
	// Check if we have a configured default database
	dbPath, err := config.LoadDefaultDBPath()
	return err == nil && dbPath != ""
}

func getDefaultDBPath() (string, error) {
	dbPath, err := config.LoadDefaultDBPath()
	if err != nil {
		return "", fmt.Errorf("failed to load default database path: %w", err)
	}

	if dbPath == "" {
		return "", fmt.Errorf("no default database configured")
	}

	return dbPath, nil
}
