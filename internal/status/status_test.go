package status

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStatusService_ShowStatus(t *testing.T) {
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	service := NewStatusService()

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

Test capture content for status

## 2025-06-04T11:00:00

Another test capture
`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test should not error
	err = service.ShowStatus(7, "", "")
	if err != nil {
		t.Errorf("ShowStatus failed: %v", err)
	}
}

func TestStatusService_ShowStatusNoDirectory(t *testing.T) {
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	service := NewStatusService()

	// Test with non-existent directory should not error
	err := service.ShowStatus(7, "", "")
	if err != nil {
		t.Errorf("ShowStatus should handle missing directory gracefully: %v", err)
	}
}

func TestStatusService_ExtractRecentCaptures(t *testing.T) {
	service := NewStatusService()

	testContent := `
## 2025-06-04T10:00:00

First capture content

## 2025-06-04T11:00:00

Second capture content with more details

## Some other heading

Not a capture

## 2025-06-04T12:00:00

Third capture
`

	captures := service.extractRecentCaptures(testContent)

	expectedCaptures := []string{
		"First capture content",
		"Second capture content with more details",
		"Third capture",
	}

	if len(captures) != len(expectedCaptures) {
		t.Errorf("Expected %d captures, got %d", len(expectedCaptures), len(captures))
	}

	for i, expected := range expectedCaptures {
		if i >= len(captures) || !strings.Contains(captures[i], expected) {
			t.Errorf("Capture %d: expected %q, got %q", i, expected, captures[i])
		}
	}
}

func TestStatusService_ExtractCapturesWithTags(t *testing.T) {
	service := NewStatusService()

	testContent := `
## 2025-06-04T10:00:00

Capture with metadata
Project: test-project
Tags: testing,unit
`

	captures := service.extractRecentCaptures(testContent)

	if len(captures) != 1 {
		t.Errorf("Expected 1 capture, got %d", len(captures))
	}

	// Should extract the capture content but not the metadata lines
	if !strings.Contains(captures[0], "Capture with metadata") {
		t.Errorf("Capture content missing. Got: %s", captures[0])
	}

	// Tags line should be filtered out
	if strings.Contains(captures[0], "Tags:") {
		t.Errorf("Tags metadata should be filtered out. Got: %s", captures[0])
	}
}

// TDD Test: Tag filtering functionality (will fail until implemented)
func TestStatusService_ShowStatusWithTagFilter(t *testing.T) {
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	service := NewStatusService()

	// Create test data directory
	dataDir := filepath.Join(tempHome, ".local", "share", "uroboro", "daily")
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test file with captures having different tags
	today := time.Now().Format("2006-01-02")
	testFile := filepath.Join(dataDir, today+".md")
	testContent := `
## 2025-06-04T10:00:00

Bugfix capture content
Project: test-project
Tags: bugfix,urgent

## 2025-06-04T11:00:00

Feature capture content  
Project: test-project
Tags: feature,enhancement

## 2025-06-04T12:00:00

Another bugfix
Project: test-project
Tags: bugfix,database
`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test filtering by "bugfix" tag - should return 2 captures
	// This will fail until we implement the --tags flag
	err = service.ShowStatusWithTags(7, "", "", "bugfix")
	if err != nil {
		t.Errorf("ShowStatusWithTags failed: %v", err)
	}

	// TODO: Add assertions for filtered output once method exists
	// For now, this test defines the contract:
	// ShowStatusWithTags(days, dbPath, project, tags) should filter by tags
}
