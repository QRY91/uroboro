package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/analytics"
	"github.com/QRY91/uroboro/internal/capture"
	"github.com/QRY91/uroboro/internal/config"
	"github.com/QRY91/uroboro/internal/context"
	"github.com/QRY91/uroboro/internal/database"
	"github.com/QRY91/uroboro/internal/feast"
	"github.com/QRY91/uroboro/internal/journey"
	"github.com/QRY91/uroboro/internal/publish"
	"github.com/QRY91/uroboro/internal/ripcord"
	"github.com/QRY91/uroboro/internal/status"
	"github.com/QRY91/uroboro/internal/tagging"
)

func main() {
	// Initialize PostHog analytics
	analytics.Initialize()

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
	case "feast", "-f":
		handleFeast(os.Args[2:])
	case "analytics", "-a":
		handleAnalytics(os.Args[2:])
	case "session":
		handleSession(os.Args[2:])
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
	journey := fs.Bool("journey", false, "Generate journey replay visualization")
	title := fs.String("title", "", "Blog post title")
	preview := fs.Bool("preview", false, "Preview content without saving")
	format := fs.String("format", "markdown", "Output format: markdown, html, text")
	project := fs.String("project", "", "Project name")
	ripcordFlag := fs.Bool("ripcord", false, "Copy published content to clipboard")
	dbFlag := fs.String("db", "", "Database path (optional)")

	// Journey-specific flags
	port := fs.Int("port", 8080, "Port for journey web server")
	autoOpen := fs.Bool("open", true, "Automatically open browser for journey")
	theme := fs.String("theme", "default", "Theme for journey visualization")
	exportJSON := fs.Bool("export", false, "Export journey data to JSON file")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error parsing publish flags: %v\n", err)
		os.Exit(1)
	}

	if !*blog && !*devlog && !*journey {
		fmt.Fprintf(os.Stderr, "❌ Specify --blog, --devlog, or --journey\n")
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
	} else if *journey {
		err = handleJourneyPublish(*days, *project, *port, *autoOpen, *theme, *exportJSON, *title, service)
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
	archivedFlag := fs.Bool("archived", false, "Show archived captures instead of active ones")
	allFlag := fs.Bool("all", false, "Show both active and archived capture counts")
	archiveStatsFlag := fs.Bool("archive-stats", false, "Show archive statistics")

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

	// Run auto-feast check if using database (only if not showing archived data)
	if dbPath != "" && !*archivedFlag && !*archiveStatsFlag {
		db, err := database.NewDB(dbPath)
		if err == nil {
			defer db.Close()
			feastEngine := feast.NewFeastEngine(db, feast.DefaultFeastConfig())
			// Run auto-feast silently before showing status
			feastEngine.AutoFeastCheck()
		}
	}

	// Handle archive-specific commands
	if *archiveStatsFlag {
		showArchiveStats(dbPath)
		return
	}

	if *archivedFlag {
		showArchivedCaptures(*days, dbPath, *project)
		return
	}

	if *allFlag {
		showAllCapturesSummary(*days, dbPath, *project)
		return
	}

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

		analyticsConfig, err := config.LoadAnalyticsConfig()
		if err != nil {
			fmt.Printf("❌ Failed to load analytics config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("🔧 uroboro configuration:")
		fmt.Println()

		if dbPath == "" {
			fmt.Println("📁 Database: No default configured (using file storage)")
		} else {
			fmt.Printf("🗄️  Database: %s\n", dbPath)
		}

		fmt.Printf("📊 Analytics: %s\n", func() string {
			if analyticsConfig.AnalyticsEnabled {
				return "✅ Enabled"
			}
			return "❌ Disabled"
		}())

		if analyticsConfig.AnalyticsEnabled {
			fmt.Printf("   PostHog Host: %s\n", analyticsConfig.PostHogHost)
			fmt.Printf("   Privacy Mode: %s\n", analyticsConfig.PrivacyMode)
			if analyticsConfig.PostHogAPIKey != "" {
				fmt.Printf("   API Key: %s...%s\n",
					analyticsConfig.PostHogAPIKey[:4],
					analyticsConfig.PostHogAPIKey[len(analyticsConfig.PostHogAPIKey)-4:])
			} else {
				fmt.Println("   API Key: Not set")
			}
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

	case "analytics-on":
		// Load current config to preserve settings
		analyticsConfig, err := config.LoadAnalyticsConfig()
		if err != nil {
			analyticsConfig = &config.Config{
				PostHogHost: "https://eu.posthog.com",
				PrivacyMode: "enhanced",
			}
		}

		err = config.SaveAnalyticsConfig(true, analyticsConfig.PostHogAPIKey,
			analyticsConfig.PostHogHost, analyticsConfig.PrivacyMode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to enable analytics: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Analytics enabled")
		if analyticsConfig.PostHogAPIKey == "" {
			fmt.Println("💡 Set PostHog API key with: uroboro config set-posthog-key <key>")
		}

	case "analytics-off":
		// Load current config to preserve other settings
		analyticsConfig, err := config.LoadAnalyticsConfig()
		if err != nil {
			analyticsConfig = &config.Config{
				PostHogHost: "https://eu.posthog.com",
				PrivacyMode: "enhanced",
			}
		}

		err = config.SaveAnalyticsConfig(false, analyticsConfig.PostHogAPIKey,
			analyticsConfig.PostHogHost, analyticsConfig.PrivacyMode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to disable analytics: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Analytics disabled")

	case "set-posthog-key":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "❌ Usage: uroboro config set-posthog-key <api-key>\n")
			os.Exit(1)
		}
		apiKey := args[1]

		// Load current config to preserve other settings
		analyticsConfig, err := config.LoadAnalyticsConfig()
		if err != nil {
			analyticsConfig = &config.Config{
				PostHogHost: "https://eu.posthog.com",
				PrivacyMode: "enhanced",
			}
		}

		err = config.SaveAnalyticsConfig(analyticsConfig.AnalyticsEnabled, apiKey,
			analyticsConfig.PostHogHost, analyticsConfig.PrivacyMode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to save PostHog API key: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ PostHog API key set to: %s...%s\n", apiKey[:4], apiKey[len(apiKey)-4:])
		if !analyticsConfig.AnalyticsEnabled {
			fmt.Println("💡 Enable analytics with: uroboro config analytics-on")
		}

	case "set-posthog-host":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "❌ Usage: uroboro config set-posthog-host <host-url>\n")
			os.Exit(1)
		}
		host := args[1]

		// Load current config to preserve other settings
		analyticsConfig, err := config.LoadAnalyticsConfig()
		if err != nil {
			analyticsConfig = &config.Config{
				PrivacyMode: "enhanced",
			}
		}

		err = config.SaveAnalyticsConfig(analyticsConfig.AnalyticsEnabled, analyticsConfig.PostHogAPIKey,
			host, analyticsConfig.PrivacyMode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to save PostHog host: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ PostHog host set to: %s\n", host)

	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown config command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Available commands:\n")
		fmt.Fprintf(os.Stderr, "  set-db <path>           Set default database path\n")
		fmt.Fprintf(os.Stderr, "  analytics-on            Enable analytics\n")
		fmt.Fprintf(os.Stderr, "  analytics-off           Disable analytics\n")
		fmt.Fprintf(os.Stderr, "  set-posthog-key <key>   Set PostHog API key\n")
		fmt.Fprintf(os.Stderr, "  set-posthog-host <url>  Set PostHog host URL\n")
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
	fmt.Println("    --archived                          # Show archived captures")
	fmt.Println("    --all                              # Show both active and archived counts")
	fmt.Println("    --archive-stats                    # Show feast and archive statistics")
	fmt.Println("  uroboro feast [flags]                 # Archive old captures (ouroboros)")
	fmt.Println("  uroboro analytics [flags]             # Show personal development analytics")
	fmt.Println("  uroboro session [command]             # Manage development sessions")
	fmt.Println("  uroboro config [command]              # Configure uroboro")
	fmt.Println()
	fmt.Println("Short aliases:")
	fmt.Println("  uro -c \"content\"    # capture")
	fmt.Println("  uro -p --devlog      # publish devlog")
	fmt.Println("  uro -p --journey     # journey replay visualization")
	fmt.Println("  uro -s              # status")
	fmt.Println("  uro -s --archived   # browse archived captures")
	fmt.Println("  uro -s --all        # active + archived summary")
	fmt.Println("  uro -f              # feast (archive old captures)")
	fmt.Println("  uro -a              # analytics")
	fmt.Println("  uro session info    # current session")
	fmt.Println()
	fmt.Println("Publish Options:")
	fmt.Println("  --blog               # Generate blog post")
	fmt.Println("  --devlog             # Generate development log")
	fmt.Println("  --journey            # Interactive timeline visualization")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  🧠 Smart project detection")
	fmt.Println("  🏷️  Auto-tagging")
	fmt.Println("  📋 Ripcord functionality")
	fmt.Println("  🗄️  Optional database storage")
	fmt.Println("  🐍 Auto-feast (archive old captures)")
	fmt.Println("  📚 Archive browsing and statistics")
	fmt.Println("  📁 File-based fallback")
	fmt.Println("  🎬 Journey replay visualization")
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

func handleAnalytics(args []string) {
	fs := flag.NewFlagSet("analytics", flag.ExitOnError)
	days := fs.Int("days", 7, "Number of days to analyze")
	sessions := fs.Bool("sessions", false, "Show session analytics")
	productivity := fs.Bool("productivity", false, "Show productivity trends")
	projects := fs.Bool("projects", false, "Show project breakdown")
	all := fs.Bool("all", false, "Show all analytics")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error parsing analytics flags: %v\n", err)
		os.Exit(1)
	}

	// If no specific flags, show summary
	if !*sessions && !*productivity && !*projects {
		*all = true
	}

	fmt.Printf("📊 Personal Development Analytics (last %d days)\n", *days)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Get current session info
	currentSession := analytics.Get().GetCurrentSession()
	if currentSession != nil {
		fmt.Printf("🎯 Current Session: %s\n", currentSession.ID)
		fmt.Printf("   Duration: %v\n", currentSession.LastActivity.Sub(currentSession.StartTime).Round(time.Minute))
		fmt.Printf("   Activities: %d\n", len(currentSession.Activities))
		fmt.Printf("   Project: %s\n", currentSession.Project)
		fmt.Println()
	}

	if *all || *sessions {
		fmt.Println("📈 Session Insights:")
		fmt.Println("   • Session-based analytics help you understand your development flow")
		fmt.Println("   • Track productivity patterns and optimize your work sessions")
		fmt.Println("   • Identify peak performance times and context switching patterns")
		fmt.Println()
	}

	if *all || *productivity {
		fmt.Println("🚀 Productivity Patterns:")
		fmt.Println("   • Self-tracking analytics provide personal insights")
		fmt.Println("   • No surveillance - just personal optimization data")
		fmt.Println("   • Privacy-first approach respects your development flow")
		fmt.Println()
	}

	if *all || *projects {
		fmt.Println("🎨 Project Analytics:")
		fmt.Println("   • Track time spent across different projects")
		fmt.Println("   • Understand your project switching patterns")
		fmt.Println("   • Optimize context management for better flow states")
		fmt.Println()
	}

	fmt.Println("💡 Tip: Analytics are completely optional and privacy-first")
	fmt.Printf("   Current status: %s\n", func() string {
		if analytics.Get().IsEnabled() {
			return "✅ Enabled - tracking your development patterns"
		}
		return "⚠️  Disabled - set POSTHOG_API_KEY to enable personal insights"
	}())

	// Track analytics dashboard usage
	if analytics.Get().IsEnabled() {
		fmt.Println()
		fmt.Println("🔗 PostHog Dashboard: Check your personal development insights!")
	}
}

func handleSession(args []string) {
	if len(args) == 0 {
		// Show current session info
		showCurrentSession()
		return
	}

	command := args[0]
	switch command {
	case "info", "current":
		showCurrentSession()
	case "end", "stop":
		endCurrentSession()
	case "new", "start":
		startNewSession(args[1:])
	case "timeout":
		checkSessionTimeout()
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown session command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Available commands: info, end, new, timeout\n")
		os.Exit(1)
	}
}

func showCurrentSession() {
	currentSession := analytics.Get().GetCurrentSession()
	if currentSession == nil {
		fmt.Println("⚠️  No active development session")
		fmt.Println("   Start a new session with: uroboro session new")
		return
	}

	duration := time.Since(currentSession.StartTime)
	lastActivity := time.Since(currentSession.LastActivity)

	fmt.Printf("🎯 Current Development Session\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Session ID: %s\n", currentSession.ID)
	fmt.Printf("Duration: %v\n", duration.Round(time.Minute))
	fmt.Printf("Last Activity: %v ago\n", lastActivity.Round(time.Second))
	fmt.Printf("Project: %s\n", func() string {
		if currentSession.Project == "" {
			return "(not set)"
		}
		return currentSession.Project
	}())
	fmt.Printf("Activities: %d\n", len(currentSession.Activities))

	if currentSession.GitContext != nil {
		fmt.Printf("Git Branch: %s\n", currentSession.GitContext.Branch)
		if currentSession.GitContext.IsDirty {
			fmt.Printf("Git Status: 🔄 dirty\n")
		} else {
			fmt.Printf("Git Status: ✅ clean\n")
		}
	}

	fmt.Printf("Working Dir: %s\n", currentSession.WorkingDir)

	if len(currentSession.Activities) > 0 {
		fmt.Println("\n📋 Recent Activities:")
		for i := len(currentSession.Activities) - 1; i >= 0 && i >= len(currentSession.Activities)-3; i-- {
			activity := currentSession.Activities[i]
			timeAgo := time.Since(activity.Timestamp).Round(time.Minute)
			fmt.Printf("   • %s (%v ago)\n", activity.Type, timeAgo)
		}
	}

	fmt.Println("\n💡 Commands:")
	fmt.Println("   uroboro session end     # End current session")
	fmt.Println("   uroboro session new     # Start fresh session")
}

func endCurrentSession() {
	currentSession := analytics.Get().GetCurrentSession()
	if currentSession == nil {
		fmt.Println("⚠️  No active session to end")
		return
	}

	fmt.Printf("🔚 Ending session: %s\n", currentSession.ID)
	analytics.Get().EndCurrentSession()
	fmt.Println("✅ Session ended successfully")

	if analytics.Get().IsEnabled() {
		fmt.Println("📊 Session analytics sent to PostHog")
	}
}

func startNewSession(args []string) {
	// End current session first
	if analytics.Get().GetCurrentSession() != nil {
		fmt.Println("🔄 Ending current session...")
		analytics.Get().EndCurrentSession()
	}

	// Parse project name if provided
	project := ""
	if len(args) > 0 {
		project = args[0]
	}

	// Start a new session by triggering an activity
	fmt.Printf("🎯 Starting new development session")
	if project != "" {
		fmt.Printf(" (project: %s)", project)
	}
	fmt.Println()

	// Track the session start
	analytics.Get().TrackCaptureSimple("Session started manually", project, []string{"session", "manual"})

	// Show the new session info
	time.Sleep(100 * time.Millisecond) // Brief delay to ensure session is created
	showCurrentSession()
}

func checkSessionTimeout() {
	fmt.Println("🕐 Checking session timeout...")
	analytics.Get().CheckSessionTimeout()

	currentSession := analytics.Get().GetCurrentSession()
	if currentSession == nil {
		fmt.Println("⏰ Session has timed out (no active session)")
	} else {
		lastActivity := time.Since(currentSession.LastActivity)
		fmt.Printf("✅ Session still active (last activity: %v ago)\n", lastActivity.Round(time.Second))
	}
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

func handleJourneyPublish(days int, project string, port int, autoOpen bool, theme string, exportJSON bool, title string, service *publish.PublishService) error {
	// Get database from service
	if service == nil {
		return fmt.Errorf("service not initialized")
	}

	// For now, we need direct database access - this could be refactored later to go through the publish service
	dbPath, err := getDefaultDBPath()
	if err != nil {
		return fmt.Errorf("database not available: %w", err)
	}

	db, err := database.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Parse projects if specified
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
		Port:     port,
		AutoOpen: autoOpen,
		Theme:    theme,
		Title:    title,
		Export:   exportJSON,
	}

	if exportJSON {
		// Export mode - generate JSON and save to file
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

	// Web server mode
	server := journey.NewServer(db, port)

	fmt.Printf("🎬 Starting Journey Replay visualization...\n")
	fmt.Printf("📊 Analyzing %d days of data\n", days)
	if len(projectList) > 0 {
		fmt.Printf("🎯 Filtering projects: %s\n", strings.Join(projectList, ", "))
	}
	fmt.Printf("🎨 Theme: %s\n", theme)

	if autoOpen {
		go func() {
			time.Sleep(2 * time.Second)
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
	var cmd string
	var args []string

	switch {
	case fileExists("/usr/bin/xdg-open"):
		cmd = "xdg-open"
	case fileExists("/usr/bin/open"):
		cmd = "open"
	case fileExists("/usr/bin/start"):
		cmd = "cmd"
		args = []string{"/c", "start"}
	default:
		fmt.Printf("🌐 Open your browser and navigate to: %s\n", url)
		return
	}

	args = append(args, url)
	exec.Command(cmd, args...).Start()
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// handleFeast manages the feast command for archiving old captures
func handleFeast(args []string) {
	var days int
	var silent bool
	var auto bool

	// Parse flags
	feastFlags := flag.NewFlagSet("feast", flag.ExitOnError)
	feastFlags.IntVar(&days, "days", 30, "Archive captures older than N days")
	feastFlags.BoolVar(&silent, "silent", false, "Archive without showing digest")
	feastFlags.BoolVar(&auto, "auto", false, "Run auto-feast (used internally)")

	err := feastFlags.Parse(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing feast flags: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	dbPath, err := getDefaultDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get default database path: %v\n", err)
		os.Exit(1)
	}
	db, err := database.NewDB(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create feast engine
	config := feast.DefaultFeastConfig()
	if silent {
		config.SilentMode = true
		config.ShowDigest = false
	}

	feastEngine := feast.NewFeastEngine(db, config)

	// Perform feast operation
	if auto {
		err = feastEngine.AutoFeastCheck()
	} else {
		err = feastEngine.ManualFeast(days, silent)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Feast operation failed: %v\n", err)
		os.Exit(1)
	}
}

// showArchivedCaptures displays archived captures
func showArchivedCaptures(days int, dbPath string, project string) {
	if dbPath == "" {
		fmt.Printf("❌ Archive browsing requires database storage\n")
		fmt.Printf("   Configure database path with: uro config set db_path <path>\n")
		return
	}

	db, err := database.NewDB(dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("🐍 uroboro status --archived")
	fmt.Printf("🗄️  Database: %s\n\n", dbPath)

	// Query archived captures
	query := `
		SELECT original_id, timestamp, content, project, archived_at, archive_reason
		FROM archived_captures
		WHERE archived_at >= datetime('now', '-' || ? || ' days')
	`
	args := []interface{}{days}

	if project != "" {
		query += " AND project = ?"
		args = append(args, project)
	}

	query += " ORDER BY archived_at DESC LIMIT 20"

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Printf("❌ Failed to query archived captures: %v\n", err)
		return
	}
	defer rows.Close()

	var count int
	fmt.Printf("📚 Archived Captures (last %d days):\n", days)

	for rows.Next() {
		var originalID int64
		var timestamp, content, archivedAt, archiveReason string
		var projectName sql.NullString

		err := rows.Scan(&originalID, &timestamp, &content, &projectName, &archivedAt, &archiveReason)
		if err != nil {
			continue
		}

		count++

		// Truncate content if too long
		if len(content) > 80 {
			content = content[:80] + "..."
		}

		project := "no project"
		if projectName.Valid && projectName.String != "" {
			project = projectName.String
		}

		// Parse archived date for display
		archivedTime, _ := time.Parse("2006-01-02 15:04:05", archivedAt)
		archivedAgo := time.Since(archivedTime)

		var timeAgo string
		if archivedAgo.Hours() < 24 {
			timeAgo = fmt.Sprintf("%.0fh ago", archivedAgo.Hours())
		} else {
			timeAgo = fmt.Sprintf("%.0fd ago", archivedAgo.Hours()/24)
		}

		fmt.Printf("  📦 [%s] %s (archived %s)\n", project, content, timeAgo)
	}

	if count == 0 {
		fmt.Println("  No archived captures found")
	} else {
		fmt.Printf("\nTotal archived captures shown: %d\n", count)
	}
}

// showAllCapturesSummary shows summary of both active and archived captures
func showAllCapturesSummary(days int, dbPath string, project string) {
	if dbPath == "" {
		fmt.Printf("❌ Archive summary requires database storage\n")
		return
	}

	db, err := database.NewDB(dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("🐍 uroboro status --all")
	fmt.Printf("🗄️  Database: %s\n\n", dbPath)

	// Count active captures
	activeQuery := `
		SELECT COUNT(*) FROM captures
		WHERE timestamp >= datetime('now', '-' || ? || ' days')
	`
	activeArgs := []interface{}{days}
	if project != "" {
		activeQuery += " AND project = ?"
		activeArgs = append(activeArgs, project)
	}

	var activeCount int
	err = db.QueryRow(activeQuery, activeArgs...).Scan(&activeCount)
	if err != nil {
		activeCount = 0
	}

	// Count archived captures
	archivedQuery := `
		SELECT COUNT(*) FROM archived_captures
		WHERE archived_at >= datetime('now', '-' || ? || ' days')
	`
	archivedArgs := []interface{}{days}
	if project != "" {
		archivedQuery += " AND project = ?"
		archivedArgs = append(archivedArgs, project)
	}

	var archivedCount int
	err = db.QueryRow(archivedQuery, archivedArgs...).Scan(&archivedCount)
	if err != nil {
		archivedCount = 0
	}

	fmt.Printf("📊 Capture Summary (last %d days):\n", days)
	fmt.Printf("  🟢 Active captures: %d\n", activeCount)
	fmt.Printf("  📦 Archived captures: %d\n", archivedCount)
	fmt.Printf("  📈 Total captures: %d\n\n", activeCount+archivedCount)

	if activeCount > 0 {
		fmt.Printf("Recent active captures:\n")
		// Show recent active captures
		service := status.NewStatusService()
		service.ShowStatus(days, dbPath, project)
	}
}

// showArchiveStats displays feast and archive statistics
func showArchiveStats(dbPath string) {
	if dbPath == "" {
		fmt.Printf("❌ Archive statistics require database storage\n")
		return
	}

	db, err := database.NewDB(dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("🐍 uroboro status --archive-stats")
	fmt.Printf("🗄️  Database: %s\n\n", dbPath)

	// Total archived items
	var totalArchived int
	err = db.QueryRow("SELECT COUNT(*) FROM archived_captures").Scan(&totalArchived)
	if err != nil {
		totalArchived = 0
	}

	// Recent archive activity (last 30 days)
	var recentArchived int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM archived_captures
		WHERE archived_at >= datetime('now', '-30 days')
	`).Scan(&recentArchived)
	if err != nil {
		recentArchived = 0
	}

	// Archive by reason
	autoFeastQuery := `
		SELECT COUNT(*) FROM archived_captures
		WHERE archive_reason = 'auto_feast' AND archived_at >= datetime('now', '-30 days')
	`
	var autoFeastCount int
	err = db.QueryRow(autoFeastQuery).Scan(&autoFeastCount)
	if err != nil {
		autoFeastCount = 0
	}

	manualFeastQuery := `
		SELECT COUNT(*) FROM archived_captures
		WHERE archive_reason = 'manual_feast' AND archived_at >= datetime('now', '-30 days')
	`
	var manualFeastCount int
	err = db.QueryRow(manualFeastQuery).Scan(&manualFeastCount)
	if err != nil {
		manualFeastCount = 0
	}

	// Top archived projects (last 30 days)
	projectQuery := `
		SELECT project, COUNT(*) as count FROM archived_captures
		WHERE archived_at >= datetime('now', '-30 days') AND project IS NOT NULL
		GROUP BY project ORDER BY count DESC LIMIT 5
	`

	fmt.Printf("📊 Archive Statistics:\n")
	fmt.Printf("  📦 Total archived: %d items\n", totalArchived)
	fmt.Printf("  📅 Last 30 days: %d items\n", recentArchived)
	fmt.Printf("  🤖 Auto-feast: %d items\n", autoFeastCount)
	fmt.Printf("  👤 Manual feast: %d items\n\n", manualFeastCount)

	rows, err := db.Query(projectQuery)
	if err == nil {
		defer rows.Close()

		fmt.Printf("🏷️  Top Archived Projects (last 30 days):\n")
		hasProjects := false
		for rows.Next() {
			var project string
			var count int
			if rows.Scan(&project, &count) == nil {
				fmt.Printf("  📁 %s: %d items\n", project, count)
				hasProjects = true
			}
		}
		if !hasProjects {
			fmt.Printf("  No project data available\n")
		}
	}
}
