package promptprofile

import (
	"encoding/json"
	"time"
)

// SessionIndex is the top-level sessions-index.json structure.
type SessionIndex struct {
	Version int           `json:"version"`
	Entries []SessionMeta `json:"entries"`
}

// SessionMeta from sessions-index.json.
type SessionMeta struct {
	SessionID   string `json:"sessionId"`
	FullPath    string `json:"fullPath"`
	FirstPrompt string `json:"firstPrompt"`
	Summary     string `json:"summary"`
	MsgCount    int    `json:"messageCount"`
	Created     string `json:"created"`
	Modified    string `json:"modified"`
	GitBranch   string `json:"gitBranch"`
	ProjectPath string `json:"projectPath"`
	IsSidechain bool   `json:"isSidechain"`
}

// rawLine is the minimal structure for fast filtering of JSONL lines.
type rawLine struct {
	Type          string      `json:"type"`
	IsMeta        bool        `json:"isMeta"`
	UserType      string      `json:"userType"`
	Timestamp     time.Time   `json:"timestamp"`
	SessionID     string      `json:"sessionId"`
	GitBranch     string      `json:"gitBranch"`
	ToolUseResult interface{} `json:"toolUseResult"`
	Message       rawMessage  `json:"message"`
}

type rawMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// contentBlock is one element of a content array.
type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// UserPrompt is the extracted, classified user message.
type UserPrompt struct {
	SessionID    string    `json:"session_id"`
	Project      string    `json:"project"`
	Timestamp    time.Time `json:"timestamp"`
	Text         string    `json:"text"`
	WordCount    int       `json:"word_count"`
	LineCount    int       `json:"line_count"`
	Branch       string    `json:"branch,omitempty"`
	IsFirstMsg   bool      `json:"is_first_message"`
	HasFilePath  bool      `json:"has_file_path"`
	HasCodeBlock bool      `json:"has_code_block"`
	IsImperative bool      `json:"is_imperative"`
	IsQuestion   bool      `json:"is_question"`
}

// PromptStats is the aggregate statistics output.
type PromptStats struct {
	TotalSessions        int
	TotalPrompts         int
	TotalWords           int
	AvgWordsPerPrompt    float64
	AvgPromptsPerSession float64
	EarliestDate         time.Time
	LatestDate           time.Time

	// Length distribution
	ShortPrompts  int // < 20 words
	MediumPrompts int // 20-100 words
	LongPrompts   int // 100-500 words
	VeryLong      int // 500+ words

	// Style percentages
	ImperativeCount int
	QuestionCount   int
	FileRefCount    int
	CodeBlockCount  int
	ContextOpeners  int // first message in session, >100 words

	// Project distribution
	ProjectCounts map[string]int

	// Temporal patterns
	HourCounts    map[int]int
	WeekdayCounts map[string]int
}

// ProjectInfo describes a discovered Claude Code project directory.
type ProjectInfo struct {
	Slug        string // directory name, e.g. "-home-qry-projects-uroboro"
	ProjectName string // e.g. "uroboro"
	DirPath     string // full path to the project dir
	Sessions    []SessionMeta
}
