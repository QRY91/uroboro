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

func NewStatusService() *StatusService {
	return &StatusService{}
}

func (s *StatusService) ShowStatus(days int, dbPath string) error {
	fmt.Println("🐍 uroboro status")

	// If database path is provided, read from database
	if dbPath != "" {
		return s.showStatusFromDatabase(days, dbPath)
	}

	// Otherwise, read from file storage
	return s.showStatusFromFiles(days)
}

func (s *StatusService) showStatusFromDatabase(days int, dbPath string) error {
	// Import database package inline to avoid import issues
	db, err := database.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Get recent captures (empty project string means all projects)
	captures, err := db.GetRecentCaptures(days, "")
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
	for i := len(captures) - 1; i >= 0 && shown < 10; i-- {
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
