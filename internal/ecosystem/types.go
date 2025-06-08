package ecosystem

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Tool names for the QRY ecosystem
const (
	ToolWherewasi  = "wherewasi"
	ToolUroboro    = "uroboro"
	ToolExaminator = "examinator"
	ToolQryAI      = "qryai"
	ToolDoggowoof  = "doggowoof"
	ToolQomoboro   = "qomoboro"
)

// Message types for cross-tool communication
const (
	MessageTypeCapture           = "capture"
	MessageTypeContextUpdate     = "context_update"
	MessageTypeFlashcardRequest  = "flashcard_request"
	MessageTypeStudySession      = "study_session"
	MessageTypeProjectActivity   = "project_activity"
	MessageTypeInsight           = "insight"
	MessageTypeAlert             = "alert"
	MessageTypePomodoroComplete  = "pomodoro_complete"
)

// Insight types for ecosystem intelligence
const (
	InsightTypeStudyRecommendation = "study_recommendation"
	InsightTypeContentSynthesis    = "content_synthesis"
	InsightTypeProductivityPattern = "productivity_pattern"
	InsightTypeProjectConnection   = "project_connection"
	InsightTypeTimeOptimization    = "time_optimization"
)

// ToolMessage represents a cross-tool communication message
type ToolMessage struct {
	ID          int64      `json:"id"`
	FromTool    string     `json:"from_tool"`
	ToTool      string     `json:"to_tool"`
	MessageType string     `json:"message_type"`
	Data        string     `json:"data"`
	Processed   bool       `json:"processed"`
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at"`
}

// Project represents a tracked project in the ecosystem
type Project struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Description  *string    `json:"description"`
	Path         string     `json:"path"`
	GitRepo      bool       `json:"git_repo"`
	LastActivity *time.Time `json:"last_activity"`
	PrimaryTool  *string    `json:"primary_tool"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// EcosystemInsight represents an AI-generated insight for cross-tool optimization
type EcosystemInsight struct {
	ID         int64      `json:"id"`
	Type       string     `json:"insight_type"`
	SourceTool string     `json:"source_tool"`
	TargetTool *string    `json:"target_tool"`
	Project    *string    `json:"project"`
	Confidence float64    `json:"confidence"`
	Data       string     `json:"data"` // JSON
	Applied    bool       `json:"applied"`
	CreatedAt  time.Time  `json:"created_at"`
	AppliedAt  *time.Time `json:"applied_at"`
}

// ContextSession represents a wherewasi context capture
type ContextSession struct {
	ID          int64     `json:"id"`
	Project     string    `json:"project"`
	Timestamp   time.Time `json:"timestamp"`
	ContextData string    `json:"context_data"`
	SessionInfo string    `json:"session_info"`
	Keywords    string    `json:"keywords"`
	GitBranch   *string   `json:"git_branch"`
	GitCommit   *string   `json:"git_commit"`
	CreatedAt   time.Time `json:"created_at"`
}

// Capture represents a uroboro content capture
type Capture struct {
	ID               int64      `json:"id"`
	Timestamp        time.Time  `json:"timestamp"`
	Content          string     `json:"content"`
	Project          *string    `json:"project"`
	Tags             *string    `json:"tags"`
	SourceTool       string     `json:"source_tool"`
	Metadata         *string    `json:"metadata"`
	ContextSessionID *int64     `json:"context_session_id"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Publication represents a uroboro published content
