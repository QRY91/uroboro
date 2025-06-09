package publish

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPublishService_CollectRecentActivity(t *testing.T) {
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	service := NewPublishService()

	// Create test data directory
	dataDir := filepath.Join(tempHome, ".local", "share", "uroboro", "daily")
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test file with sample content
	today := time.Now().Format("2006-01-02")
	testFile := filepath.Join(dataDir, today+".md")
	testContent := `
## 2025-06-04T10:00:00

Test capture for publish

## 2025-06-04T11:00:00

Another capture for testing
Tags: test,publish
`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test collecting activity
	activity, err := service.collectRecentActivity(7, "test-project")
	if err != nil {
		t.Fatalf("collectRecentActivity failed: %v", err)
	}

	if len(activity) == 0 {
		t.Error("Expected some activity, got none")
	}

	// Check that captures are extracted
	found := false
	for _, capture := range activity {
		if strings.Contains(capture, "Test capture for publish") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected capture content not found in activity")
	}
}

func TestPublishService_ExtractCaptures(t *testing.T) {
	service := NewPublishService()

	testContent := `
## 2025-06-04T10:00:00

First capture content

## 2025-06-04T11:00:00

Second capture with more details
Project: test-project

## 2025-06-04T12:00:00

Third capture
Tags: testing,unit
`

	captures := service.extractCaptures(testContent)

	expectedCount := 3
	if len(captures) != expectedCount {
		t.Errorf("Expected %d captures, got %d", expectedCount, len(captures))
	}

	// Check specific captures
	if !strings.Contains(captures[0], "First capture content") {
		t.Errorf("First capture incorrect: %s", captures[0])
	}

	if !strings.Contains(captures[1], "Second capture with more details") {
		t.Errorf("Second capture incorrect: %s", captures[1])
	}

	// Should not include Tags: lines
	if strings.Contains(captures[2], "Tags:") {
		t.Errorf("Third capture should not include Tags line: %s", captures[2])
	}
}

func TestPublishService_ExtractCapturesEmpty(t *testing.T) {
	service := NewPublishService()

	testContent := `
# Some heading

Regular content without timestamps

## Not a timestamp

More content
`

	captures := service.extractCaptures(testContent)

	if len(captures) != 0 {
		t.Errorf("Expected no captures from non-timestamp content, got %d", len(captures))
	}
}

func TestPublishService_BuildDevlogPrompt(t *testing.T) {
	service := NewPublishService()

	activity := []string{
		"Fixed auth bug",
		"Added new feature",
		"Updated documentation",
	}

	prompt := service.buildDevlogPrompt(activity, "markdown")

	// Check that prompt contains activity
	for _, item := range activity {
		if !strings.Contains(prompt, item) {
			t.Errorf("Prompt missing activity item: %s", item)
		}
	}

	// Check that prompt has expected structure instructions
	if !strings.Contains(prompt, "Technical Implementation") {
		t.Error("Prompt missing Technical Implementation section instruction")
	}

	if !strings.Contains(prompt, "Impact") {
		t.Error("Prompt missing Impact section instruction")
	}
}

func TestPublishService_GetFormatInstructions(t *testing.T) {
	service := NewPublishService()

	tests := []struct {
		format   string
		expected string
	}{
		{"html", "semantic"},
		{"text", "plain text"},
		{"markdown", "Markdown"},
		{"unknown", "Markdown"}, // default case
	}

	for _, test := range tests {
		result := service.getFormatInstructions(test.format)
		if !strings.Contains(result, test.expected) {
			t.Errorf("Format %s: expected %q in result %q", test.format, test.expected, result)
		}
	}
}
