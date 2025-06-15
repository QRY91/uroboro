package journey

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/database"
)

// JourneyService handles journey data processing and timeline generation
type JourneyService struct {
	db *database.DB
}

// NewJourneyService creates a new journey service
func NewJourneyService(db *database.DB) *JourneyService {
	return &JourneyService{db: db}
}

// GenerateJourney creates a complete journey dataset for the specified options
func (j *JourneyService) GenerateJourney(options JourneyOptions) (*JourneyData, error) {
	// Determine date range
	dateRange := j.getDateRange(options)

	// Get captures from database
	captures, err := j.getCapturesInRange(dateRange, options.Projects)
	if err != nil {
		return nil, fmt.Errorf("failed to get captures: %w", err)
	}

	// Get git commits if available
	commits := j.getGitCommitsInRange(dateRange)

	// Process captures into timeline events
	events := j.processCapturesToEvents(captures)

	// Process commits into timeline events
	commitEvents := j.processCommitsToEvents(commits)

	// Combine and sort all events
	allEvents := append(events, commitEvents...)
	sort.Slice(allEvents, func(i, k int) bool {
		return allEvents[i].Timestamp.Before(allEvents[k].Timestamp)
	})

	// Generate project summaries
	projects := j.generateProjectSummaries(allEvents)

	// Calculate journey statistics
	stats := j.calculateJourneyStats(allEvents, projects)

	// Identify milestones
	milestones := j.identifyMilestones(allEvents)

	return &JourneyData{
		Events:     allEvents,
		DateRange:  dateRange,
		Projects:   projects,
		Stats:      stats,
		Milestones: milestones,
	}, nil
}

// getDateRange determines the date range based on options
func (j *JourneyService) getDateRange(options JourneyOptions) DateRange {
	if options.DateRange != nil {
		return *options.DateRange
	}

	end := time.Now()
	start := end.AddDate(0, 0, -options.Days)

	return DateRange{Start: start, End: end}
}

// getCapturesInRange retrieves captures from the database within the specified date range
func (j *JourneyService) getCapturesInRange(dateRange DateRange, projects []string) ([]database.Capture, error) {
	if len(projects) > 0 {
		// Filter by specific projects
		var allCaptures []database.Capture
		for _, project := range projects {
			captures, err := j.db.GetCapturesByProject(project)
			if err != nil {
				return nil, err
			}
			// Filter by date range
			for _, capture := range captures {
				if capture.Timestamp.After(dateRange.Start) && capture.Timestamp.Before(dateRange.End) {
					allCaptures = append(allCaptures, capture)
				}
			}
		}
		return allCaptures, nil
	}

	// Get all captures in date range
	return j.db.GetCapturesSince(dateRange.Start)
}

// getGitCommitsInRange retrieves git commits within the specified date range
func (j *JourneyService) getGitCommitsInRange(dateRange DateRange) []GitCommit {
	cmd := exec.Command("git", "log",
		"--pretty=format:%H|%s|%at|%an",
		fmt.Sprintf("--since=%s", dateRange.Start.Format("2006-01-02")),
		fmt.Sprintf("--until=%s", dateRange.End.Format("2006-01-02")))

	output, err := cmd.Output()
	if err != nil {
		return []GitCommit{}
	}

	var commits []GitCommit
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			timestamp, err := time.Parse("1136239445", parts[2])
			if err != nil {
				continue
			}
			commits = append(commits, GitCommit{
				Hash:      parts[0],
				Message:   parts[1],
				Timestamp: timestamp,
				Author:    parts[3],
			})
		}
	}
	return commits
}

// processCapturesToEvents converts database captures to timeline events
func (j *JourneyService) processCapturesToEvents(captures []database.Capture) []TimelineEvent {
	var events []TimelineEvent

	for _, capture := range captures {
		var project string
		if capture.Project.Valid {
			project = capture.Project.String
		}

		var tags []string
		if capture.Tags.Valid && capture.Tags.String != "" {
			tags = strings.Split(capture.Tags.String, ",")
		}

		event := TimelineEvent{
			Timestamp:  capture.Timestamp,
			Content:    capture.Content,
			Project:    project,
			Tags:       tags,
			EventType:  j.determineEventType(capture),
			Importance: j.calculateImportance(capture),
		}

		// Clean up tags
		for i, tag := range event.Tags {
			event.Tags[i] = strings.TrimSpace(tag)
		}

		events = append(events, event)
	}

	return events
}

// processCommitsToEvents converts git commits to timeline events
func (j *JourneyService) processCommitsToEvents(commits []GitCommit) []TimelineEvent {
	var events []TimelineEvent

	for _, commit := range commits {
		event := TimelineEvent{
			Timestamp:  commit.Timestamp,
			Content:    commit.Message,
			Project:    j.inferProjectFromCommit(commit),
			Tags:       []string{"git", "commit"},
			EventType:  EventTypeCommit,
			Importance: j.calculateCommitImportance(commit),
			GitHash:    commit.Hash,
		}

		events = append(events, event)
	}

	return events
}

