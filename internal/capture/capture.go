package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/QRY91/uroboro/internal/common"
	"github.com/QRY91/uroboro/internal/database"
)

type CaptureService struct {
	db *database.DB
}

func NewCaptureService() *CaptureService {
	return &CaptureService{}
}

func NewCaptureServiceWithDB(dbPath string) (*CaptureService, error) {
	db, err := database.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &CaptureService{db: db}, nil
}

func (c *CaptureService) Capture(content, project, tags string) error {
	// If database is available, use it
	if c.db != nil {
		return c.captureToDatabase(content, project, tags)
	}

	// Otherwise, fall back to file storage
	return c.captureToFile(content, project, tags)
}

func (c *CaptureService) captureToDatabase(content, project, tags string) error {
	capture, err := c.db.InsertCapture(content, project, tags)
	if err != nil {
		return fmt.Errorf("failed to capture to database: %w", err)
	}

	fmt.Printf("✅ Captured [ID:%d]: %s\n", capture.ID, truncateContent(content, 60))
	if capture.Project.Valid && capture.Project.String != "" {
		fmt.Printf("   Project: %s\n", capture.Project.String)
	}
	return nil
}

func (c *CaptureService) captureToFile(content, project, tags string) error {
	timestamp := time.Now().Format("2006-01-02T15:04:05")

	// Get cross-platform data directory
	dataDir := common.GetDataDir()
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create daily note filename
	today := time.Now().Format("2006-01-02")
	filename := filepath.Join(dataDir, fmt.Sprintf("%s.md", today))

	// Prepare entry
	entry := fmt.Sprintf("\n## %s\n\n%s\n", timestamp, content)

	if project != "" {
		entry += fmt.Sprintf("Project: %s\n", project)
	}

	if tags != "" {
		entry += fmt.Sprintf("Tags: %s\n", tags)
	}

	// Append to daily notes file
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open daily notes file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write to daily notes: %w", err)
	}

	fmt.Printf("✅ Captured: %s\n", truncateContent(content, 60))
	return nil
}

func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}

// ProcessToolMessages processes unprocessed tool messages for cross-tool integration
func (c *CaptureService) ProcessToolMessages() error {
	if c.db == nil {
		return fmt.Errorf("database not available for tool message processing")
	}
	
	return c.db.ProcessUroboroCaptures()
}
