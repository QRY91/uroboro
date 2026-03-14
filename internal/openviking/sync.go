package openviking

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/database"
)

// SyncOptions controls the sync behavior.
type SyncOptions struct {
	BatchSize int
	DryRun    bool
	Verbose   bool
}

// SyncResult tracks the outcome of a sync run.
type SyncResult struct {
	CapturesProcessed int
	MemoriesExtracted int
	BatchCount        int
	Errors            int
}

// typeTags are recognized as capture type classifiers.
var typeTags = map[string]bool{
	"decision": true,
	"blocker":  true,
	"question": true,
}

// captureType classifies a capture based on its tags (mirrors distill logic).
func captureType(tags string) string {
	for _, tag := range strings.Split(tags, ",") {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if typeTags[tag] {
			return tag
		}
	}
	return "capture"
}

// formatCapture renders a capture as structured text for OpenViking memory extraction.
func formatCapture(c database.Capture) string {
	ctype := captureType(c.Tags)
	var lines []string
	lines = append(lines, fmt.Sprintf("[%s] %s", strings.ToUpper(ctype), c.Content))

	var meta []string
	if c.Project != "" {
		meta = append(meta, "project: "+c.Project)
	}
	if c.Branch != "" {
		meta = append(meta, "branch: "+c.Branch)
	}
	// Non-type tags only
	var extraTags []string
	for _, tag := range strings.Split(c.Tags, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" && !typeTags[strings.ToLower(tag)] {
			extraTags = append(extraTags, tag)
		}
	}
	if len(extraTags) > 0 {
		meta = append(meta, "tags: "+strings.Join(extraTags, ", "))
	}
	if len(meta) > 0 {
		lines = append(lines, "("+strings.Join(meta, " | ")+")")
	}
	return strings.Join(lines, "\n")
}

// batchByProject groups captures into batches, grouped by project.
func batchByProject(captures []database.Capture, batchSize int) [][]database.Capture {
	byProject := make(map[string][]database.Capture)
	var projectOrder []string
	for _, c := range captures {
		key := c.Project
		if key == "" {
			key = "_none"
		}
		if _, exists := byProject[key]; !exists {
			projectOrder = append(projectOrder, key)
		}
		byProject[key] = append(byProject[key], c)
	}

	var batches [][]database.Capture
	for _, project := range projectOrder {
		caps := byProject[project]
		for i := 0; i < len(caps); i += batchSize {
			end := i + batchSize
			if end > len(caps) {
				end = len(caps)
			}
			batches = append(batches, caps[i:end])
		}
	}
	return batches
}

// Sync pushes captures to OpenViking as structured memories.
func Sync(client *Client, captures []database.Capture, opts SyncOptions) (*SyncResult, error) {
	if len(captures) == 0 {
		return &SyncResult{}, nil
	}

	batchSize := opts.BatchSize
	if batchSize <= 0 {
		batchSize = 10
	}

	batches := batchByProject(captures, batchSize)
	result := &SyncResult{
		CapturesProcessed: len(captures),
		BatchCount:        len(batches),
	}

	for i, batch := range batches {
		project := batch[0].Project
		if project == "" {
			project = "unknown"
		}

		if opts.DryRun {
			fmt.Printf("  [%d/%d] %s: %d captures (dry run)\n", i+1, len(batches), project, len(batch))
			for _, c := range batch {
				fmt.Printf("    [%s] %s\n", captureType(c.Tags), truncate(c.Content, 80))
			}
			continue
		}

		memories, err := syncBatch(client, batch, project, opts.Verbose)
		if err != nil {
			fmt.Printf("  [%d/%d] %s: ERROR: %v\n", i+1, len(batches), project, err)
			result.Errors++
			continue
		}

		result.MemoriesExtracted += memories
		fmt.Printf("  [%d/%d] %s: %d captures → %d memories\n", i+1, len(batches), project, len(batch), memories)
	}

	return result, nil
}

// syncBatch pushes one batch to OpenViking via session lifecycle.
func syncBatch(client *Client, batch []database.Capture, project string, verbose bool) (int, error) {
	// Create session
	session, err := client.CreateSession()
	if err != nil {
		return 0, fmt.Errorf("create session: %w", err)
	}

	// Context message
	contextMsg := fmt.Sprintf(
		"Development captures from project '%s' (%d entries). "+
			"These record decisions, blockers, questions, and context from active development sessions.",
		project, len(batch),
	)
	if err := client.AddMessage(session.SessionID, "user", contextMsg); err != nil {
		return 0, fmt.Errorf("add context message: %w", err)
	}

	// Each capture as an assistant message
	for _, c := range batch {
		msg := formatCapture(c)
		if verbose {
			fmt.Printf("      → %s\n", truncate(msg, 100))
		}
		if err := client.AddMessage(session.SessionID, "assistant", msg); err != nil {
			return 0, fmt.Errorf("add capture message (id=%d): %w", c.ID, err)
		}
	}

	// Async commit → poll for completion (local LLMs can be slow)
	asyncResult, err := client.CommitSessionAsync(session.SessionID)
	if err != nil {
		return 0, fmt.Errorf("start commit: %w", err)
	}

	if verbose {
		fmt.Printf("      commit started (task: %s)\n", asyncResult.TaskID)
	}

	// Poll with 30min timeout (local models need time)
	task, err := client.WaitForTask(asyncResult.TaskID, 5*time.Second, 30*time.Minute)
	if err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	// Parse memories_extracted from task result
	var commitData struct {
		MemoriesExtracted int `json:"memories_extracted"`
	}
	if task.Result != nil {
		json.Unmarshal(task.Result, &commitData)
	}

	return commitData.MemoriesExtracted, nil
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
