package main

import (
	"flag"
	"fmt"
	"os"
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

	switch os.Args[1] {
	case "capture":
		handleCapture(os.Args[2:])
	case "publish":
		handlePublish(os.Args[2:])
	case "status":
		handleStatus(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func handleCapture(args []string) {
	fs := flag.NewFlagSet("capture", flag.ExitOnError)
	project := fs.String("project", "", "Project name")
	tags := fs.String("tags", "", "Comma-separated tags")

	fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Error: No content provided\n")
		fmt.Fprintf(os.Stderr, "Usage: uroboro capture \"your insight here\"\n")
		os.Exit(1)
	}

	content := fs.Arg(0)

	service := capture.NewCaptureService()
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

	fs.Parse(args)

	service := publish.NewPublishService()

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
		fmt.Fprintf(os.Stderr, "Usage: uroboro publish --blog [--title \"Title\"] [--days N] [--format FORMAT]\n")
		fmt.Fprintf(os.Stderr, "       uroboro publish --devlog [--days N]\n")
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
	fmt.Println("uroboro - The Self-Documenting Content Pipeline")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  uroboro capture \"your insight here\" [--project name] [--tags tag1,tag2]")
	fmt.Println("  uroboro publish --blog [--title \"Title\"] [--days N] [--preview]")
	fmt.Println("  uroboro publish --devlog [--days N]")
	fmt.Println("  uroboro status [--days N]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  uroboro capture \"Fixed auth timeout - cut query time from 3s to 200ms\"")
	fmt.Println("  uroboro publish --blog --title \"This Week's Fixes\"")
	fmt.Println("  uroboro publish --devlog")
	fmt.Println("  uroboro status")
}
