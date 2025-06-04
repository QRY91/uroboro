package publish

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type PublishService struct {
	model string
}

func NewPublishService() *PublishService {
	model := os.Getenv("UROBORO_MODEL")
	if model == "" {
		model = "mistral:latest"
	}
	return &PublishService{model: model}
}

func (p *PublishService) callOllama(prompt string) (string, error) {
	fmt.Printf("[DEBUG] Calling ollama with model: %s, prompt length: %d chars\n", p.model, len(prompt))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ollama", "run", p.model)
	cmd.Stdin = strings.NewReader(prompt)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ollama error: %w", err)
	}

	result := strings.TrimSpace(string(output))
	fmt.Printf("[DEBUG] Ollama success, response length: %d chars\n", len(result))
	return result, nil
}

func (p *PublishService) GenerateDevlog(days int) error {
	fmt.Printf("🔍 Collecting activity from last %d day(s)...\n", days)

	// Get recent captures - simplified version
	activity, err := p.collectRecentActivity(days)
	if err != nil {
		return fmt.Errorf("failed to collect activity: %w", err)
	}

	if len(activity) == 0 {
		fmt.Println("❌ No recent activity found to process")
		fmt.Println("💡 Try: uroboro capture 'your development insight' first")
		return nil
	}

	fmt.Printf("✅ Found %d recent captures\n", len(activity))
	fmt.Println("📋 Generating development log...")

	prompt := p.buildDevlogPrompt(activity)
	content, err := p.callOllama(prompt)
	if err != nil {
		return fmt.Errorf("failed to generate devlog: %w", err)
	}

	fmt.Println("--- DEVLOG SUMMARY ---")
	fmt.Println(content)
	fmt.Println("--- END DEVLOG ---")

	return nil
}

func (p *PublishService) GenerateBlog(days int, title string, preview bool, format string) error {
	fmt.Printf("🔍 Collecting activity from last %d day(s)...\n", days)

	activity, err := p.collectRecentActivity(days)
	if err != nil {
		return fmt.Errorf("failed to collect activity: %w", err)
	}

	if len(activity) == 0 {
		fmt.Println("❌ No recent activity found to process")
		fmt.Println("💡 Try: uroboro capture 'your development insight' first")
		return nil
	}

	fmt.Printf("✅ Found %d recent captures\n", len(activity))
	fmt.Printf("📝 Generating blog post (%s format)...\n", format)

	if title == "" {
		title = fmt.Sprintf("Dev Update - %s", time.Now().Format("January 2, 2006"))
	}

	prompt := p.buildBlogPrompt(activity, title, format)
	content, err := p.callOllama(prompt)
	if err != nil {
		return fmt.Errorf("failed to generate blog: %w", err)
	}

	// Format the content based on the requested format
	fullContent := p.formatContent(content, title, format)

	if preview {
		fmt.Printf("--- %s PREVIEW ---\n", strings.ToUpper(format))
		fmt.Println(fullContent)
		fmt.Printf("--- END %s PREVIEW ---\n", strings.ToUpper(format))
	} else {
		filename := p.saveBlogPost(fullContent, title, format)
		fmt.Printf("✅ Blog post saved to: %s\n", filename)
	}

	return nil
}

func (p *PublishService) collectRecentActivity(days int) ([]string, error) {
	// Get XDG data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dataDir := filepath.Join(homeDir, ".local", "share", "uroboro", "daily")

	var activity []string

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().AddDate(0, 0, -days)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(cutoff) {
			fullPath := filepath.Join(dataDir, entry.Name())
			content, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}

			// Extract captures from markdown content
			captures := p.extractCaptures(string(content))
			activity = append(activity, captures...)
		}
	}

	return activity, nil
}

func (p *PublishService) extractCaptures(content string) []string {
	var captures []string
	scanner := bufio.NewScanner(strings.NewReader(content))

	inCapture := false
	var currentCapture strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Look for timestamp markers
		if strings.HasPrefix(line, "## 2025-") {
			if inCapture && currentCapture.Len() > 0 {
				captures = append(captures, strings.TrimSpace(currentCapture.String()))
				currentCapture.Reset()
			}
			inCapture = true
			continue
		}

		if inCapture {
			if strings.HasPrefix(line, "## ") {
				// New section, end current capture
				if currentCapture.Len() > 0 {
					captures = append(captures, strings.TrimSpace(currentCapture.String()))
					currentCapture.Reset()
				}
				inCapture = false
			} else if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "Tags:") {
				if currentCapture.Len() > 0 {
					currentCapture.WriteString(" ")
				}
				currentCapture.WriteString(strings.TrimSpace(line))
			}
		}
	}

	// Don't forget the last capture
	if inCapture && currentCapture.Len() > 0 {
		captures = append(captures, strings.TrimSpace(currentCapture.String()))
	}

	return captures
}

