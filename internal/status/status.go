package status

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/common"
	"github.com/QRY91/uroboro/internal/database"
)

type Insight struct {
	Text      string    `json:"text"`
	Tags      []string  `json:"tags"`
	Timestamp time.Time `json:"timestamp"`
}

type StatusService struct{}

// NewStatusService creates a new instance of StatusService for displaying
// development activity status and recent captures.
func NewStatusService() *StatusService {
	return &StatusService{}
}

// ShowStatus displays recent development activity for the specified number of days.
// If dbPath is provided, reads from database; otherwise uses file storage.
// If project is specified, filters results to only show that project's activity.
func (s *StatusService) ShowStatus(days int, dbPath string, project string) error {
	fmt.Println("🐍 uroboro status")

	// If database path is provided, read from database
	if dbPath != "" {
		return s.showStatusFromDatabase(days, dbPath, project)
	}

	// Otherwise, read from file storage
	return s.showStatusFromFiles(days)
}

// ShowStatusWithTags displays recent development activity filtered by tags.
func (s *StatusService) ShowStatusWithTags(days int, dbPath string, project string, tags string) error {
	fmt.Println("🐍 uroboro status")

	// If database path is provided, read from database
	if dbPath != "" {
		return s.showStatusFromDatabaseWithTags(days, dbPath, project, tags)
	}

	// Otherwise, read from file storage
	return s.showStatusFromFilesWithTags(days, project, tags)
}

// showStatusFromDatabase retrieves and displays captures from the SQLite database.
// Filters by project if specified, orders by timestamp descending (newest first).
func (s *StatusService) showStatusFromDatabase(days int, dbPath string, project string) error {
	// Import database package inline to avoid import issues
	db, err := database.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get recent captures (empty project string means all projects)
	captures, err := db.GetRecentCaptures(days, project)
	if err != nil {
		return fmt.Errorf("failed to query captures: %w", err)
	}

	fmt.Printf("Recent activity (%d days): %d items\n", days, len(captures))
	fmt.Printf("\n📝 Recent Captures (last %d days):\n", days)

	if len(captures) == 0 {
		fmt.Println("  No recent captures found")
		return nil
	}

	// Show up to 10 most recent captures
	shown := 0
	for i := 0; i < len(captures) && shown < 10; i++ {
		capture := captures[i]

		// Truncate content if too long
		content := capture.Content
		if len(content) > 80 {
			content = content[:80] + "..."
		}

		// Format with project if available
		if capture.Project.Valid && capture.Project.String != "" {
			fmt.Printf("  📄 [%s] %s\n", capture.Project.String, content)
		} else {
			fmt.Printf("  📄 %s\n", content)
		}
		shown++
	}

	return nil
}

// showStatusFromFiles retrieves and displays captures from markdown files.
// Fallback method when database is not available or configured.
func (s *StatusService) showStatusFromFiles(days int) error {
	// Get cross-platform data directory
	dataDir := common.GetDataDir()

	// Count recent activity
	cutoff := time.Now().AddDate(0, 0, -days)
	activityCount := 0

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		// Data directory doesn't exist yet
		fmt.Printf("Recent activity (%d days): 0 items\n", days)
		fmt.Printf("\n📝 Recent Captures (last %d days):\n", days)
		fmt.Println("  No recent captures found")
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(cutoff) {
			activityCount++
		}
	}

	fmt.Printf("Recent activity (%d days): %d items\n", days, activityCount)

	// Show recent captures
	fmt.Printf("\n📝 Recent Captures (last %d days):\n", days)

	shown := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(cutoff) {
			fullPath := filepath.Join(dataDir, entry.Name())
			content, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}

			captures := s.extractRecentCaptures(string(content))
			for _, capture := range captures {
				if shown >= 10 {
					break
				}
				fmt.Printf("  📄 %s\n", capture)
				shown++
			}
		}

		if shown >= 10 {
			break
		}
	}

	if shown == 0 {
		fmt.Println("  No recent captures found")
	}

	return nil
}

