package capture

import (
	"fmt"
	"os"
	"time"
)

type CaptureService struct{}

func NewCaptureService() *CaptureService {
	return &CaptureService{}
}

func (c *CaptureService) Capture(content, project, tags string) error {
	timestamp := time.Now().Format("2006-01-02T15:04:05")

	// Create daily note filename
	today := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("daily-notes-%s.md", today)

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
