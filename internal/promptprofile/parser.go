package promptprofile

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultClaudeDir = ".claude/projects"

// DiscoverProjects finds all Claude Code project directories.
func DiscoverProjects(claudeDir string) ([]ProjectInfo, error) {
	if claudeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		claudeDir = filepath.Join(home, defaultClaudeDir)
	}

	entries, err := os.ReadDir(claudeDir)
	if err != nil {
		return nil, err
	}

	var projects []ProjectInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirPath := filepath.Join(claudeDir, entry.Name())
		indexPath := filepath.Join(dirPath, "sessions-index.json")

		idx, err := parseSessionIndex(indexPath)
		if err != nil {
			continue // skip dirs without valid index
		}

		projectName := deriveProjectName(entry.Name(), idx)

		projects = append(projects, ProjectInfo{
			Slug:        entry.Name(),
			ProjectName: projectName,
			DirPath:     dirPath,
			Sessions:    idx.Entries,
		})
	}
	return projects, nil
}

func parseSessionIndex(path string) (*SessionIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var idx SessionIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

// deriveProjectName extracts a readable project name.
// Prefers projectPath from the first session entry, falls back to slug parsing.
func deriveProjectName(slug string, idx *SessionIndex) string {
	for _, e := range idx.Entries {
		if e.ProjectPath != "" {
			return filepath.Base(e.ProjectPath)
		}
	}
	// Fallback: parse slug like "-home-qry-projects-uroboro"
	parts := strings.Split(slug, "-")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return slug
}

// ExtractFromProject extracts all user prompts from a project's sessions.
// Filters by days if > 0.
func ExtractFromProject(project ProjectInfo, days int) ([]UserPrompt, error) {
	var cutoff time.Time
	if days > 0 {
		cutoff = time.Now().AddDate(0, 0, -days)
	}

	var allPrompts []UserPrompt
	for _, session := range project.Sessions {
		if session.IsSidechain {
			continue
		}

		// Filter by recency using Modified timestamp
		if days > 0 {
			modified, err := time.Parse(time.RFC3339, session.Modified)
			if err == nil && modified.Before(cutoff) {
				continue
			}
		}

		prompts, err := extractFromSession(session.FullPath, project.ProjectName, session.SessionID)
		if err != nil {
			continue // skip unreadable sessions
		}

		// Mark first real user message
		if len(prompts) > 0 {
			prompts[0].IsFirstMsg = true
		}

		allPrompts = append(allPrompts, prompts...)
	}
	return allPrompts, nil
}

// extractFromSession streams a single session JSONL file and extracts user prompts.
func extractFromSession(jsonlPath, projectName, sessionID string) ([]UserPrompt, error) {
	f, err := os.Open(jsonlPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	var prompts []UserPrompt
	for scanner.Scan() {
		line := scanner.Bytes()

		prompt, ok := parseLine(line, projectName, sessionID)
		if ok {
			prompts = append(prompts, prompt)
		}
	}
	return prompts, scanner.Err()
}

// parseLine parses a single JSONL line and returns a UserPrompt if it's a valid user message.
func parseLine(line []byte, projectName, sessionID string) (UserPrompt, bool) {
	var raw rawLine
	if err := json.Unmarshal(line, &raw); err != nil {
		return UserPrompt{}, false
	}

	// Filter: only user messages
	if raw.Type != "user" {
		return UserPrompt{}, false
	}
	if raw.IsMeta {
		return UserPrompt{}, false
	}
	if raw.ToolUseResult != nil {
		return UserPrompt{}, false
	}
	if raw.Message.Role != "user" {
		return UserPrompt{}, false
	}

	// Extract text content
	text := extractText(raw.Message.Content)
	if text == "" {
		return UserPrompt{}, false
	}

	// Strip injected blocks from user messages
	text = stripXMLBlocks(text, "system-reminder")
	text = stripXMLBlocks(text, "task-notification")
	text = stripXMLBlocks(text, "command-name")
	text = stripXMLBlocks(text, "local-command-result")
	text = stripXMLBlocks(text, "command-message")
	text = stripXMLBlocks(text, "command-args")
	text = stripXMLBlocks(text, "local-command-stdout")
	text = stripXMLBlocks(text, "local-command-stderr")
	text = stripXMLBlocks(text, "bash-notification")
	text = strings.TrimSpace(text)
	if text == "" {
		return UserPrompt{}, false
	}

	// Skip system-generated messages that leak through without XML tags
	if isSystemNoise(text) {
		return UserPrompt{}, false
	}

	prompt := UserPrompt{
		SessionID: sessionID,
		Project:   projectName,
		Timestamp: raw.Timestamp,
		Text:      text,
		WordCount: countWords(text),
		LineCount: strings.Count(text, "\n") + 1,
		Branch:    raw.GitBranch,
	}

	// Classify
	classify(&prompt)

	return prompt, true
}

// extractText handles both string content and array content.
func extractText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Try string first (most common for user messages)
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}

	// Try array of content blocks
	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var texts []string
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				texts = append(texts, b.Text)
			}
		}
		return strings.Join(texts, "\n")
	}

	return ""
}

// stripXMLBlocks removes <tag>...</tag> blocks injected into user message content.
func stripXMLBlocks(text, tag string) string {
	openTag := "<" + tag + ">"
	closeTag := "</" + tag + ">"
	for {
		start := strings.Index(text, openTag)
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], closeTag)
		if end == -1 {
			text = text[:start]
			break
		}
		text = text[:start] + text[start+end+len(closeTag):]
	}
	return text
}

// isSystemNoise detects system-generated messages that lack XML wrapper tags.
func isSystemNoise(text string) bool {
	lower := strings.ToLower(text)
	noisePatterns := []string{
		"read the output file to retrieve the result:",
		"this session is being continued from a previous conversation",
		"<user-prompt-submit-hook>",
	}
	for _, pattern := range noisePatterns {
		if strings.HasPrefix(lower, pattern) || strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func countWords(s string) int {
	return len(strings.Fields(s))
}
