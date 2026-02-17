package distill

import (
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/database"
)

// codeKeywords are content/tag signals that indicate a capture is code-related.
var codeKeywords = []string{
	"architecture", "refactor", "implement", "interface", "schema", "database",
	"api", "struct", "class", "function", "module", "package", "endpoint",
	"middleware", "handler", "pattern", "abstraction", "dependency", "migration",
	"query", "index", "cache", "config", "deploy", "build", "test",
	"lint", "format", "convention", "style", "naming", "error handling",
}

type UroExtractor struct {
	db      *database.DB
	project string
	days    int
	since   *time.Time
}

func NewUroExtractor(db *database.DB, project string, days int, since *time.Time) *UroExtractor {
	return &UroExtractor{db: db, project: project, days: days, since: since}
}

func (e *UroExtractor) Extract() ([]UroExtract, error) {
	captures, err := e.db.QueryCaptures(database.CaptureQuery{
		Project: e.project,
		Days:    e.days,
		Since:   e.since,
	})
	if err != nil {
		return nil, err
	}

	var results []UroExtract
	for _, c := range captures {
		captureType := classifyCaptureType(c.Tags)
		if !shouldInclude(captureType, c.Content, c.Tags) {
			continue
		}
		results = append(results, UroExtract{
			Source:    "uroboro",
			Type:      captureType,
			Content:   c.Content,
			Project:   c.Project,
			Tags:      splitTags(c.Tags),
			Timestamp: c.Timestamp,
		})
	}
	return results, nil
}

// classifyCaptureType returns decision|blocker|question|capture
// by scanning the comma-separated tags string.
func classifyCaptureType(tags string) string {
	for _, t := range strings.Split(tags, ",") {
		switch strings.TrimSpace(t) {
		case "decision":
			return "decision"
		case "blocker":
			return "blocker"
		case "question":
			return "question"
		}
	}
	return "capture"
}

// shouldInclude returns true for decisions/blockers/questions (always),
// and for generic captures only if they contain code-related signals.
func shouldInclude(captureType, content, tags string) bool {
	if captureType != "capture" {
		return true
	}
	lower := strings.ToLower(content + " " + tags)
	for _, kw := range codeKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
