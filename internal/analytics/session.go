package analytics

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Session represents a development work session
type Session struct {
	ID           string                 `json:"id"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	LastActivity time.Time              `json:"last_activity"`
	Project      string                 `json:"project"`
	GitContext   *GitContext            `json:"git_context,omitempty"`
	Activities   []SessionActivity      `json:"activities"`
	Context      map[string]interface{} `json:"context"`
	WorkingDir   string                 `json:"working_dir"`
}

// SessionActivity represents an action within a session
type SessionActivity struct {
	Type        string                 `json:"type"`
	Timestamp   time.Time              `json:"timestamp"`
	Project     string                 `json:"project"`
	Metadata    map[string]interface{} `json:"metadata"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Success     bool                   `json:"success"`
	ContentSize int                    `json:"content_size,omitempty"`
}

// SessionManager handles session lifecycle and analytics
type SessionManager struct {
	currentSession   *Session
	sessionTimeout   time.Duration
	analyticsClient  *AnalyticsClient
	sessionFile      string
	mutex            sync.RWMutex
	autoSaveInterval time.Duration
	lastSave         time.Time
}

// SessionSummary represents data sent to PostHog
type SessionSummary struct {
	SessionID         string        `json:"session_id"`
	Duration          time.Duration `json:"duration"`
	ActivityCount     int           `json:"activity_count"`
	CaptureCount      int           `json:"capture_count"`
	PublishCount      int           `json:"publish_count"`
	StatusCheckCount  int           `json:"status_check_count"`
	ProjectsWorkedOn  []string      `json:"projects_worked_on"`
	FlowQuality       float64       `json:"flow_quality"`
	InterruptionCount int           `json:"interruption_count"`
	ContextSwitches   int           `json:"context_switches"`
	SessionType       string        `json:"session_type"`
	GitBranches       []string      `json:"git_branches"`
	TotalContentSize  int           `json:"total_content_size"`
	AverageTaskTime   time.Duration `json:"average_task_time"`
	ProductivityScore float64       `json:"productivity_score"`
}

// NewSessionManager creates a new session manager
func NewSessionManager(analyticsClient *AnalyticsClient) *SessionManager {
	sessionFile := getSessionFilePath()

	sm := &SessionManager{
		sessionTimeout:   15 * time.Minute, // 15 minutes of inactivity ends session
		sessionFile:      sessionFile,
		autoSaveInterval: 2 * time.Minute, // Save session state every 2 minutes
		analyticsClient:  analyticsClient,
	}

	if analyticsClient != nil {
		log.Printf("🔍 DEBUG: SessionManager initialized with analytics client")
	} else {
		log.Printf("⚠️  DEBUG: SessionManager initialized without analytics client")
	}

	// Try to restore previous session
	sm.restoreSession()

	return sm
}

// StartActivity records a new activity and manages session lifecycle
func (sm *SessionManager) StartActivity(activityType, project string, metadata map[string]interface{}) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()

	// Check if we need to start a new session
	if sm.shouldStartNewSession(now, project) {
		sm.endCurrentSession()
		sm.startNewSession(now, project)
	}

	// Add activity to current session
	activity := SessionActivity{
		Type:      activityType,
		Timestamp: now,
		Project:   project,
		Metadata:  metadata,
		Success:   true, // Will be updated by CompleteActivity if needed
	}

	if sm.currentSession != nil {
		sm.currentSession.Activities = append(sm.currentSession.Activities, activity)
		sm.currentSession.LastActivity = now

		// Update session project if it changed
		if sm.currentSession.Project == "" || sm.currentSession.Project != project {
			sm.currentSession.Project = project
		}
	}

	// Auto-save session state
	sm.autoSaveSession()
}

// CompleteActivity updates the last activity with completion info
func (sm *SessionManager) CompleteActivity(success bool, contentSize int, duration time.Duration) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.currentSession != nil && len(sm.currentSession.Activities) > 0 {
		lastActivity := &sm.currentSession.Activities[len(sm.currentSession.Activities)-1]
		lastActivity.Success = success
		lastActivity.ContentSize = contentSize
		lastActivity.Duration = duration
		sm.currentSession.LastActivity = time.Now()
	}
}

