package analytics

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Configuration holds PostHog analytics configuration
type Configuration struct {
	Enabled       bool
	APIKey        string
	Host          string
	ProjectID     string
	Environment   string
	PrivacyMode   string
	LocalFirst    bool
	Debug         bool
	BatchSize     int
	FlushInterval time.Duration
	RetryAttempts int
}

// AnalyticsClient handles PostHog integration for uroboro
type AnalyticsClient struct {
	client *DirectHTTPClient
	config *Configuration
	userID string
}

// PrivacyLevel defines data collection levels
type PrivacyLevel string

const (
	PrivacySafe     PrivacyLevel = "safe"     // Only essential metrics
	PrivacyEnhanced PrivacyLevel = "enhanced" // Development patterns + AI collaboration
	PrivacyFull     PrivacyLevel = "full"     // All non-sensitive analytics
)

// CaptureEvent represents a uroboro capture event
type CaptureEvent struct {
	Content       string
	Project       string
	Tags          []string
	Method        string
	InsightLength int
	CaptureTime   time.Duration
	AIAssisted    bool
	GitContext    *GitContext
	WorkflowStage string
	ContextWeight float64
}

// PublishEvent represents a uroboro publish event
type PublishEvent struct {
	Format            string
	WordCount         int
	Platform          string
	Success           bool
	ContentType       string
	AIAssistanceLevel string
	CapturesUsed      int
	GenerationTime    time.Duration
	QualityScore      float64
}

// StatusEvent represents a uroboro status check event
type StatusEvent struct {
	TotalCaptures       int
	CapturesToday       int
	ActiveProjects      int
	ProductivityTrend   float64
	UnpublishedCaptures int
	DatabaseSizeMB      float64
}

// GitContext holds git-related information
type GitContext struct {
	Branch     string
	IsDirty    bool
	CommitHash string
	RepoName   string
}

// NewAnalyticsClient creates a new PostHog analytics client
func NewAnalyticsClient() (*AnalyticsClient, error) {
	config := loadConfiguration()

	if !config.Enabled {
		return &AnalyticsClient{config: config}, nil
	}

	if config.APIKey == "" {
		return &AnalyticsClient{config: config}, fmt.Errorf("PostHog API key not configured")
	}

	userID := generateUserID()

	// Create direct HTTP client instead of Go SDK
	client := NewDirectHTTPClient(config.APIKey, config.Host, userID, config.Debug)

	return &AnalyticsClient{
		client: client,
		config: config,
		userID: userID,
	}, nil
}

// TrackCapture tracks a uroboro capture event
func (ac *AnalyticsClient) TrackCapture(event CaptureEvent) error {
	if !ac.isEnabled() {
		return nil
	}

	properties := map[string]interface{}{
		// Core capture metrics
		"insight_length":  event.InsightLength,
		"capture_method":  event.Method,
		"capture_time_ms": event.CaptureTime.Milliseconds(),

		// Project context
		"project_name": ac.filterProject(event.Project),
		"project_type": ac.categorizeProject(event.Project),
		"tags":         ac.filterTags(event.Tags),

		// Development context
		"workflow_stage": event.WorkflowStage,
		"context_weight": event.ContextWeight,
		"ai_assisted":    event.AIAssisted,

		// System context
		"os":             runtime.GOOS,
		"uroboro_method": "capture",
		"timestamp":      time.Now().UTC(),
	}

	// Add git context if available
	if event.GitContext != nil {
		properties["git_branch"] = ac.filterGitBranch(event.GitContext.Branch)
		properties["git_is_dirty"] = event.GitContext.IsDirty
		properties["git_repo"] = ac.filterRepoName(event.GitContext.RepoName)
	}

	// Add system metrics
	ac.addSystemMetrics(properties)

	return ac.captureEvent("uroboro_capture", properties)
}

// TrackPublish tracks a uroboro publish event
func (ac *AnalyticsClient) TrackPublish(event PublishEvent) error {
	if !ac.isEnabled() {
		return nil
	}

	properties := map[string]interface{}{
		// Publication metrics
		"format":              event.Format,
		"word_count":          event.WordCount,
		"publish_platform":    event.Platform,
		"publish_success":     event.Success,
		"content_type":        event.ContentType,
		"ai_assistance_level": event.AIAssistanceLevel,
		"captures_used":       event.CapturesUsed,
		"generation_time_ms":  event.GenerationTime.Milliseconds(),
		"quality_score":       event.QualityScore,

		// System context
		"uroboro_method": "publish",
		"timestamp":      time.Now().UTC(),
	}

	ac.addSystemMetrics(properties)

	return ac.captureEvent("uroboro_publish", properties)
}

// TrackStatus tracks a uroboro status check event
func (ac *AnalyticsClient) TrackStatus(event StatusEvent) error {
	if !ac.isEnabled() {
		return nil
	}

	properties := map[string]interface{}{
		// Status metrics
		"total_captures":       event.TotalCaptures,
		"captures_today":       event.CapturesToday,
		"active_projects":      event.ActiveProjects,
		"productivity_trend":   event.ProductivityTrend,
		"unpublished_captures": event.UnpublishedCaptures,
		"database_size_mb":     event.DatabaseSizeMB,

		// System context
		"uroboro_method": "status",
		"timestamp":      time.Now().UTC(),
	}

	ac.addSystemMetrics(properties)

	return ac.captureEvent("uroboro_status", properties)
}

