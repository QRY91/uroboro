package openviking

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/database"
)

// DirectSync writes captures directly to the AGFS filesystem and generates
// embeddings via Ollama, bypassing OpenViking's VLM-dependent session pipeline.
type DirectSync struct {
	AGFSRoot     string // e.g., ~/.openviking/data/viking/default/user/default/memories/events
	OllamaURL    string // e.g., http://localhost:11434
	EmbedModel   string // e.g., nomic-embed-text
	OpenVikingURL string // e.g., http://localhost:1933
}

// NewDirectSync creates a DirectSync with sensible defaults.
func NewDirectSync() *DirectSync {
	home, _ := os.UserHomeDir()
	return &DirectSync{
		AGFSRoot:     filepath.Join(home, ".openviking/data/viking/default/user/default/memories/events"),
		OllamaURL:    "http://localhost:11434",
		EmbedModel:   "nomic-embed-text",
		OpenVikingURL: "http://localhost:1933",
	}
}

// DirectSyncResult tracks outcomes.
type DirectSyncResult struct {
	Written   int
	Embedded  int
	Skipped   int
	Errors    int
}

// captureToFilename generates a deterministic filename from capture ID.
func captureToFilename(c database.Capture) string {
	// Deterministic: hash of ID ensures idempotent sync
	h := sha256.Sum256([]byte(fmt.Sprintf("uro-%d", c.ID)))
	return fmt.Sprintf("uro_%x.md", h[:8])
}

// captureToMarkdown converts a capture to a structured markdown document.
func captureToMarkdown(c database.Capture) string {
	ctype := captureType(c.Tags)
	var b strings.Builder

	// Title line
	b.WriteString(fmt.Sprintf("# [%s] %s\n\n", strings.ToUpper(ctype), firstLine(c.Content)))

	// Full content if multi-line
	lines := strings.Split(c.Content, "\n")
	if len(lines) > 1 {
		b.WriteString(strings.Join(lines[1:], "\n"))
		b.WriteString("\n\n")
	}

	// Metadata block
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("type: %s\n", ctype))
	if c.Project != "" {
		b.WriteString(fmt.Sprintf("project: %s\n", c.Project))
	}
	if c.Branch != "" {
		b.WriteString(fmt.Sprintf("branch: %s\n", c.Branch))
	}
	nonTypeTags := filterTypeTags(c.Tags)
	if len(nonTypeTags) > 0 {
		b.WriteString(fmt.Sprintf("tags: %s\n", strings.Join(nonTypeTags, ", ")))
	}
	b.WriteString(fmt.Sprintf("captured: %s\n", c.Timestamp.Format("2006-01-02 15:04")))
	b.WriteString(fmt.Sprintf("uroboro_id: %d\n", c.ID))
	return b.String()
}

// captureToAbstract generates a concise L0 abstract from the capture.
// No LLM needed — the capture format IS the abstract.
func captureToAbstract(c database.Capture) string {
	ctype := captureType(c.Tags)
	content := firstLine(c.Content)
	if len(content) > 200 {
		content = content[:200] + "..."
	}
	project := c.Project
	if project == "" {
		project = "unknown"
	}
	return fmt.Sprintf("[%s] %s (project: %s)", strings.ToUpper(ctype), content, project)
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}

func filterTypeTags(tags string) []string {
	var out []string
	for _, tag := range strings.Split(tags, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" && !typeTags[strings.ToLower(tag)] {
			out = append(out, tag)
		}
	}
	return out
}

// ollamaEmbed calls Ollama's embedding API directly.
func (ds *DirectSync) ollamaEmbed(text string) ([]float64, error) {
	body := map[string]string{
		"model":  ds.EmbedModel,
		"prompt": text,
	}
	data, _ := json.Marshal(body)

	resp, err := http.Post(ds.OllamaURL+"/api/embeddings", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("ollama embed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embedding: %w", err)
	}
	if len(result.Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}
	return result.Embedding, nil
}

// upsertVector inserts a vector entry into OpenViking's index via REST.
func (ds *DirectSync) upsertVector(uri, abstract, contextType string, parentURI string, embedding []float64) error {
	body := map[string]interface{}{
		"uri":          uri,
		"parent_uri":   parentURI,
		"abstract":     abstract,
		"context_type": contextType,
		"is_leaf":      true,
		"level":        2, // L2 = full content
		"vector":       embedding,
	}
	data, _ := json.Marshal(body)

	// Use OpenViking's internal debug/upsert endpoint
	req, _ := http.NewRequest("POST", ds.OpenVikingURL+"/api/v1/debug/upsert", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upsert vector: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upsert failed (%d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// SyncDirect writes captures as markdown files and generates embeddings.
func (ds *DirectSync) SyncDirect(captures []database.Capture, opts SyncOptions) (*DirectSyncResult, error) {
	if err := os.MkdirAll(ds.AGFSRoot, 0755); err != nil {
		return nil, fmt.Errorf("create AGFS dir: %w", err)
	}

	result := &DirectSyncResult{}
	parentURI := "viking://user/default/memories/events"

	for i, c := range captures {
		filename := captureToFilename(c)
		filepath := filepath.Join(ds.AGFSRoot, filename)
		uri := parentURI + "/" + filename

		// Skip if already written (idempotent)
		if _, err := os.Stat(filepath); err == nil {
			result.Skipped++
			if opts.Verbose {
				fmt.Printf("  [%d/%d] skip (exists): %s\n", i+1, len(captures), truncate(c.Content, 60))
			}
			continue
		}

		if opts.DryRun {
			fmt.Printf("  [%d/%d] would write: %s → %s\n", i+1, len(captures), truncate(c.Content, 60), filename)
			result.Written++
			continue
		}

		// 1. Write markdown file
		md := captureToMarkdown(c)
		if err := os.WriteFile(filepath, []byte(md), 0644); err != nil {
			fmt.Printf("  [%d/%d] ERROR writing: %v\n", i+1, len(captures), err)
			result.Errors++
			continue
		}
		result.Written++

		// 2. Generate embedding via Ollama
		abstract := captureToAbstract(c)
		embedding, err := ds.ollamaEmbed(abstract)
		if err != nil {
			fmt.Printf("  [%d/%d] ERROR embedding: %v\n", i+1, len(captures), err)
			// File is written but not indexed — can retry later
			continue
		}
		result.Embedded++

		// 3. Insert into vector index
		if err := ds.upsertVector(uri, abstract, "memory", parentURI, embedding); err != nil {
			if opts.Verbose {
				fmt.Printf("  [%d/%d] WARNING: vector upsert failed (search may not find it): %v\n", i+1, len(captures), err)
			}
			// File is still on disk, just not searchable yet
		}

		if opts.Verbose {
			fmt.Printf("  [%d/%d] ✓ %s\n", i+1, len(captures), truncate(c.Content, 60))
		}
	}

	return result, nil
}