// EndSession explicitly ends the current session
func (sm *SessionManager) EndSession() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.endCurrentSession()
}

// GetCurrentSession returns the current session (read-only)
func (sm *SessionManager) GetCurrentSession() *Session {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if sm.currentSession == nil {
		return nil
	}

	// Return a copy to avoid race conditions
	sessionCopy := *sm.currentSession
	return &sessionCopy
}

// CheckSessionTimeout checks if current session should timeout
func (sm *SessionManager) CheckSessionTimeout() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.currentSession != nil {
		if time.Since(sm.currentSession.LastActivity) > sm.sessionTimeout {
			sm.endCurrentSession()
		}
	}
}

// Private methods

func (sm *SessionManager) shouldStartNewSession(now time.Time, project string) bool {
	// No current session
	if sm.currentSession == nil {
		return true
	}

	// Session timed out
	if now.Sub(sm.currentSession.LastActivity) > sm.sessionTimeout {
		return true
	}

	// Project changed significantly (context switch)
	if sm.currentSession.Project != "" && sm.currentSession.Project != project {
		// Allow some flexibility for related projects
		if !sm.isRelatedProject(sm.currentSession.Project, project) {
			return true
		}
	}

	// Git context changed (different branch/repo)
	currentGit := GetGitContext()
	if sm.currentSession.GitContext != nil && currentGit != nil {
		if sm.currentSession.GitContext.Branch != currentGit.Branch ||
			sm.currentSession.GitContext.RepoName != currentGit.RepoName {
			return true
		}
	}

	return false
}

func (sm *SessionManager) startNewSession(now time.Time, project string) {
	sessionID := fmt.Sprintf("session_%d", now.Unix())

	wd, _ := os.Getwd()

	sm.currentSession = &Session{
		ID:           sessionID,
		StartTime:    now,
		LastActivity: now,
		Project:      project,
		GitContext:   GetGitContext(),
		Activities:   make([]SessionActivity, 0),
		Context:      make(map[string]interface{}),
		WorkingDir:   wd,
	}

	log.Printf("🎯 Started new development session: %s (project: %s)", sessionID, project)
}

func (sm *SessionManager) endCurrentSession() {
	if sm.currentSession == nil {
		return
	}

	now := time.Now()
	sm.currentSession.EndTime = &now

	// Calculate session summary and send to analytics
	if sm.analyticsClient != nil && sm.analyticsClient.isEnabled() {
		log.Printf("🔍 DEBUG: Generating session summary for analytics...")
		summary := sm.generateSessionSummary()
		log.Printf("🔍 DEBUG: Session summary: duration=%.1fm, activities=%d, captures=%d",
			summary.Duration.Minutes(), summary.ActivityCount, summary.CaptureCount)
		sm.sendSessionAnalytics(summary)
	} else {
		log.Printf("⚠️  DEBUG: Session analytics not sent - client disabled or nil")
	}

	// Save final session state
	sm.saveSessionToFile()

	duration := now.Sub(sm.currentSession.StartTime)
	activityCount := len(sm.currentSession.Activities)

	log.Printf("📊 Ended session: %s (duration: %v, activities: %d)",
		sm.currentSession.ID, duration.Round(time.Minute), activityCount)

	sm.currentSession = nil
	sm.clearSessionFile()
}

