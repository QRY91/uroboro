package ripcord

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/context"
	"github.com/QRY91/uroboro/internal/database"
)

// DatabaseInterface defines the methods needed by RipcordService
type DatabaseInterface interface {
	GetRecentCapturesWithLimit(limit int) ([]database.Capture, error)
	GetCapturesSince(since time.Time) ([]database.Capture, error)
	GetCapturesByProject(project string) ([]database.Capture, error)
}

// RipcordService handles instant context extraction and clipboard operations
type RipcordService struct {
	db DatabaseInterface
}

// NewRipcordService creates a new ripcord service
func NewRipcordService(db DatabaseInterface) *RipcordService {
	return &RipcordService{db: db}
}

// ContextSummary represents different types of context summaries
type ContextSummary struct {
	Type        string
	Content     string
	Project     string
	Timestamp   time.Time
	RecentWork  []string
	Suggestions []string
}

// ExtractCurrentContext generates a comprehensive context summary
func (r *RipcordService) ExtractCurrentContext() (*ContextSummary, error) {
	detector := context.NewProjectDetector()
	project := detector.DetectProject()

	summary := &ContextSummary{
		Type:      "current_context",
		Project:   project,
		Timestamp: time.Now(),
	}

	// Get recent captures for context
	if r.db != nil {
		recentCaptures, err := r.getRecentCaptures(project, 5)
		if err == nil {
			summary.RecentWork = recentCaptures
		}
	}

	// Generate context content
	summary.Content = r.formatContextSummary(summary)

	// Add AI collaboration suggestions
	summary.Suggestions = r.generateCollaborationSuggestions(summary)

	return summary, nil
}

// ExtractRecentWork generates a summary of recent work activity
func (r *RipcordService) ExtractRecentWork(days int) (*ContextSummary, error) {
	detector := context.NewProjectDetector()
	project := detector.DetectProject()

	summary := &ContextSummary{
		Type:      "recent_work",
		Project:   project,
		Timestamp: time.Now(),
	}

	if r.db != nil {
		// Get captures from the last N days
		recentCaptures, err := r.getRecentCapturesInDays(project, days)
		if err == nil {
			summary.RecentWork = recentCaptures
		}
	}

	summary.Content = r.formatRecentWorkSummary(summary, days)
	return summary, nil
}

// ExtractProjectSummary generates a comprehensive project overview
func (r *RipcordService) ExtractProjectSummary(projectName string) (*ContextSummary, error) {
	if projectName == "" {
		detector := context.NewProjectDetector()
		projectName = detector.DetectProject()
	}

	summary := &ContextSummary{
		Type:      "project_summary",
		Project:   projectName,
		Timestamp: time.Now(),
	}

	if r.db != nil {
		// Get all captures for this project
		allCaptures, err := r.getAllProjectCaptures(projectName)
		if err == nil {
			summary.RecentWork = allCaptures
		}
	}

	summary.Content = r.formatProjectSummary(summary)
	return summary, nil
}

// CopyToClipboard copies the given content to system clipboard
func (r *RipcordService) CopyToClipboard(content string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("pbcopy")
	case "linux":
		// Try xclip first, fall back to xsel
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard utility found (install xclip or xsel)")
		}
	case "windows":
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if cmd == nil {
		return fmt.Errorf("failed to create clipboard command")
	}

	cmd.Stdin = strings.NewReader(content)
	return cmd.Run()
}

// QuickRipcord performs instant context extraction and clipboard copy
func (r *RipcordService) QuickRipcord() error {
	context, err := r.ExtractCurrentContext()
	if err != nil {
		return fmt.Errorf("failed to extract context: %w", err)
	}

	err = r.CopyToClipboard(context.Content)
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	fmt.Printf("📋 Context copied to clipboard (%d chars)\n", len(context.Content))
	if context.Project != "" {
		fmt.Printf("   Project: %s\n", context.Project)
	}

	return nil
}

// WorkRipcord extracts recent work and copies to clipboard
func (r *RipcordService) WorkRipcord(days int) error {
	summary, err := r.ExtractRecentWork(days)
	if err != nil {
		return fmt.Errorf("failed to extract recent work: %w", err)
	}

	err = r.CopyToClipboard(summary.Content)
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	fmt.Printf("📋 Recent work summary copied to clipboard (%d days, %d chars)\n",
		days, len(summary.Content))

	return nil
}

