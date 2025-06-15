package journey

import (
	"time"
)

// TimelineEvent represents a single event in the development journey
type TimelineEvent struct {
	Timestamp    time.Time              `json:"timestamp"`
	Content      string                 `json:"content"`
	Project      string                 `json:"project"`
	Tags         []string               `json:"tags"`
	Context      map[string]interface{} `json:"context"`
	EventType    string                 `json:"eventType"`
	Importance   int                    `json:"importance"`
	GitHash      string                 `json:"gitHash,omitempty"`
	FilesChanged []string               `json:"filesChanged,omitempty"`
}

// JourneyData represents the complete journey data for a time period
type JourneyData struct {
	Events     []TimelineEvent  `json:"events"`
	DateRange  DateRange        `json:"dateRange"`
	Projects   []ProjectSummary `json:"projects"`
	Stats      JourneyStats     `json:"stats"`
	Milestones []TimelineEvent  `json:"milestones"`
}

// DateRange represents a time range for journey data
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ProjectSummary provides summary information about a project
type ProjectSummary struct {
	Name       string    `json:"name"`
	EventCount int       `json:"eventCount"`
	Color      string    `json:"color"`
	StartDate  time.Time `json:"startDate"`
	LastActive time.Time `json:"lastActive"`
}

// JourneyStats provides statistical information about the journey
type JourneyStats struct {
	TotalEvents       int     `json:"totalEvents"`
	ProjectCount      int     `json:"projectCount"`
	MilestoneCount    int     `json:"milestoneCount"`
	ProductivityScore float64 `json:"productivityScore"`
	LearningMoments   int     `json:"learningMoments"`
}

// JourneyOptions represents configuration options for journey generation
type JourneyOptions struct {
	Days      int        `json:"days"`
	DateRange *DateRange `json:"dateRange,omitempty"`
	Projects  []string   `json:"projects,omitempty"`
	Export    bool       `json:"export"`
	Live      bool       `json:"live"`
	Port      int        `json:"port"`
	AutoOpen  bool       `json:"autoOpen"`
	Share     bool       `json:"share"`
	Title     string     `json:"title"`
	Theme     string     `json:"theme"`
}

// EventType constants for different types of timeline events
const (
	EventTypeCapture     = "capture"
	EventTypeCommit      = "commit"
	EventTypeMilestone   = "milestone"
	EventTypeLearning    = "learning"
	EventTypeDecision    = "decision"
	EventTypeIntegration = "integration"
	EventTypeBugfix      = "bugfix"
	EventTypeFeature     = "feature"
	EventTypeRefactor    = "refactor"
)

// Importance levels for events
const (
	ImportanceLow      = 1
	ImportanceMedium   = 2
	ImportanceHigh     = 3
	ImportanceCritical = 4
)

// Theme constants for journey visualization
const (
	ThemeDefault = "default"
	ThemeDark    = "dark"
	ThemeLight   = "light"
	ThemeMatrix  = "matrix"
	ThemeNeon    = "neon"
)