func (sm *SessionManager) generateSessionSummary() SessionSummary {
	session := sm.currentSession
	if session == nil {
		return SessionSummary{}
	}

	duration := time.Since(session.StartTime)
	if session.EndTime != nil {
		duration = session.EndTime.Sub(session.StartTime)
	}

	summary := SessionSummary{
		SessionID:     session.ID,
		Duration:      duration,
		ActivityCount: len(session.Activities),
	}

	// Analyze activities
	projects := make(map[string]bool)
	branches := make(map[string]bool)
	totalContentSize := 0
	var taskDurations []time.Duration

	for _, activity := range session.Activities {
		projects[activity.Project] = true
		totalContentSize += activity.ContentSize

		if activity.Duration > 0 {
			taskDurations = append(taskDurations, activity.Duration)
		}

		switch activity.Type {
		case "capture":
			summary.CaptureCount++
		case "publish":
			summary.PublishCount++
		case "status":
			summary.StatusCheckCount++
		}
	}

	// Convert maps to slices
	for project := range projects {
		summary.ProjectsWorkedOn = append(summary.ProjectsWorkedOn, project)
	}
	for branch := range branches {
		summary.GitBranches = append(summary.GitBranches, branch)
	}

	summary.TotalContentSize = totalContentSize
	summary.ContextSwitches = len(projects) - 1 // Project switches

	// Calculate metrics
	summary.FlowQuality = sm.calculateFlowQuality(session)
	summary.InterruptionCount = sm.calculateInterruptions(session)
	summary.SessionType = sm.determineSessionType(session)
	summary.ProductivityScore = sm.calculateProductivityScore(session)

	if len(taskDurations) > 0 {
		var total time.Duration
		for _, d := range taskDurations {
			total += d
		}
		summary.AverageTaskTime = total / time.Duration(len(taskDurations))
	}

	return summary
}

func (sm *SessionManager) calculateFlowQuality(session *Session) float64 {
	if len(session.Activities) < 2 {
		return 0.5
	}

	// Higher quality = consistent activity with minimal gaps
	gaps := make([]time.Duration, 0)
	for i := 1; i < len(session.Activities); i++ {
		gap := session.Activities[i].Timestamp.Sub(session.Activities[i-1].Timestamp)
		gaps = append(gaps, gap)
	}

	// Calculate average gap
	var totalGap time.Duration
	for _, gap := range gaps {
		totalGap += gap
	}
	avgGap := totalGap / time.Duration(len(gaps))

	// Flow quality decreases with longer gaps
	// 0-5 min gaps = high quality (0.8-1.0)
	// 5-15 min gaps = medium quality (0.4-0.8)
	// 15+ min gaps = low quality (0.0-0.4)

	if avgGap < 5*time.Minute {
		return 0.8 + (5*time.Minute-avgGap).Seconds()/(5*60*5) // 0.8-1.0 range
	} else if avgGap < 15*time.Minute {
		return 0.4 + (15*time.Minute-avgGap).Seconds()/(10*60*2.5) // 0.4-0.8 range
	} else {
		return 0.2 // Low flow quality for very long gaps
	}
}

func (sm *SessionManager) calculateInterruptions(session *Session) int {
	// Count gaps longer than 10 minutes as interruptions
	interruptions := 0
	for i := 1; i < len(session.Activities); i++ {
		gap := session.Activities[i].Timestamp.Sub(session.Activities[i-1].Timestamp)
		if gap > 10*time.Minute {
			interruptions++
		}
	}
	return interruptions
}

func (sm *SessionManager) determineSessionType(session *Session) string {
	duration := time.Since(session.StartTime)
	activityCount := len(session.Activities)
	captureRatio := float64(sm.countActivityType(session, "capture")) / float64(activityCount)

	if duration < 30*time.Minute {
		return "quick_session"
	} else if duration > 2*time.Hour && captureRatio > 0.7 {
		return "deep_work"
	} else if captureRatio > 0.8 {
		return "capture_heavy"
	} else if sm.countActivityType(session, "publish") > 0 {
		return "content_creation"
	} else {
		return "mixed_work"
	}
}