func (p *PublishService) buildDevlogPrompt(activity []string) string {
	allContent := strings.Join(activity, "\n")

	return fmt.Sprintf(`Analyze this development activity and create a concise development log summary.

IMPORTANT: These are today's most important work captures - prioritize them.

%s

Create a structured summary with:
## Technical Work
- Brief bullet points of what was accomplished

## Key Insights  
- Important discoveries or decisions made

## Next Steps
- What should be tackled next

Keep it professional but conversational. Focus on the most significant items.`, allContent)
}

func (p *PublishService) buildBlogPrompt(activity []string, title string, format string) string {
	todayWork := strings.Join(activity[:min(5, len(activity))], "\n")

	formatInstructions := p.getFormatInstructions(format)

	return fmt.Sprintf(`Write an engaging blog post about today's development work:

TODAY'S CAPTURES:
%s

Structure:
1. Brief intro (1-2 sentences)
2. Main highlights from today's work
3. Brief technical insight or lesson learned
4. Quick note on what's next

Keep it 200-300 words, focused on today's specific achievements.
Write in engaging, professional tone.

%s`, todayWork, formatInstructions)
}

func (p *PublishService) getFormatInstructions(format string) string {
	switch format {
	case "html":
		return "Output clean, semantic content suitable for HTML conversion."
	case "text":
		return "Write in plain text format - no markup, just clean readable content with clear structure."
	default: // markdown
		return "Use standard Markdown formatting with ## headings and `code` blocks."
	}
}

func (p *PublishService) formatContent(content, title, format string) string {
	switch format {
	case "markdown":
		return fmt.Sprintf(`# %s

*%s*

%s`, title, time.Now().Format("January 2, 2006"), content)

	case "html":
		return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
    <meta charset="utf-8">
</head>
<body>
    <h1>%s</h1>
    <p><em>%s</em></p>
    %s
</body>
</html>`, title, title, time.Now().Format("January 2, 2006"), p.markdownToHTML(content))

	case "text":
		return fmt.Sprintf(`%s

%s

%s`, title, time.Now().Format("January 2, 2006"), p.stripMarkdown(content))

	default:
		return content
	}
}

func (p *PublishService) markdownToHTML(markdown string) string {
	// Basic markdown to HTML conversion
	lines := strings.Split(markdown, "\n")
	var htmlLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "## ") {
			htmlLines = append(htmlLines, fmt.Sprintf("    <h2>%s</h2>", line[3:]))
		} else if strings.HasPrefix(line, "# ") {
			htmlLines = append(htmlLines, fmt.Sprintf("    <h1>%s</h1>", line[2:]))
		} else {
			// Convert inline code
			line = strings.ReplaceAll(line, "`", "<code>")
			htmlLines = append(htmlLines, fmt.Sprintf("    <p>%s</p>", line))
		}
	}

	return strings.Join(htmlLines, "\n")
}

func (p *PublishService) stripMarkdown(markdown string) string {
	// Remove markdown formatting for plain text
	text := markdown

	// Remove headers
	text = strings.ReplaceAll(text, "## ", "")
	text = strings.ReplaceAll(text, "# ", "")

	// Remove code formatting
	text = strings.ReplaceAll(text, "`", "")

	// Remove bold/italic
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "*", "")

	// Clean up extra whitespace
	lines := strings.Split(text, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n\n")
}

func (p *PublishService) saveBlogPost(content, title, format string) string {
	// Get the project root (parent of go directory)
	executablePath, err := os.Executable()
	if err != nil {
		executablePath = "."
	}
	execDir := filepath.Dir(executablePath)

	// If we're in the go directory, go up one level to project root
	projectRoot := execDir
	if strings.HasSuffix(execDir, "/go") || strings.HasSuffix(execDir, "\\go") {
		projectRoot = filepath.Dir(execDir)
	}

	// Create output directory at project root
	outputDir := filepath.Join(projectRoot, "output", "posts")
	os.MkdirAll(outputDir, 0755)

	// Generate filename from title and format
	filename := strings.ToLower(title)
	filename = strings.ReplaceAll(filename, " ", "-")
	filename = strings.ReplaceAll(filename, ",", "")

	// Choose extension based on format
	var ext string
	switch format {
	case "html":
		ext = "html"
	case "text":
		ext = "txt"
	default: // markdown
		ext = "md"
	}

	filename = fmt.Sprintf("%s-%s.%s", time.Now().Format("2006-01-02"), filename, ext)

	filepath := filepath.Join(outputDir, filename)
	os.WriteFile(filepath, []byte(content), 0644)

	return filepath
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
