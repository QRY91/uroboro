package journey

import "time"

type TimelineEvent struct {
	Timestamp  time.Time `json:"timestamp"`
	Content    string    `json:"content"`
	Project    string    `json:"project"`
	Branch     string    `json:"branch,omitempty"`
	Tags       []string  `json:"tags"`
	EventType  string    `json:"eventType"`
	Importance int       `json:"importance"`
	GitHash    string    `json:"gitHash,omitempty"`
}

type JourneyData struct {
	Events     []TimelineEvent  `json:"events"`
	Timeline   Timeline         `json:"timeline"`
	Projects   []ProjectSummary `json:"projects"`
	Stats      JourneyStats     `json:"stats"`
	Milestones []TimelineEvent  `json:"milestones"`
}

type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type Timeline struct {
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	TotalDuration int64     `json:"totalDuration"`
}

type ProjectSummary struct {
	Name       string    `json:"name"`
	EventCount int       `json:"eventCount"`
	Color      string    `json:"color"`
	StartDate  time.Time `json:"startDate"`
	LastActive time.Time `json:"lastActive"`
}

type JourneyStats struct {
	TotalEvents       int     `json:"totalEvents"`
	ProjectCount      int     `json:"projectCount"`
	MilestoneCount    int     `json:"milestoneCount"`
	ProductivityScore float64 `json:"productivityScore"`
	LearningMoments   int     `json:"learningMoments"`
}

type Options struct {
	Days      int
	DateRange *DateRange
	Projects  []string
	Port      int
}

const (
	EventTypeCapture   = "capture"
	EventTypeCommit    = "commit"
	EventTypeMilestone = "milestone"
	EventTypeLearning  = "learning"
	EventTypeDecision  = "decision"
	EventTypeBugfix    = "bugfix"
	EventTypeFeature   = "feature"
	EventTypeBlocker   = "blocker"
	EventTypeQuestion  = "question"
)

const (
	ImportanceLow      = 1
	ImportanceMedium   = 2
	ImportanceHigh     = 3
	ImportanceCritical = 4
)
