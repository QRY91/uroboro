package capture

import (
	"testing"
)

func TestNewService(t *testing.T) {
	dbPath := t.TempDir() + "/test.sqlite"

	svc, err := NewService(dbPath)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}
	defer svc.Close()
}

func TestCapture(t *testing.T) {
	dbPath := t.TempDir() + "/test.sqlite"

	svc, err := NewService(dbPath)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}
	defer svc.Close()

	err = svc.Capture("Test capture content", "testproject", "tag1,tag2", nil)
	if err != nil {
		t.Fatalf("Capture failed: %v", err)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a very long string", 10, "this is a ..."},
		{"exact", 5, "exact"},
	}

	for _, test := range tests {
		result := truncate(test.input, test.maxLen)
		if result != test.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", test.input, test.maxLen, result, test.expected)
		}
	}
}