// ProjectRipcord extracts project summary and copies to clipboard
func (r *RipcordService) ProjectRipcord(projectName string) error {
	summary, err := r.ExtractProjectSummary(projectName)
	if err != nil {
		return fmt.Errorf("failed to extract project summary: %w", err)
	}

	err = r.CopyToClipboard(summary.Content)
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	fmt.Printf("📋 Project summary copied to clipboard (%s, %d chars)\n",
		summary.Project, len(summary.Content))

	return nil
}

// Private helper methods

func (r *RipcordService) getRecentCaptures(project string, limit int) ([]string, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	captures, err := r.db.GetRecentCapturesWithLimit(limit)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, capture := range captures {
		// Filter by project if specified and capture has project
		if project != "" && capture.Project.Valid && capture.Project.String != project {
			continue
		}
		results = append(results, capture.Content)
	}

	return results, nil
}

func (r *RipcordService) getRecentCapturesInDays(project string, days int) ([]string, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	since := time.Now().AddDate(0, 0, -days)
	captures, err := r.db.GetCapturesSince(since)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, capture := range captures {
		if project != "" && capture.Project.Valid && capture.Project.String != project {
			continue
		}
		results = append(results, capture.Content)
	}

	return results, nil
}

func (r *RipcordService) getAllProjectCaptures(project string) ([]string, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	captures, err := r.db.GetCapturesByProject(project)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, capture := range captures {
		results = append(results, capture.Content)
	}

	return results, nil
}

func (r *RipcordService) formatContextSummary(summary *ContextSummary) string {
	var content strings.Builder

	content.WriteString("# Current Development Context\n\n")
	content.WriteString(fmt.Sprintf("**Timestamp**: %s\n", summary.Timestamp.Format("2006-01-02 15:04:05")))

	if summary.Project != "" {
		content.WriteString(fmt.Sprintf("**Project**: %s\n", summary.Project))
	}

	content.WriteString("\n## Recent Work\n\n")

	if len(summary.RecentWork) > 0 {
		for i, work := range summary.RecentWork {
			if i >= 5 { // Limit to 5 most recent
				break
			}
			content.WriteString(fmt.Sprintf("- %s\n", work))
		}
	} else {
		content.WriteString("No recent captures available.\n")
	}

	content.WriteString("\n## AI Collaboration Context\n\n")
	content.WriteString("This is the current state of development work. ")
	content.WriteString("Use this context to understand what I'm working on and help with related tasks.\n")

	return content.String()
}

func (r *RipcordService) formatRecentWorkSummary(summary *ContextSummary, days int) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# Recent Work Summary (%d days)\n\n", days))
	content.WriteString(fmt.Sprintf("**Generated**: %s\n", summary.Timestamp.Format("2006-01-02 15:04:05")))

	if summary.Project != "" {
		content.WriteString(fmt.Sprintf("**Project**: %s\n", summary.Project))
	}

	content.WriteString("\n## Activity Overview\n\n")

	if len(summary.RecentWork) > 0 {
		for _, work := range summary.RecentWork {
			content.WriteString(fmt.Sprintf("- %s\n", work))
		}
	} else {
		content.WriteString("No recorded activity in the specified time period.\n")
	}

	return content.String()
}

func (r *RipcordService) formatProjectSummary(summary *ContextSummary) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# Project Summary: %s\n\n", summary.Project))
	content.WriteString(fmt.Sprintf("**Generated**: %s\n", summary.Timestamp.Format("2006-01-02 15:04:05")))

	content.WriteString("\n## All Captured Work\n\n")

	if len(summary.RecentWork) > 0 {
		for _, work := range summary.RecentWork {
			content.WriteString(fmt.Sprintf("- %s\n", work))
		}

		content.WriteString(fmt.Sprintf("\n**Total Captures**: %d\n", len(summary.RecentWork)))
	} else {
		content.WriteString("No captures recorded for this project.\n")
	}

	return content.String()
}

func (r *RipcordService) generateCollaborationSuggestions(summary *ContextSummary) []string {
	suggestions := []string{
		"Ask AI to help analyze recent work patterns",
		"Request code review or architecture feedback",
		"Generate documentation from recent captures",
		"Identify knowledge gaps for learning",
	}

	if len(summary.RecentWork) >= 3 {
		suggestions = append(suggestions, "Ask for synthesis of recent work themes")
	}

	if summary.Project != "" {
		suggestions = append(suggestions, fmt.Sprintf("Request project-specific guidance for %s", summary.Project))
	}

	return suggestions
}