// TrackAICollaboration tracks AI collaboration sessions
func (ac *AnalyticsClient) TrackAICollaboration(duration time.Duration, queries int, satisfaction float64) error {
	if !ac.isEnabled() {
		return nil
	}

	properties := map[string]interface{}{
		"session_duration_minutes": duration.Minutes(),
		"queries_sent":             queries,
		"human_satisfaction":       satisfaction,
		"collaboration_type":       "ai_enhanced_capture",
		"ai_model":                 "context_aware",
		"timestamp":                time.Now().UTC(),
	}

	return ac.captureEvent("ai_collaboration_session", properties)
}

// TrackWorkflowTransition tracks workflow state changes
func (ac *AnalyticsClient) TrackWorkflowTransition(fromState, toState, trigger string, duration time.Duration) error {
	if !ac.isEnabled() {
		return nil
	}

	properties := map[string]interface{}{
		"from_state":             fromState,
		"to_state":               toState,
		"transition_trigger":     trigger,
		"state_duration_minutes": duration.Minutes(),
		"context_preserved":      true,
		"timestamp":              time.Now().UTC(),
	}

	return ac.captureEvent("workflow_transition", properties)
}

// Close gracefully shuts down the analytics client
func (ac *AnalyticsClient) Close() error {
	// Direct HTTP client doesn't need explicit closing
	return nil
}

// Helper methods

func (ac *AnalyticsClient) isEnabled() bool {
	return ac.config.Enabled && ac.client != nil
}

func (ac *AnalyticsClient) captureEvent(eventName string, properties map[string]interface{}) error {
	if !ac.isEnabled() {
		log.Printf("⚠️  DEBUG: Analytics not enabled, skipping event: %s", eventName)
		return nil
	}

	log.Printf("🔍 DEBUG: Capturing event: %s", eventName)
	log.Printf("🔍 DEBUG: User ID: %s", ac.userID)
	log.Printf("🔍 DEBUG: Raw properties count: %d", len(properties))

	// Apply privacy filtering
	filteredProperties := ac.applyPrivacyFilter(properties)
	log.Printf("🔍 DEBUG: Filtered properties count: %d", len(filteredProperties))

	// Add common properties
	filteredProperties["qry_integration"] = "uroboro_posthog"
	filteredProperties["qry_version"] = "1.0.0"

	log.Printf("🔍 DEBUG: Final properties: %+v", filteredProperties)

	// Use direct HTTP client instead of Go SDK
	err := ac.client.SendEvent(eventName, filteredProperties)

	if err != nil {
		log.Printf("❌ DEBUG: Failed to send event %s: %v", eventName, err)
	} else {
		log.Printf("✅ DEBUG: Successfully sent event: %s", eventName)
	}

	return err
}

func (ac *AnalyticsClient) applyPrivacyFilter(properties map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})

	privacyLevel := PrivacyLevel(ac.config.PrivacyMode)

	for key, value := range properties {
		switch privacyLevel {
		case PrivacySafe:
			// Only include essential metrics
			if ac.isEssentialProperty(key) {
				filtered[key] = value
			}
		case PrivacyEnhanced:
			// Include development patterns but anonymize sensitive data
			if !ac.isSensitiveProperty(key) {
				filtered[key] = ac.anonymizeIfNeeded(key, value)
			}
		case PrivacyFull:
			// Include all non-sensitive data
			if !ac.isSensitiveProperty(key) {
				filtered[key] = value
			}
		}
	}

	return filtered
}

func (ac *AnalyticsClient) isEssentialProperty(key string) bool {
	essentialProps := []string{
		"uroboro_method", "timestamp", "capture_method", "publish_success",
		"ai_assisted", "workflow_stage", "format", "os",
	}

	for _, prop := range essentialProps {
		if key == prop {
			return true
		}
	}
	return false
}

func (ac *AnalyticsClient) isSensitiveProperty(key string) bool {
	sensitiveProps := []string{
		"content", "full_path", "absolute_path", "user_name", "email",
		"api_key", "token", "password", "secret",
	}

	keyLower := strings.ToLower(key)
	for _, prop := range sensitiveProps {
		if strings.Contains(keyLower, prop) {
			return true
		}
	}
	return false
}

func (ac *AnalyticsClient) anonymizeIfNeeded(key string, value interface{}) interface{} {
	if strings.Contains(strings.ToLower(key), "path") {
		if strVal, ok := value.(string); ok {
			return ac.anonymizePath(strVal)
		}
	}
	return value
}

func (ac *AnalyticsClient) anonymizePath(path string) string {
	// Convert absolute paths to relative patterns
	if strings.Contains(path, "/home/") {
		return "~/..." + filepath.Base(path)
	}
	if strings.Contains(path, "C:\\Users\\") {
		return "~\\..." + filepath.Base(path)
	}
	return filepath.Base(path)
}