// generateProjectSummaries creates project summaries from events
func (j *JourneyService) generateProjectSummaries(events []TimelineEvent) []ProjectSummary {
	projectMap := make(map[string]*ProjectSummary)
	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FECA57", "#FF9FF3", "#54A0FF", "#5F27CD"}
	colorIndex := 0

	for _, event := range events {
		if event.Project == "" {
			continue
		}

		if _, exists := projectMap[event.Project]; !exists {
			projectMap[event.Project] = &ProjectSummary{
				Name:       event.Project,
				EventCount: 0,
				Color:      colors[colorIndex%len(colors)],
				StartDate:  event.Timestamp,
				LastActive: event.Timestamp,
			}
			colorIndex++
		}

		summary := projectMap[event.Project]
		summary.EventCount++

		if event.Timestamp.Before(summary.StartDate) {
			summary.StartDate = event.Timestamp
		}
		if event.Timestamp.After(summary.LastActive) {
			summary.LastActive = event.Timestamp
		}
	}

	var summaries []ProjectSummary
	for _, summary := range projectMap {
		summaries = append(summaries, *summary)
	}

	// Sort by event count descending
	sort.Slice(summaries, func(i, k int) bool {
		return summaries[i].EventCount > summaries[k].EventCount
	})

	return summaries
}

// calculateJourneyStats computes statistics for the journey
func (j *JourneyService) calculateJourneyStats(events []TimelineEvent, projects []ProjectSummary) JourneyStats {
	stats := JourneyStats{
		TotalEvents:     len(events),
		ProjectCount:    len(projects),
		MilestoneCount:  0,
		LearningMoments: 0,
	}

	for _, event := range events {
		if event.EventType == EventTypeMilestone {
			stats.MilestoneCount++
		}
		if event.EventType == EventTypeLearning || containsLearningTag(event.Tags) {
			stats.LearningMoments++
		}
	}

	// Calculate productivity score based on events per day and importance
	if len(events) > 0 {
		totalImportance := 0
		for _, event := range events {
			totalImportance += event.Importance
		}
		stats.ProductivityScore = float64(totalImportance) / float64(len(events))
	}

	return stats
}

// identifyMilestones finds significant milestone events
func (j *JourneyService) identifyMilestones(events []TimelineEvent) []TimelineEvent {
	var milestones []TimelineEvent

	for _, event := range events {
		if j.isMilestone(event) {
			milestones = append(milestones, event)
		}
	}

	return milestones
}

// Helper methods for event processing
func (j *JourneyService) determineEventType(capture database.Capture) string {
	var tags string
	if capture.Tags.Valid {
		tags = strings.ToLower(capture.Tags.String)
	}
	content := strings.ToLower(capture.Content)

	if strings.Contains(tags, "milestone") || strings.Contains(content, "milestone") {
		return EventTypeMilestone
	}
	if strings.Contains(tags, "learning") || strings.Contains(content, "learned") {
		return EventTypeLearning
	}
	if strings.Contains(tags, "decision") || strings.Contains(content, "decided") {
		return EventTypeDecision
	}
	if strings.Contains(tags, "integration") || strings.Contains(content, "integrated") {
		return EventTypeIntegration
	}
	if strings.Contains(tags, "bug") || strings.Contains(content, "fixed") {
		return EventTypeBugfix
	}
	if strings.Contains(tags, "feature") || strings.Contains(content, "implemented") {
		return EventTypeFeature
	}

	return EventTypeCapture
}

func (j *JourneyService) calculateImportance(capture database.Capture) int {
	var tags string
	if capture.Tags.Valid {
		tags = strings.ToLower(capture.Tags.String)
	}
	content := strings.ToLower(capture.Content)

	if strings.Contains(tags, "critical") || strings.Contains(content, "critical") {
		return ImportanceCritical
	}
	if strings.Contains(tags, "milestone") || strings.Contains(tags, "important") {
		return ImportanceHigh
	}
	if strings.Contains(tags, "decision") || strings.Contains(tags, "integration") {
		return ImportanceMedium
	}

	return ImportanceLow
}

func (j *JourneyService) calculateCommitImportance(commit GitCommit) int {
	message := strings.ToLower(commit.Message)

	if strings.Contains(message, "feat") || strings.Contains(message, "feature") {
		return ImportanceHigh
	}
	if strings.Contains(message, "fix") || strings.Contains(message, "bug") {
		return ImportanceMedium
	}
	if strings.Contains(message, "refactor") || strings.Contains(message, "improve") {
		return ImportanceMedium
	}

	return ImportanceLow
}

func (j *JourneyService) inferProjectFromCommit(commit GitCommit) string {
	// Try to infer project from commit message or current directory
	// This is a simple heuristic - could be improved
	return "default"
}

func (j *JourneyService) isMilestone(event TimelineEvent) bool {
	return event.EventType == EventTypeMilestone || event.Importance >= ImportanceHigh
}

func containsLearningTag(tags []string) bool {
	for _, tag := range tags {
		if strings.ToLower(tag) == "learning" || strings.ToLower(tag) == "insight" {
			return true
		}
	}
	return false
}

// GitCommit represents a git commit for timeline processing
type GitCommit struct {
	Hash      string
	Message   string
	Timestamp time.Time
	Author    string
}
