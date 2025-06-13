package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	fmt.Println("  uroboro analytics [flags]             # Show personal development analytics")
	fmt.Println("  uroboro session [command]             # Manage development sessions")
	fmt.Println("  uroboro config [command]              # Configure uroboro")
	fmt.Println()
	fmt.Println("Short aliases:")
	fmt.Println("  uro -c \"content\"    # capture")
	fmt.Println("  uro -p --devlog      # publish devlog")
	fmt.Println("  uro -s              # status")
	fmt.Println("  uro -a              # analytics")
	fmt.Println("  uro session info    # current session")
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