// showStatusFromDatabaseWithTags retrieves and displays captures from database filtered by tags.
func (s *StatusService) showStatusFromDatabaseWithTags(days int, dbPath string, project string, tags string) error {
	db, err := database.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get recent captures
	captures, err := db.GetRecentCaptures(days, project)
	if err != nil {
		return fmt.Errorf("failed to query captures: %w", err)
	}

	// Filter by tags if specified
	if tags != "" {
		captures = s.filterCapturesByTags(captures, tags)
	}

	fmt.Printf("Recent activity (%d days): %d items\n", days, len(captures))
	fmt.Printf("\n📝 Recent Captures (last %d days):\n", days)

	if len(captures) == 0 {
		fmt.Println("  No recent captures found")
		return nil
	}

	// Show up to 10 most recent captures
	shown := 0
	for i := 0; i < len(captures) && shown < 10; i++ {
		capture := captures[i]

		// Truncate content if too long
		content := capture.Content
		if len(content) > 80 {
			content = content[:80] + "..."
		}

		// Format with project if available
		if capture.Project.Valid && capture.Project.String != "" {
			fmt.Printf("  📄 [%s] %s\n", capture.Project.String, content)
		} else {
			fmt.Printf("  📄 %s\n", content)
		}
		shown++
	}

	return nil
}

// showStatusFromFilesWithTags retrieves and displays captures from files filtered by tags.
func (s *StatusService) showStatusFromFilesWithTags(days int, project string, tags string) error {
	// Get cross-platform data directory
	dataDir := common.GetDataDir()

	// Count recent activity
	cutoff := time.Now().AddDate(0, 0, -days)
	activityCount := 0

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		// Data directory doesn't exist yet
		fmt.Printf("Recent activity (%d days): 0 items\n", days)
		fmt.Printf("\n📝 Recent Captures (last %d days):\n", days)
		fmt.Println("  No recent captures found")
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(cutoff) {
			activityCount++
		}
	}

	fmt.Printf("Recent activity (%d days): %d items\n", days, activityCount)

	// Show recent captures
	fmt.Printf("\n📝 Recent Captures (last %d days):\n", days)

	shown := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(cutoff) {
			fullPath := filepath.Join(dataDir, entry.Name())
			content, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}

			captures := s.extractRecentCapturesWithTags(string(content), project, tags)
			for _, capture := range captures {
				if shown >= 10 {
					break
				}
				fmt.Printf("  📄 %s\n", capture)
				shown++
			}
		}

		if shown >= 10 {
			break
		}
	}

	if shown == 0 {
		fmt.Println("  No recent captures found")
	}

	return nil
}

// filterCapturesByTags filters database captures by tags
func (s *StatusService) filterCapturesByTags(captures []database.Capture, filterTags string) []database.Capture {
	if filterTags == "" {
		return captures
	}

	var filtered []database.Capture
	for _, capture := range captures {
		if capture.Tags.Valid && s.captureHasTags(capture.Tags.String, filterTags) {
			filtered = append(filtered, capture)
		}
	}
	return filtered
}

// captureHasTags checks if a capture's tags contain any of the filter tags
func (s *StatusService) captureHasTags(captureTags, filterTags string) bool {
	if captureTags == "" || filterTags == "" {
		return false
	}

	// Split both capture tags and filter tags
	captureTagList := strings.Split(strings.ToLower(captureTags), ",")
	filterTagList := strings.Split(strings.ToLower(filterTags), ",")

	// Check if any filter tag is present in capture tags
	for _, filterTag := range filterTagList {
		filterTag = strings.TrimSpace(filterTag)
		for _, captureTag := range captureTagList {
			captureTag = strings.TrimSpace(captureTag)
			if captureTag == filterTag {
				return true
			}
		}
	}
	return false
}

