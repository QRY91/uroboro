package openviking

import (
	"strings"
	"testing"
	"time"

	"github.com/QRY91/uroboro/internal/database"
)

func TestCaptureToMarkdown_Decision(t *testing.T) {
	c := database.Capture{
		ID:        42,
		Timestamp: time.Date(2026, 3, 14, 10, 30, 0, 0, time.UTC),
		Content:   "PostgreSQL over MongoDB — ACID transactions needed",
		Project:   "uroboro",
		Tags:      "decision, architecture",
		Branch:    "main",
	}

	md := captureToMarkdown(c)

	// Title line with type
	if !strings.HasPrefix(md, "# [DECISION]") {
		t.Errorf("expected title to start with # [DECISION], got: %s", firstLine(md))
	}

	// Content preserved
	if !strings.Contains(md, "PostgreSQL over MongoDB") {
		t.Error("content not preserved in markdown")
	}

	// Metadata block
	if !strings.Contains(md, "type: decision") {
		t.Error("missing type in metadata")
	}
	if !strings.Contains(md, "project: uroboro") {
		t.Error("missing project in metadata")
	}
	if !strings.Contains(md, "branch: main") {
		t.Error("missing branch in metadata")
	}
	if !strings.Contains(md, "uroboro_id: 42") {
		t.Error("missing uroboro_id in metadata")
	}
	if !strings.Contains(md, "captured: 2026-03-14 10:30") {
		t.Error("missing captured timestamp in metadata")
	}

	// Type tag filtered out of tags list
	if !strings.Contains(md, "tags: architecture") {
		t.Error("non-type tags should be preserved")
	}
	// "decision" should NOT appear in the tags line
	lines := strings.Split(md, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "tags:") && strings.Contains(line, "decision") {
			t.Error("type tag 'decision' should be filtered from tags metadata")
		}
	}
}

func TestCaptureToMarkdown_Blocker(t *testing.T) {
	c := database.Capture{
		ID:        99,
		Timestamp: time.Date(2026, 3, 14, 9, 0, 0, 0, time.UTC),
		Content:   "Blocked on Stripe API key",
		Project:   "qallo",
		Tags:      "blocker, payments",
	}

	md := captureToMarkdown(c)

	if !strings.HasPrefix(md, "# [BLOCKER]") {
		t.Errorf("expected [BLOCKER] prefix, got: %s", firstLine(md))
	}
	if !strings.Contains(md, "type: blocker") {
		t.Error("missing type: blocker in metadata")
	}
	if !strings.Contains(md, "tags: payments") {
		t.Error("non-type tags should be preserved")
	}
}

func TestCaptureToMarkdown_GenericCapture(t *testing.T) {
	c := database.Capture{
		ID:        7,
		Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Content:   "Built fractal explorer with WebGL2",
		Project:   "fractals",
		Tags:      "webgl, visualization",
	}

	md := captureToMarkdown(c)

	if !strings.HasPrefix(md, "# [CAPTURE]") {
		t.Errorf("expected [CAPTURE] prefix, got: %s", firstLine(md))
	}
	if !strings.Contains(md, "type: capture") {
		t.Error("missing type: capture")
	}
}

func TestCaptureToMarkdown_NoProject(t *testing.T) {
	c := database.Capture{
		ID:        1,
		Timestamp: time.Now(),
		Content:   "Something with no project",
		Tags:      "",
	}

	md := captureToMarkdown(c)

	if strings.Contains(md, "project:") {
		t.Error("should not include project line when empty")
	}
}

func TestCaptureToMarkdown_MetadataSeparator(t *testing.T) {
	c := database.Capture{
		ID:        1,
		Timestamp: time.Now(),
		Content:   "Test content",
		Project:   "test",
		Tags:      "decision",
	}

	md := captureToMarkdown(c)

	if !strings.Contains(md, "\n---\n") {
		t.Error("metadata block should be separated by ---")
	}
}

func TestCaptureToAbstract(t *testing.T) {
	c := database.Capture{
		ID:      1,
		Content: "PostgreSQL over MongoDB — ACID transactions",
		Project: "uroboro",
		Tags:    "decision",
	}

	abs := captureToAbstract(c)

	if !strings.HasPrefix(abs, "[DECISION]") {
		t.Errorf("abstract should start with type, got: %s", abs)
	}
	if !strings.Contains(abs, "PostgreSQL over MongoDB") {
		t.Error("abstract should contain content")
	}
	if !strings.Contains(abs, "(project: uroboro)") {
		t.Error("abstract should contain project")
	}
}

func TestCaptureToAbstract_Truncation(t *testing.T) {
	long := strings.Repeat("x", 300)
	c := database.Capture{
		ID:      1,
		Content: long,
		Project: "test",
		Tags:    "capture",
	}

	abs := captureToAbstract(c)

	if len(abs) > 250 {
		t.Errorf("abstract should truncate long content, got length %d", len(abs))
	}
	if !strings.Contains(abs, "...") {
		t.Error("truncated abstract should end with ...")
	}
}

func TestCaptureToFilename_Deterministic(t *testing.T) {
	c := database.Capture{ID: 42}

	f1 := captureToFilename(c)
	f2 := captureToFilename(c)

	if f1 != f2 {
		t.Errorf("filenames should be deterministic, got %s and %s", f1, f2)
	}
	if !strings.HasPrefix(f1, "uro_") {
		t.Errorf("filename should start with uro_, got: %s", f1)
	}
	if !strings.HasSuffix(f1, ".md") {
		t.Errorf("filename should end with .md, got: %s", f1)
	}
}

func TestCaptureToFilename_UniquePerID(t *testing.T) {
	f1 := captureToFilename(database.Capture{ID: 1})
	f2 := captureToFilename(database.Capture{ID: 2})

	if f1 == f2 {
		t.Error("different IDs should produce different filenames")
	}
}

func TestFilterTypeTags(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"decision, architecture", []string{"architecture"}},
		{"blocker, payments, external", []string{"payments", "external"}},
		{"question", nil},
		{"mcp, dogfood", []string{"mcp", "dogfood"}},
		{"", nil},
	}

	for _, tt := range tests {
		got := filterTypeTags(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("filterTypeTags(%q) = %v, want %v", tt.input, got, tt.expected)
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("filterTypeTags(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.expected[i])
			}
		}
	}
}
