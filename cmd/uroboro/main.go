package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/QRY91/uroboro/internal/capture"
	"github.com/QRY91/uroboro/internal/publish"
	"github.com/QRY91/uroboro/internal/status"
)

func main() {
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
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleCapture(args []string) {
	// Parse --db flag manually to support both --db and --db=path
	dbPath := ""
	useDefaultDB := false
	filteredArgs := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--db" {
			// --db without value, use default
			useDefaultDB = true
		} else if len(arg) > 5 && arg[:5] == "--db=" {
			// --db=path format
			dbPath = arg[5:]
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	fs := flag.NewFlagSet("capture", flag.ExitOnError)
	project := fs.String("project", "", "Project name")
	tags := fs.String("tags", "", "Comma-separated tags")

	fs.Parse(filteredArgs)

	if fs.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Error: No content provided\n")
		fmt.Fprintf(os.Stderr, "Usage: uroboro capture \"your insight here\"\n")
		os.Exit(1)
	}

	content := fs.Arg(0)

	var service *capture.CaptureService
	var err error

	// Handle database path logic
	finalDBPath := dbPath
	if useDefaultDB && finalDBPath == "" {
		// User specified --db without value, get default path
		defaultPath, err := getOrSetDefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Default database setup failed: %v\n", err)
			os.Exit(1)
		}
		finalDBPath = defaultPath
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

	if err := service.Capture(content, *project, *tags); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Capture failed: %v\n", err)
		os.Exit(1)
	}
}

func handlePublish(args []string) {
	// Parse --db flag manually to support both --db and --db=path
	dbPath := ""
	useDefaultDB := false
	filteredArgs := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--db" {
			// --db without value, use default
			useDefaultDB = true
		} else if len(arg) > 5 && arg[:5] == "--db=" {
			// --db=path format
			dbPath = arg[5:]
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

	fs.Parse(filteredArgs)

	var service *publish.PublishService
	var err error

	// Handle database path logic
	finalDBPath := dbPath
	if useDefaultDB && finalDBPath == "" {
		// User specified --db without a value, use default
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
		if err := service.GenerateBlog(*days, *title, *preview, *format); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Blog generation failed: %v\n", err)
			os.Exit(1)
		}
	} else if *devlog {
		if err := service.GenerateDevlog(*days); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Devlog generation failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: Specify --blog or --devlog\n")
		fmt.Fprintf(os.Stderr, "Usage: uroboro publish --blog [--title \"Title\"] [--days N] [--format FORMAT] [--db PATH]\n")
		fmt.Fprintf(os.Stderr, "       uroboro publish --devlog [--days N] [--db PATH]\n")
		fmt.Fprintf(os.Stderr, "Formats: markdown (default), html, text\n")
		os.Exit(1)
	}
}

func handleStatus(args []string) {
	// Parse --db flag manually to support both --db and --db=path
	dbPath := ""
	useDefaultDB := false
	filteredArgs := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--db" {
			// --db without value, use default
			useDefaultDB = true
		} else if len(arg) > 5 && arg[:5] == "--db=" {
			// --db=path format
			dbPath = arg[5:]
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	fs := flag.NewFlagSet("status", flag.ExitOnError)
	days := fs.String("days", "7", "Number of days to show")

	fs.Parse(filteredArgs)

	daysInt, err := strconv.Atoi(*days)
	if err != nil {
		daysInt = 7
	}

	// Handle database path logic
	finalDBPath := dbPath
	if useDefaultDB && finalDBPath == "" {
		// User specified --db without value, get default path
		defaultPath, err := getOrSetDefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Default database setup failed: %v\n", err)
			os.Exit(1)
		}
		finalDBPath = defaultPath
	}

	service := status.NewStatusService()
	if err := service.ShowStatus(daysInt, finalDBPath); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Status check failed: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	// Detect command name from how binary was called
	binaryName := filepath.Base(os.Args[0])

	fmt.Println("uroboro - The Self-Documenting Content Pipeline")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s capture \"your insight here\" [--project name] [--tags tag1,tag2]\n", binaryName)
	fmt.Printf("  %s publish --blog [--title \"Title\"] [--days N] [--preview]\n", binaryName)
	fmt.Printf("  %s publish --devlog [--days N]\n", binaryName)
	fmt.Printf("  %s status [--days N]\n", binaryName)
	fmt.Println()
	fmt.Println("Short flags:")
	fmt.Printf("  %s -c \"your insight here\"    # capture\n", binaryName)
	fmt.Printf("  %s -p --blog                  # publish blog\n", binaryName)
	fmt.Printf("  %s -s                         # status\n", binaryName)
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s capture \"Fixed auth timeout - cut query time from 3s to 200ms\"\n", binaryName)
	fmt.Printf("  %s -c \"Implemented OAuth2 with JWT tokens\"\n", binaryName)
	fmt.Printf("  %s publish --blog --title \"This Week's Fixes\"\n", binaryName)
	fmt.Printf("  %s -p --blog\n", binaryName)
	fmt.Printf("  %s status\n", binaryName)
	fmt.Printf("  %s -s\n", binaryName)
}

// getOrSetDefaultDBPath tries to get default database path from config,
// or prompts user to set one if not configured
func getOrSetDefaultDBPath() (string, error) {
	// For now, use a simple approach - just return the standard XDG path
	// Later we can implement the interactive config system
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	defaultPath := filepath.Join(homeDir, ".local", "share", "uroboro", "uroboro.sqlite")

	// Create directory if needed
	dbDir := filepath.Dir(defaultPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create database directory: %w", err)
	}

	fmt.Printf("🗄️  Using default database: %s\n", defaultPath)
	return defaultPath, nil
}