func (ac *AnalyticsClient) filterProject(project string) string {
	if ac.config.PrivacyMode == "safe" {
		return "private_project"
	}
	// Remove potentially sensitive project names
	if strings.Contains(strings.ToLower(project), "private") ||
		strings.Contains(strings.ToLower(project), "secret") ||
		strings.Contains(strings.ToLower(project), "internal") {
		return "private_project"
	}
	return project
}

func (ac *AnalyticsClient) categorizeProject(project string) string {
	if strings.Contains(strings.ToLower(project), "qry") {
		return "qry_ecosystem"
	}
	if strings.Contains(strings.ToLower(project), "uroboro") {
		return "uroboro_development"
	}
	if strings.Contains(strings.ToLower(project), "posthog") {
		return "integration_development"
	}
	return "general_development"
}

func (ac *AnalyticsClient) filterTags(tags []string) []string {
	if ac.config.PrivacyMode == "safe" {
		return []string{"development"}
	}

	filtered := make([]string, 0, len(tags))
	for _, tag := range tags {
		if !ac.isSensitiveTag(tag) {
			filtered = append(filtered, tag)
		}
	}
	return filtered
}

func (ac *AnalyticsClient) isSensitiveTag(tag string) bool {
	sensitivePatterns := []string{"private", "secret", "internal", "confidential"}
	tagLower := strings.ToLower(tag)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(tagLower, pattern) {
			return true
		}
	}
	return false
}

func (ac *AnalyticsClient) filterGitBranch(branch string) string {
	if ac.config.PrivacyMode == "safe" {
		return "development_branch"
	}
	// Remove potentially sensitive branch names
	if strings.Contains(strings.ToLower(branch), "private") ||
		strings.Contains(strings.ToLower(branch), "secret") {
		return "private_branch"
	}
	return branch
}

func (ac *AnalyticsClient) filterRepoName(repo string) string {
	if ac.config.PrivacyMode == "safe" {
		return "private_repo"
	}
	return filepath.Base(repo) // Only return the repo name, not full path
}

func (ac *AnalyticsClient) addSystemMetrics(properties map[string]interface{}) {
	properties["go_version"] = runtime.Version()
	properties["num_cpu"] = runtime.NumCPU()
	properties["arch"] = runtime.GOARCH
}

// GetGitContext extracts git information from the current directory
func GetGitContext() *GitContext {
	ctx := context.Background()

	// Get current branch
	branchCmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return nil
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Check if repo is dirty
	statusCmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	isDirty := err == nil && len(statusOutput) > 0

	// Get commit hash
	commitCmd := exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD")
	commitOutput, err := commitCmd.Output()
	var commitHash string
	if err == nil {
		commitHash = strings.TrimSpace(string(commitOutput))
	}

	// Get repo name
	repoCmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	repoOutput, err := repoCmd.Output()
	var repoName string
	if err == nil {
		repoPath := strings.TrimSpace(string(repoOutput))
		repoName = filepath.Base(repoPath)
	}

	return &GitContext{
		Branch:     branch,
		IsDirty:    isDirty,
		CommitHash: commitHash,
		RepoName:   repoName,
	}
}

// loadConfiguration loads PostHog configuration from environment variables
func loadConfiguration() *Configuration {
	config := &Configuration{
		Enabled:       getEnvBool("QRY_POSTHOG_ENABLED", false),
		APIKey:        os.Getenv("POSTHOG_API_KEY"),
		Host:          getEnvString("POSTHOG_HOST", "https://us.posthog.com"),
		ProjectID:     os.Getenv("POSTHOG_PROJECT_ID"),
		Environment:   getEnvString("QRY_POSTHOG_ENVIRONMENT", "development"),
		PrivacyMode:   getEnvString("QRY_PRIVACY_MODE", "enhanced"),
		LocalFirst:    getEnvBool("QRY_LOCAL_FIRST", true),
		Debug:         getEnvBool("QRY_DEBUG_ANALYTICS", false),
		BatchSize:     getEnvInt("POSTHOG_BATCH_SIZE", 100),
		FlushInterval: getEnvDuration("POSTHOG_FLUSH_INTERVAL", 10*time.Second),
		RetryAttempts: getEnvInt("POSTHOG_RETRY_ATTEMPTS", 3),
	}

	// Debug logging for configuration
	log.Printf("🔍 DEBUG: PostHog configuration loaded:")
	log.Printf("🔍 DEBUG: Enabled=%v", config.Enabled)
	log.Printf("🔍 DEBUG: APIKey=%s", maskAPIKey(config.APIKey))
	log.Printf("🔍 DEBUG: Host=%s", config.Host)
	log.Printf("🔍 DEBUG: Environment=%s", config.Environment)
	log.Printf("🔍 DEBUG: PrivacyMode=%s", config.PrivacyMode)

	return config
}

func generateUserID() string {
	// Generate a stable user ID for this installation
	hostname, _ := os.Hostname()
	return fmt.Sprintf("uroboro_user_%s", hostname)
}

func maskAPIKey(key string) string {
	if len(key) < 8 {
		return "***"
	}
	return key[:4] + "***" + key[len(key)-4:]
}

// Environment variable helpers

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
