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
	client         *AnalyticsClient
	sessionManager *SessionManager
}

// Initialize sets up the global analytics instance
func Initialize() {
	once.Do(func() {
		client, err := NewAnalyticsClient()
		if err != nil {
			log.Printf("Warning: Analytics initialization failed: %v", err)
			// Create disabled instance with session manager
			instance = &Analytics{
				client:         nil,
				sessionManager: NewSessionManager(nil),
			}
			return
		}
		instance = &Analytics{
			client:         client,
			sessionManager: NewSessionManager(client),
		}
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
	startTime := time.Now()

	// Track in session
	metadata := map[string]interface{}{
		"content_length": len(content),
		"tags":           tags,
		"method":         "cli",
	}
	a.sessionManager.StartActivity("capture", project, metadata)

	// Complete the activity
	duration := time.Since(startTime)
	a.sessionManager.CompleteActivity(true, len(content), duration)

	// ALSO send individual event for debugging
	if a.client != nil {
		event := CaptureEvent{
			Content:       content,
			Project:       project,
			Tags:          tags,
			Method:        "cli",
			InsightLength: len(content),
			CaptureTime:   duration,
			AIAssisted:    false,
			GitContext:    GetGitContext(),
			WorkflowStage: "capture",
			ContextWeight: calculateContextWeight(content, project),
		}

		if err := a.client.TrackCapture(event); err != nil {
			log.Printf("Individual capture tracking failed: %v", err)
		} else {
			log.Printf("✅ DEBUG: Individual capture event sent")
		}
	}
}

// TrackCaptureDetailed tracks a capture event with full context
func (a *Analytics) TrackCaptureDetailed(content, project string, tags []string, aiAssisted bool, captureTime time.Duration) {
	// Track in session with detailed metadata
	metadata := map[string]interface{}{
		"content_length": len(content),
		"tags":           tags,
		"method":         "cli",
		"ai_assisted":    aiAssisted,
		"context_weight": calculateContextWeight(content, project),
	}
	a.sessionManager.StartActivity("capture", project, metadata)
	a.sessionManager.CompleteActivity(true, len(content), captureTime)
}

// TrackPublishSimple tracks a basic publish event
func (a *Analytics) TrackPublishSimple(format string, success bool, wordCount int) {
	startTime := time.Now()

	// Track in session
	metadata := map[string]interface{}{
		"format":      format,
		"word_count":  wordCount,
		"platform":    "local",
		"ai_assisted": false,
	}
	a.sessionManager.StartActivity("publish", "", metadata)

	// Complete the activity
	duration := time.Since(startTime)
	a.sessionManager.CompleteActivity(success, wordCount, duration)

	// ALSO send individual event for debugging
	if a.client != nil {
		event := PublishEvent{
			Format:            format,
			WordCount:         wordCount,
			Platform:          "local",
			Success:           success,
			ContentType:       format,
			AIAssistanceLevel: "none",
			CapturesUsed:      0,
			GenerationTime:    duration,
			QualityScore:      0.5,
		}

		if err := a.client.TrackPublish(event); err != nil {
			log.Printf("Individual publish tracking failed: %v", err)
		} else {
			log.Printf("✅ DEBUG: Individual publish event sent")
		}
	}
}

// TrackPublishDetailed tracks a publish event with full context
func (a *Analytics) TrackPublishDetailed(format string, success bool, wordCount, capturesUsed int, generationTime time.Duration, aiLevel string) {
	// Track in session with detailed metadata
	metadata := map[string]interface{}{
		"format":        format,
		"word_count":    wordCount,
		"platform":      "local",
		"captures_used": capturesUsed,
		"ai_level":      aiLevel,
		"quality_score": calculateQualityScore(wordCount, capturesUsed),
	}
	a.sessionManager.StartActivity("publish", "", metadata)
	a.sessionManager.CompleteActivity(success, wordCount, generationTime)
}

// TrackStatusCheck tracks a status command execution
func (a *Analytics) TrackStatusCheck(totalCaptures, capturesToday, activeProjects, unpublishedCaptures int, dbSizeMB float64) {
	startTime := time.Now()

	// Track in session
	metadata := map[string]interface{}{
		"total_captures":       totalCaptures,
		"captures_today":       capturesToday,
		"active_projects":      activeProjects,
		"unpublished_captures": unpublishedCaptures,
		"database_size_mb":     dbSizeMB,
		"productivity_trend":   calculateProductivityTrend(capturesToday),
	}
	a.sessionManager.StartActivity("status", "", metadata)

	// Complete the activity
	duration := time.Since(startTime)
	a.sessionManager.CompleteActivity(true, totalCaptures, duration)

	// ALSO send individual event for debugging
	if a.client != nil {
		event := StatusEvent{
			TotalCaptures:       totalCaptures,
			CapturesToday:       capturesToday,
			ActiveProjects:      activeProjects,
			ProductivityTrend:   calculateProductivityTrend(capturesToday),
			UnpublishedCaptures: unpublishedCaptures,
			DatabaseSizeMB:      dbSizeMB,
		}

		if err := a.client.TrackStatus(event); err != nil {
			log.Printf("Individual status tracking failed: %v", err)
		} else {
			log.Printf("✅ DEBUG: Individual status event sent")
		}
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
	// Check for session timeout but don't force end
	if a.sessionManager != nil {
		a.sessionManager.CheckSessionTimeout()
	}

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

// GetCurrentSession returns information about the current session
func (a *Analytics) GetCurrentSession() *Session {
	if a.sessionManager == nil {
		return nil
	}
	return a.sessionManager.GetCurrentSession()
}

// EndCurrentSession explicitly ends the current session
func (a *Analytics) EndCurrentSession() {
	if a.sessionManager != nil {
		a.sessionManager.EndSession()
	}
}

// CheckSessionTimeout checks if the current session should timeout
func (a *Analytics) CheckSessionTimeout() {
	if a.sessionManager != nil {
		a.sessionManager.CheckSessionTimeout()
	}
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
