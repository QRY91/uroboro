package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/QRY91/uroboro/internal/common"
	"github.com/QRY91/uroboro/internal/context"
	"github.com/QRY91/uroboro/internal/database"
	"github.com/QRY91/uroboro/internal/tagging"
)

// localHostname returns the machine name for capture attribution.
// Falls back to empty string if not determinable.
func localHostname() string {
	if h := os.Getenv("UROBORO_MACHINE"); h != "" {
		return h
	}
	h, _ := os.Hostname()
	return h
}

type Service struct {
	db *database.DB
}

func NewService(dbPath string) (*Service, error) {
	db, err := database.NewDB(dbPath)
	if err != nil {
		return nil, err
	}
	return &Service{db: db}, nil
}

func (s *Service) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Service) Capture(content, project, tags string, timestamp *time.Time) error {
	detector := context.NewProjectDetector()

	// Auto-detect project if not provided
	if project == "" {
		if detected := detector.DetectProject(); detected != "" {
			project = detected
			fmt.Printf("🔍 Auto-detected project: %s\n", project)
		}
	}

	// Auto-detect branch and machine
	branch := detector.DetectBranch()
	machine := localHostname()

	// Auto-enhance tags
	analyzer := tagging.NewTagAnalyzer()
	originalTags := tags
	tags = analyzer.EnhanceTags(content, tags)
	if tags != originalTags {
		if suggested := analyzer.GetSuggestedTags(content); suggested != "" {
			fmt.Printf("🏷️  Auto-detected tags: %s\n", suggested)
		}
	}

	// Store to database if available, otherwise file
	if s.db != nil {
		capture, err := s.db.InsertCapture(content, project, tags, branch, machine, timestamp)
		if err != nil {
			return err
		}
		fmt.Printf("✅ Captured [ID:%d]: %s\n", capture.ID, truncate(content, 60))
		if capture.Project != "" {
			fmt.Printf("   Project: %s\n", capture.Project)
		}
		if timestamp != nil {
			fmt.Printf("   Time: %s\n", timestamp.Format("2006-01-02 15:04"))
		}
		return nil
	}

	return s.captureToFile(content, project, tags, timestamp)
}

func (s *Service) captureToFile(content, project, tags string, timestamp *time.Time) error {
	dataDir := common.GetDataDir()
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	ts := time.Now()
	if timestamp != nil {
		ts = *timestamp
	}
	filename := filepath.Join(dataDir, ts.Format("2006-01-02")+".md")
	entry := fmt.Sprintf("\n## %s\n\n%s\n", ts.Format("2006-01-02T15:04:05"), content)
	if project != "" {
		entry += fmt.Sprintf("Project: %s\n", project)
	}
	if tags != "" {
		entry += fmt.Sprintf("Tags: %s\n", tags)
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return err
	}

	fmt.Printf("✅ Captured: %s\n", truncate(content, 60))
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
