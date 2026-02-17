package distill

import "time"

// GitExtract is one JSONL record from the git extractor.
type GitExtract struct {
	Source     string    `json:"source"`
	Repo       string    `json:"repo"`
	Hash       string    `json:"hash"`
	ParentHash string    `json:"parent_hash"`
	Message    string    `json:"message"`
	Date       time.Time `json:"date"`
	Files      []string  `json:"files"`
	Language   string    `json:"language"`
	Diff       string    `json:"diff"`
	DiffStats  DiffStats `json:"diff_stats"`
}

type DiffStats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Files     int `json:"files"`
}

// UroExtract is one JSONL record from the uroboro extractor.
type UroExtract struct {
	Source            string    `json:"source"`
	Type              string    `json:"type"`
	Content           string    `json:"content"`
	Project           string    `json:"project"`
	Tags              []string  `json:"tags"`
	Timestamp         time.Time `json:"timestamp"`
	CorrelatedGitHash string    `json:"correlated_git_hash,omitempty"`
}

// LanguageFromExt maps file extensions to language names.
var LanguageFromExt = map[string]string{
	".go":   "go",
	".py":   "python",
	".ts":   "typescript",
	".tsx":  "typescript",
	".js":   "javascript",
	".jsx":  "javascript",
	".rs":   "rust",
	".rb":   "ruby",
	".java": "java",
	".c":    "c",
	".cpp":  "cpp",
	".h":    "c",
	".cs":   "csharp",
	".sh":   "shell",
	".sql":  "sql",
	".html": "html",
	".css":  "css",
	".yaml": "yaml",
	".yml":  "yaml",
	".json": "json",
	".toml": "toml",
}