func (sm *SessionManager) calculateProductivityScore(session *Session) float64 {
	// Combine multiple factors for productivity score
	activityCount := len(session.Activities)
	duration := time.Since(session.StartTime).Hours()

	// Base score on activity density
	activityRate := float64(activityCount) / duration

	// Normalize to 0-1 scale (assuming 4 activities/hour is highly productive)
	score := activityRate / 4.0
	if score > 1.0 {
		score = 1.0
	}

	// Adjust based on flow quality
	flowQuality := sm.calculateFlowQuality(session)
	score = (score + flowQuality) / 2.0

	// Boost for content creation
	if sm.countActivityType(session, "publish") > 0 {
		score += 0.1
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (sm *SessionManager) countActivityType(session *Session, activityType string) int {
	count := 0
	for _, activity := range session.Activities {
		if activity.Type == activityType {
			count++
		}
	}
	return count
}

func (sm *SessionManager) sendSessionAnalytics(summary SessionSummary) {
	properties := map[string]interface{}{
		"session_id":            summary.SessionID,
		"duration_minutes":      summary.Duration.Minutes(),
		"activity_count":        summary.ActivityCount,
		"capture_count":         summary.CaptureCount,
		"publish_count":         summary.PublishCount,
		"status_check_count":    summary.StatusCheckCount,
		"projects_worked_on":    summary.ProjectsWorkedOn,
		"flow_quality":          summary.FlowQuality,
		"interruption_count":    summary.InterruptionCount,
		"context_switches":      summary.ContextSwitches,
		"session_type":          summary.SessionType,
		"git_branches":          summary.GitBranches,
		"total_content_size":    summary.TotalContentSize,
		"average_task_time_min": summary.AverageTaskTime.Minutes(),
		"productivity_score":    summary.ProductivityScore,
		"timestamp":             time.Now().UTC(),
	}

	log.Printf("🔍 DEBUG: Sending session analytics to PostHog...")
	log.Printf("🔍 DEBUG: Event: uroboro_session_complete")
	log.Printf("🔍 DEBUG: Properties: %+v", properties)

	if sm.analyticsClient == nil {
		log.Printf("❌ DEBUG: Analytics client is nil, cannot send session analytics")
		return
	}

	if !sm.analyticsClient.isEnabled() {
		log.Printf("❌ DEBUG: Analytics client is disabled")
		return
	}

	if err := sm.analyticsClient.captureEvent("uroboro_session_complete", properties); err != nil {
		log.Printf("❌ Failed to send session analytics: %v", err)
	} else {
		log.Printf("✅ DEBUG: Session analytics sent successfully to PostHog")
	}
}

func (sm *SessionManager) isRelatedProject(project1, project2 string) bool {
	// Consider projects related if they share common prefixes
	// or are both in the QRY ecosystem
	qryProjects := []string{"uroboro", "doggowoof", "qry", "deskhog", "wherewasi"}

	project1InQry := false
	project2InQry := false

	for _, qryProject := range qryProjects {
		if project1 == qryProject {
			project1InQry = true
		}
		if project2 == qryProject {
			project2InQry = true
		}
	}

	return project1InQry && project2InQry
}

func (sm *SessionManager) autoSaveSession() {
	if time.Since(sm.lastSave) > sm.autoSaveInterval {
		sm.saveSessionToFile()
		sm.lastSave = time.Now()
	}
}

func (sm *SessionManager) saveSessionToFile() {
	if sm.currentSession == nil {
		return
	}

	data, err := json.MarshalIndent(sm.currentSession, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal session: %v", err)
		return
	}

	if err := os.WriteFile(sm.sessionFile, data, 0644); err != nil {
		log.Printf("Failed to save session: %v", err)
	}
}

func (sm *SessionManager) restoreSession() {
	data, err := os.ReadFile(sm.sessionFile)
	if err != nil {
		return // No existing session file
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		log.Printf("Failed to restore session: %v", err)
		return
	}

	// Only restore if session is recent (within timeout period)
	if time.Since(session.LastActivity) <= sm.sessionTimeout {
		sm.currentSession = &session
		log.Printf("🔄 Restored session: %s", session.ID)
	} else {
		// Session expired, clean up
		sm.clearSessionFile()
	}
}

func (sm *SessionManager) clearSessionFile() {
	os.Remove(sm.sessionFile)
}

func getSessionFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/uroboro_session.json"
	}

	return filepath.Join(homeDir, ".uroboro", "current_session.json")
}
