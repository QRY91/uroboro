package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCaptureService_Capture(t *testing.T) {
	// Create temp directory for testing
	tempHome := t.TempDir()

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	service := NewCaptureService()

	// Test basic capture
	content := "Test capture content"
	err := service.Capture(content, "", "")
	if err != nil {
		t.Fatalf("Capture failed: %v", err)
	}

	// Verify file was created in correct XDG location
	today := time.Now().Format("2006-01-02")
	expectedPath := filepath.Join(tempHome, ".local", "share", "uroboro", "daily", today+".md")

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("File not created at expected path: %s", expectedPath)
	}

	// Verify content
	fileContent, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(fileContent)
	if !strings.Contains(contentStr, content) {
		t.Errorf("File content doesn't contain captured text. Got: %s", contentStr)
	}

	// Verify timestamp format
	if !strings.Contains(contentStr, "## 2025-") {
		t.Errorf("File content doesn't contain proper timestamp format. Got: %s", contentStr)
	}
}

func TestCaptureService_CaptureWithProjectAndTags(t *testing.T) {
	tempHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	service := NewCaptureService()

	content := "Test with metadata"
	project := "test-project"
	tags := "testing,unit-test"

	err := service.Capture(content, project, tags)
	if err != nil {
		t.Fatalf("Capture with metadata failed: %v", err)
	}

	// Read and verify content includes metadata
	today := time.Now().Format("2006-01-02")
	filePath := filepath.Join(tempHome, ".local", "share", "uroboro", "daily", today+".md")

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(fileContent)
	if !strings.Contains(contentStr, "Project: "+project) {
		t.Errorf("File content missing project metadata. Got: %s", contentStr)
	}

	if !strings.Contains(contentStr, "Tags: "+tags) {
		t.Errorf("File content missing tags metadata. Got: %s", contentStr)
	}
}

func TestTruncateContent(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a very long string that should be truncated", 20, "this is a very long ..."},
		{"exact length", 12, "exact length"},
	}

	for _, test := range tests {
		result := truncateContent(test.input, test.maxLen)
		if result != test.expected {
			t.Errorf("truncateContent(%q, %d) = %q, want %q", test.input, test.maxLen, result, test.expected)
		}
	}
}
