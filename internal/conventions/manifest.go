package conventions

import (
	"encoding/json"
	"os"
)

type Manifest struct {
	Date            string         `json:"date"`
	Repos           []string       `json:"repos"`
	GitCount        int            `json:"git_count"`
	UroCount        int            `json:"uro_count"`
	CorrelatedCount int            `json:"correlated_count"`
	TotalRecords    int            `json:"total_records"`
	Days            int            `json:"days,omitempty"`
	Since           string         `json:"since,omitempty"`
	OutputFile      string         `json:"output"`
	AuditDir        string         `json:"audit_dir,omitempty"`
	ConventionsDir  string         `json:"conventions_dir"`
	LanguageCounts  map[string]int `json:"language_counts"`
	RepoCounts      map[string]int `json:"repo_counts"`
}

func WriteManifest(path string, m Manifest) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}