// extractRecentCapturesWithTags parses markdown content and filters by tags and project
func (s *StatusService) extractRecentCapturesWithTags(content string, project string, tags string) []string {
	var captures []string
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "## 2025-") {
			// Found a timestamp, extract the capture block
			captureLines := []string{}
			j := i + 1
			
			// Skip empty line after timestamp
			if j < len(lines) && strings.TrimSpace(lines[j]) == "" {
				j++
			}
			
			// Collect capture content until next header or end
			var captureProject, captureTags string
			for j < len(lines) {
				nextLine := lines[j]
				if strings.HasPrefix(nextLine, "## ") {
					break
				}
				
				// Extract metadata
				if strings.HasPrefix(nextLine, "Project: ") {
					captureProject = strings.TrimSpace(strings.TrimPrefix(nextLine, "Project: "))
				} else if strings.HasPrefix(nextLine, "Tags: ") {
					captureTags = strings.TrimSpace(strings.TrimPrefix(nextLine, "Tags: "))
				} else if strings.TrimSpace(nextLine) != "" && !strings.HasPrefix(nextLine, "Project:") && !strings.HasPrefix(nextLine, "Tags:") {
					captureLines = append(captureLines, nextLine)
				}
				j++
			}
			
			// Filter by project if specified
			if project != "" && captureProject != project {
				continue
			}
			
			// Filter by tags if specified
			if tags != "" && !s.captureHasTags(captureTags, tags) {
				continue
			}
			
			// Build capture text
			if len(captureLines) > 0 {
				capture := strings.TrimSpace(strings.Join(captureLines, " "))
				if len(capture) > 80 {
					capture = capture[:80] + "..."
				}
				captures = append(captures, capture)
			}
		}
	}

	return captures
}

// extractRecentCaptures parses markdown content to extract capture entries.
// Looks for timestamp headers and extracts the content below them.
func (s *StatusService) extractRecentCaptures(content string) []string {

	var captures []string
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "## 2025-") {
			// Found a timestamp, get the content
			if i+2 < len(lines) && strings.TrimSpace(lines[i+2]) != "" {
				capture := strings.TrimSpace(lines[i+2])
				if len(capture) > 80 {
					capture = capture[:80] + "..."
				}
				captures = append(captures, capture)
			}
		}
	}

	return captures
}

func Show(days int, recent bool) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home directory: %w", err)
	}

	insightsFile := filepath.Join(homeDir, ".uroboro", "insights.jsonl")

	file, err := os.Open(insightsFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No insights captured yet. Use 'uroboro capture' to get started!")
			return nil
		}
		return fmt.Errorf("could not open insights file: %w", err)
	}
	defer file.Close()

	var insights []Insight
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var insight Insight
		if err := json.Unmarshal(scanner.Bytes(), &insight); err != nil {
			continue // Skip malformed lines
		}
		insights = append(insights, insight)
	}

	if len(insights) == 0 {
		fmt.Println("No insights found.")
		return nil
	}

	// Filter by days if specified
	if days > 0 {
		cutoff := time.Now().AddDate(0, 0, -days)
		var filtered []Insight
		for _, insight := range insights {
			if insight.Timestamp.After(cutoff) {
				filtered = append(filtered, insight)
			}
		}
		insights = filtered
	}

	fmt.Printf("📊 Uroboro Status\n")
	fmt.Printf("Total insights: %d\n", len(insights))

	if days > 0 {
		fmt.Printf("Recent activity (%d days): %d items\n", days, len(insights))
	}

	if recent && len(insights) > 0 {
		fmt.Println("\nRecent insights:")
		for i := len(insights) - 1; i >= 0 && i >= len(insights)-5; i-- {
			insight := insights[i]
			fmt.Printf("  • %s (%s)\n", insight.Text, insight.Timestamp.Format("2006-01-02 15:04"))
		}
	}

	return nil
}
