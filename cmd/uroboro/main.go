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
	fs := flag.NewFlagSet("capture", flag.ExitOnError)
	project := fs.String("project", "", "Project name")
	tags := fs.String("tags", "", "Comma-separated tags")
	dbPath := fs.String("db", "", "SQLite database path (uses file storage if not specified)")

	fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Error: No content provided\n")
		fmt.Fprintf(os.Stderr, "Usage: uroboro capture \"your insight here\"\n")
		os.Exit(1)
	}

	content := fs.Arg(0)

	var service *capture.CaptureService
	var err error

	if *dbPath != "" {
		service, err = capture.NewCaptureServiceWithDB(*dbPath)
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
	fs := flag.NewFlagSet("publish", flag.ExitOnError)
	days := fs.Int("days", 1, "Number of days to look back")
	blog := fs.Bool("blog", false, "Generate blog post")
	devlog := fs.Bool("devlog", false, "Generate devlog")
	title := fs.String("title", "", "Blog post title")
	preview := fs.Bool("preview", false, "Preview content without saving")
	format := fs.String("format", "markdown", "Output format: markdown, html, text")
	dbPath := fs.String("db", "", "SQLite database path (reads from files if not specified)")

	fs.Parse(args)

	var service *publish.PublishService
	var err error

	if *dbPath != "" {
		service, err = publish.NewPublishServiceWithDB(*dbPath)
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
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	days := fs.String("days", "7", "Number of days to show")

	fs.Parse(args)

	daysInt, err := strconv.Atoi(*days)
	if err != nil {
		daysInt = 7
	}

	service := status.NewStatusService()
	if err := service.ShowStatus(daysInt); err != nil {
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