type Publication struct {
	ID             int64     `json:"id"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Format         string    `json:"format"`
	Type           string    `json:"type"`
	SourceCaptures string    `json:"source_captures"` // JSON array of capture IDs
	Project        string    `json:"project"`
	TargetPath     string    `json:"target_path"`
	CreatedAt      time.Time `json:"created_at"`
}

// Flashcard represents an examinator flashcard
type Flashcard struct {
	ID               int64      `json:"id"`
	Question         string     `json:"question"`
	Answer           string     `json:"answer"`
	Category         *string    `json:"category"`
	Difficulty       int        `json:"difficulty"`
	SourceCaptureID  *int64     `json:"source_capture_id"`
	ContextSessionID *int64     `json:"context_session_id"`
	Project          *string    `json:"project"`
	CreatedAt        time.Time  `json:"created_at"`
	LastReviewed     *time.Time `json:"last_reviewed"`
	NextReview       *time.Time `json:"next_review"`
	EaseFactor       float64    `json:"ease_factor"`
	ReviewCount      int        `json:"review_count"`
	CorrectStreak    int        `json:"correct_streak"`
}

// Message data structures for specific message types

// CaptureMessageData represents data for capture messages
type CaptureMessageData struct {
	Content          string  `json:"content"`
	Project          string  `json:"project"`
	Tags             string  `json:"tags"`
	ContextSessionID *int64  `json:"context_session_id,omitempty"`
	Metadata         *string `json:"metadata,omitempty"`
}

// ContextUpdateMessageData represents data for context update messages
type ContextUpdateMessageData struct {
	Project     string  `json:"project"`
	ContextData string  `json:"context_data"`
	SessionInfo string  `json:"session_info"`
	Keywords    string  `json:"keywords"`
	GitBranch   *string `json:"git_branch,omitempty"`
	GitCommit   *string `json:"git_commit,omitempty"`
}

// FlashcardRequestMessageData represents data for flashcard generation requests
type FlashcardRequestMessageData struct {
	Project          string   `json:"project"`
	SourceCaptureIDs []int64  `json:"source_capture_ids"`
	Category         *string  `json:"category,omitempty"`
	Difficulty       int      `json:"difficulty"`
	ContextSessionID *int64   `json:"context_session_id,omitempty"`
}

// StudySessionMessageData represents data for study session tracking
type StudySessionMessageData struct {
	Project              string `json:"project"`
	FlashcardsReviewed   int    `json:"flashcards_reviewed"`
	CorrectAnswers       int    `json:"correct_answers"`
	DurationMinutes      int    `json:"duration_minutes"`
	ContextSessionID     *int64 `json:"context_session_id,omitempty"`
}

// ProjectActivityMessageData represents data for project activity tracking
type ProjectActivityMessageData struct {
	Project     string `json:"project"`
	Activity    string `json:"activity"`
	Tool        string `json:"tool"`
	GitBranch   *string `json:"git_branch,omitempty"`
	GitCommit   *string `json:"git_commit,omitempty"`
}

// Utility functions

// NewToolMessage creates a new tool message with proper validation
func NewToolMessage(fromTool, toTool, messageType string, data interface{}) (*ToolMessage, error) {
	if err := ValidateToolName(fromTool); err != nil {
		return nil, fmt.Errorf("invalid from_tool: %w", err)
	}
	if err := ValidateToolName(toTool); err != nil {
		return nil, fmt.Errorf("invalid to_tool: %w", err)
	}
	if err := ValidateMessageType(messageType); err != nil {
		return nil, fmt.Errorf("invalid message_type: %w", err)
	}

	// Serialize data to JSON
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message data: %w", err)
	}

	return &ToolMessage{
		FromTool:    fromTool,
		ToTool:      toTool,
		MessageType: messageType,
		Data:        string(dataJSON),
		Processed:   false,
		CreatedAt:   time.Now(),
	}, nil
}

// ParseMessageData unmarshals message data into the appropriate struct
func (tm *ToolMessage) ParseMessageData(dest interface{}) error {
	return json.Unmarshal([]byte(tm.Data), dest)
}

// IsValid checks if a tool message has valid fields
func (tm *ToolMessage) IsValid() error {
	if err := ValidateToolName(tm.FromTool); err != nil {
		return fmt.Errorf("invalid from_tool: %w", err)
	}
	if err := ValidateToolName(tm.ToTool); err != nil {
		return fmt.Errorf("invalid to_tool: %w", err)
	}
	if err := ValidateMessageType(tm.MessageType); err != nil {
		return fmt.Errorf("invalid message_type: %w", err)
	}
	if tm.Data == "" {
		return fmt.Errorf("message data cannot be empty")
	}
	return nil
}

// ValidateToolName checks if a tool name is valid
func ValidateToolName(toolName string) error {
	validTools := []string{
		ToolWherewasi, ToolUroboro, ToolExaminator, 
		ToolQryAI, ToolDoggowoof, ToolQomoboro,
	}
	
	for _, valid := range validTools {
		if toolName == valid {
			return nil
		}
	}
	
	return fmt.Errorf("unknown tool name: %s (valid: %s)", 
		toolName, strings.Join(validTools, ", "))
}

// ValidateMessageType checks if a message type is valid
func ValidateMessageType(messageType string) error {
	validTypes := []string{
		MessageTypeCapture, MessageTypeContextUpdate, MessageTypeFlashcardRequest,
		MessageTypeStudySession, MessageTypeProjectActivity, MessageTypeInsight,
		MessageTypeAlert, MessageTypePomodoroComplete,
	}
	
	for _, valid := range validTypes {
		if messageType == valid {
			return nil
		}
	}
	
	return fmt.Errorf("unknown message type: %s (valid: %s)", 
		messageType, strings.Join(validTypes, ", "))
}

// ValidateInsightType checks if an insight type is valid
func ValidateInsightType(insightType string) error {
	validTypes := []string{
		InsightTypeStudyRecommendation, InsightTypeContentSynthesis,
		InsightTypeProductivityPattern, InsightTypeProjectConnection,
		InsightTypeTimeOptimization,
	}
	
	for _, valid := range validTypes {
		if insightType == valid {
			return nil
		}
	}
	
	return fmt.Errorf("unknown insight type: %s (valid: %s)", 
		insightType, strings.Join(validTypes, ", "))
}

// GetAllToolNames returns a slice of all valid tool names
func GetAllToolNames() []string {
	return []string{
		ToolWherewasi, ToolUroboro, ToolExaminator,
		ToolQryAI, ToolDoggowoof, ToolQomoboro,
	}
}

// GetAllMessageTypes returns a slice of all valid message types
func GetAllMessageTypes() []string {
	return []string{
		MessageTypeCapture, MessageTypeContextUpdate, MessageTypeFlashcardRequest,
		MessageTypeStudySession, MessageTypeProjectActivity, MessageTypeInsight,
		MessageTypeAlert, MessageTypePomodoroComplete,
	}
}

// GetAllInsightTypes returns a slice of all valid insight types
func GetAllInsightTypes() []string {
	return []string{
		InsightTypeStudyRecommendation, InsightTypeContentSynthesis,
		InsightTypeProductivityPattern, InsightTypeProjectConnection,
		InsightTypeTimeOptimization,
	}
}

// Helper functions for working with projects

// SanitizeProjectName cleans a project name for database storage
func SanitizeProjectName(name string) string {
	// Remove leading/trailing whitespace
	name = strings.TrimSpace(name)
	
	// Replace problematic characters
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	
	// Limit length
	if len(name) > 255 {
		name = name[:255]
	}
	
	return name
}

// ExtractProjectFromPath attempts to extract a project name from a file path
func ExtractProjectFromPath(path string) string {
	// Get the last two path components (assuming /path/to/project/file.ext)
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) >= 2 {
		return SanitizeProjectName(parts[len(parts)-2])
	} else if len(parts) == 1 {
		return SanitizeProjectName(parts[0])
	}
	return "unknown"
}

// FormatDuration formats a duration in milliseconds to a human-readable string
func FormatDuration(durationMs int) string {
	duration := time.Duration(durationMs) * time.Millisecond
	
	if duration < time.Second {
		return fmt.Sprintf("%dms", durationMs)
	} else if duration < time.Minute {
		return fmt.Sprintf("%.1fs", duration.Seconds())
	} else {
		return fmt.Sprintf("%.1fm", duration.Minutes())
	}
}