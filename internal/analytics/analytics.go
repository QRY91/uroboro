package analytics

import (
	"log"
	"sync"
	"time"
)

var (
	instance *Analytics
	once     sync.Once
)

// Analytics provides a simple interface for uroboro analytics
type Analytics struct {
	client *AnalyticsClient
}

// Initialize sets up the global analytics instance
func Initialize() {
	once.Do(func() {
		client, err := NewAnalyticsClient()
		if err != nil {
			log.Printf("Warning: Analytics initialization failed: %v", err)
			// Create disabled instance
			instance = &Analytics{client: nil}
			return
		}
		instance = &Analytics{client: client}
	})
}

// Get returns the global analytics instance
func Get() *Analytics {
	if instance == nil {
		Initialize()
	}
	return instance
}

// TrackCaptureSimple tracks a basic capture event with minimal parameters
func (a *Analytics) TrackCaptureSimple(content, project string, tags []string) {
	if a.client == nil {
		return
	}

	event := CaptureEvent{
		Content:       content,
		Project:       project,
		Tags:          tags,
		Method:        "cli",
		InsightLength: len(content),
		CaptureTime:   0, // Will be set by caller if available
		AIAssisted:    false,
		GitContext:    GetGitContext(),
		WorkflowStage: "capture",
		ContextWeight: 0.5,
	}

	if err := a.client.TrackCapture(event); err != nil {
		log.Printf("Analytics capture tracking failed: %v", err)
	}
}

// TrackCaptureDetailed tracks a capture event with full context
func (a *Analytics) TrackCaptureDetailed(content, project string, tags []string, aiAssisted bool, captureTime time.Duration) {
	if a.client == nil {
		return
	}

	event := CaptureEvent{
		Content:       content,
		Project:       project,
		Tags:          tags,
		Method:        "cli",
		InsightLength: len(content),
		CaptureTime:   captureTime,
		AIAssisted:    aiAssisted,
		GitContext:    GetGitContext(),
		WorkflowStage: "capture",
		ContextWeight: calculateContextWeight(content, project),
	}

	if err := a.client.TrackCapture(event); err != nil {
		log.Printf("Analytics capture tracking failed: %v", err)
	}
}

// TrackPublishSimple tracks a basic publish event
func (a *Analytics) TrackPublishSimple(format string, success bool, wordCount int) {
	if a.client == nil {
		return
	}

	event := PublishEvent{
		Format:            format,
		WordCount:         wordCount,
		Platform:          "local",
		Success:           success,
		ContentType:       format,
		AIAssistanceLevel: "none",
		CapturesUsed:      0, // Will be set by caller if available
		GenerationTime:    0,
		QualityScore:      0.5,
	}

	if err := a.client.TrackPublish(event); err != nil {
		log.Printf("Analytics publish tracking failed: %v", err)
	}
}

// TrackPublishDetailed tracks a publish event with full context
func (a *Analytics) TrackPublishDetailed(format string, success bool, wordCount, capturesUsed int, generationTime time.Duration, aiLevel string) {
	if a.client == nil {
		return
	}

	event := PublishEvent{
		Format:            format,
		WordCount:         wordCount,
		Platform:          "local",
		Success:           success,
		ContentType:       format,
		AIAssistanceLevel: aiLevel,
		CapturesUsed:      capturesUsed,
		GenerationTime:    generationTime,
		QualityScore:      calculateQualityScore(wordCount, capturesUsed),
	}

	if err := a.client.TrackPublish(event); err != nil {
		log.Printf("Analytics publish tracking failed: %v", err)
	}
}

// TrackStatusCheck tracks a status command execution
func (a *Analytics) TrackStatusCheck(totalCaptures, capturesToday, activeProjects, unpublishedCaptures int, dbSizeMB float64) {
	if a.client == nil {
		return
	}

	event := StatusEvent{
		TotalCaptures:       totalCaptures,
		CapturesToday:       capturesToday,
		ActiveProjects:      activeProjects,
		ProductivityTrend:   calculateProductivityTrend(capturesToday),
		UnpublishedCaptures: unpublishedCaptures,
		DatabaseSizeMB:      dbSizeMB,
	}

	if err := a.client.TrackStatus(event); err != nil {
		log.Printf("Analytics status tracking failed: %v", err)
	}
}

// TrackAISession tracks an AI collaboration session
func (a *Analytics) TrackAISession(duration time.Duration, queries int, satisfaction float64) {
	if a.client == nil {
		return
	}

	if err := a.client.TrackAICollaboration(duration, queries, satisfaction); err != nil {
		log.Printf("Analytics AI session tracking failed: %v", err)
	}
}

// TrackWorkflowChange tracks a workflow state transition
func (a *Analytics) TrackWorkflowChange(from, to, trigger string, duration time.Duration) {
	if a.client == nil {
		return
	}

	if err := a.client.TrackWorkflowTransition(from, to, trigger, duration); err != nil {
		log.Printf("Analytics workflow tracking failed: %v", err)
	}
}

// Close gracefully shuts down analytics
func (a *Analytics) Close() {
	if a.client != nil {
		if err := a.client.Close(); err != nil {
			log.Printf("Analytics shutdown error: %v", err)
		}
	}
}

// IsEnabled returns true if analytics is enabled and working
func (a *Analytics) IsEnabled() bool {
	return a.client != nil && a.client.isEnabled()
}

// Helper functions

func calculateContextWeight(content, project string) float64 {
	weight := 0.5 // base weight

	// Increase weight for longer, more detailed content
	if len(content) > 100 {
		weight += 0.2
	}
	if len(content) > 500 {
		weight += 0.2
	}

	// Increase weight for QRY ecosystem projects
	if project == "uroboro" || project == "doggowoof" || project == "qry-labs" {
		weight += 0.1
	}

	// Cap at 1.0
	if weight > 1.0 {
		weight = 1.0
	}

	return weight
}

func calculateQualityScore(wordCount, capturesUsed int) float64 {
	if wordCount == 0 {
		return 0.0
	}

	// Base score on word count and capture usage
	score := 0.5

	// Higher score for substantial content
	if wordCount > 500 {
		score += 0.2
	}
	if wordCount > 1000 {
		score += 0.2
	}

	// Higher score for using more captures (better synthesis)
	if capturesUsed > 5 {
		score += 0.1
	}
	if capturesUsed > 10 {
		score += 0.1
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

func calculateProductivityTrend(capturesToday int) float64 {
	// Simple productivity calculation based on captures today
	// In a real implementation, this would compare to historical data
	if capturesToday >= 10 {
		return 1.0 // high productivity
	}
	if capturesToday >= 5 {
		return 0.7 // good productivity
	}
	if capturesToday >= 2 {
		return 0.5 // moderate productivity
	}
	return 0.2 // low productivity
}

// QuickStart initializes analytics with minimal configuration
func QuickStart() {
	Initialize()
	if Get().IsEnabled() {
		log.Printf("✅ PostHog analytics enabled for uroboro")
	} else {
		log.Printf("ℹ️  PostHog analytics disabled (configure POSTHOG_API_KEY to enable)")
	}
}
