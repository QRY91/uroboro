package ai

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// ChromaDBBridge provides integration with ChromaDB through Python
type ChromaDBBridge struct {
	pythonPath    string
	scriptPath    string
	uroboroDBPath string
	chromaDBPath  string
	ollamaURL     string
	embedModel    string
	enabled       bool
}

// ChromaDBSearchResult represents a search result from ChromaDB
type ChromaDBSearchResult struct {
	CaptureID  int                    `json:"capture_id"`
	Content    string                 `json:"content"`
	Similarity float64                `json:"similarity"`
	Distance   float64                `json:"distance"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ChromaDBStats represents statistics from ChromaDB
type ChromaDBStats struct {
	TotalCaptures    int     `json:"total_captures"`
	EmbeddedCaptures int     `json:"embedded_captures"`
	CoveragePercent  float64 `json:"coverage_percent"`
	ChromaDBPath     string  `json:"chromadb_path"`
	UroboroDBPath    string  `json:"uroboro_db_path"`
	EmbedModel       string  `json:"embed_model"`
	CollectionName   string  `json:"collection_name"`
}

// ChromaDBConnectionTest represents connection test results
type ChromaDBConnectionTest struct {
	ChromaDB  bool `json:"chromadb"`
	Ollama    bool `json:"ollama"`
	UroboroDB bool `json:"uroboro_db"`
}

// ChromaDBEmbedStats represents embedding operation statistics
type ChromaDBEmbedStats struct {
	Total    int `json:"total"`
	Embedded int `json:"embedded"`
	Skipped  int `json:"skipped"`
	Failed   int `json:"failed"`
}

// NewChromaDBBridge creates a new ChromaDB bridge
func NewChromaDBBridge(config AIConfig) (*ChromaDBBridge, error) {
	// Find Python executable
	pythonPath, err := findPythonExecutable()
	if err != nil {
		return nil, fmt.Errorf("Python not found: %w", err)
	}

	// Determine script path (relative to current location)
	scriptPath, err := filepath.Abs("./ai/chromadb_integration.py")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve script path: %w", err)
	}

	bridge := &ChromaDBBridge{
		pythonPath:    pythonPath,
		scriptPath:    scriptPath,
		uroboroDBPath: config.DatabasePath,
		ollamaURL:     config.OllamaURL,
		embedModel:    config.EmbedModel,
		enabled:       config.Enabled,
	}

	return bridge, nil
}

// runPythonCommand executes a ChromaDB integration command
func (c *ChromaDBBridge) runPythonCommand(args []string) ([]byte, error) {
	if !c.enabled {
		return nil, fmt.Errorf("ChromaDB bridge is disabled")
	}

	// Prepare command
	cmdArgs := append([]string{c.scriptPath}, args...)
	cmd := exec.Command(c.pythonPath, cmdArgs...)

	// Set environment variables if needed
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("UROBORO_DB_PATH=%s", c.uroboroDBPath),
		fmt.Sprintf("OLLAMA_URL=%s", c.ollamaURL),
		fmt.Sprintf("EMBED_MODEL=%s", c.embedModel),
	)

	// Execute command
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("command failed: %w\nStderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// EmbedAllCaptures embeds all captures using ChromaDB
func (c *ChromaDBBridge) EmbedAllCaptures(forceReembed bool) (*ChromaDBEmbedStats, error) {
	args := []string{"embed"}
	if forceReembed {
		args = append(args, "--force")
	}

	output, err := c.runPythonCommand(args)
	if err != nil {
		return nil, fmt.Errorf("embed command failed: %w", err)
	}

	// Parse output to extract statistics
	// The Python script outputs structured text, we need to parse it
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	stats := &ChromaDBEmbedStats{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Total:") {
			if val, err := extractNumber(line); err == nil {
				stats.Total = val
			}
		} else if strings.Contains(line, "Embedded:") {
			if val, err := extractNumber(line); err == nil {
				stats.Embedded = val
			}
		} else if strings.Contains(line, "Skipped:") {
			if val, err := extractNumber(line); err == nil {
				stats.Skipped = val
			}
		} else if strings.Contains(line, "Failed:") {
			if val, err := extractNumber(line); err == nil {
				stats.Failed = val
			}
		}
	}

	return stats, nil
}

// SemanticSearch performs semantic search using ChromaDB
func (c *ChromaDBBridge) SemanticSearch(query string, limit int) ([]ChromaDBSearchResult, error) {
	args := []string{"search", query, "--limit", strconv.Itoa(limit)}

	output, err := c.runPythonCommand(args)
	if err != nil {
		return nil, fmt.Errorf("search command failed: %w", err)
	}

	// Parse the structured output from the Python script
	// For now, we'll parse the text output; ideally we'd use JSON
	outputStr := string(output)
	results := parseSearchResults(outputStr)

	return results, nil
}

// GetStats retrieves ChromaDB statistics
func (c *ChromaDBBridge) GetStats() (*ChromaDBStats, error) {
	args := []string{"stats"}

	output, err := c.runPythonCommand(args)
	if err != nil {
		return nil, fmt.Errorf("stats command failed: %w", err)
	}

	// Parse the stats output
	outputStr := string(output)
	stats := parseStatsOutput(outputStr)

	return stats, nil
}

// TestConnection tests ChromaDB and related connections
func (c *ChromaDBBridge) TestConnection() (*ChromaDBConnectionTest, error) {
	args := []string{"test"}

	output, err := c.runPythonCommand(args)
	if err != nil {
		return nil, fmt.Errorf("test command failed: %w", err)
	}

	// Parse the test output
	outputStr := string(output)
	result := parseTestOutput(outputStr)

	return result, nil
}

// ResetCollection resets the ChromaDB collection
func (c *ChromaDBBridge) ResetCollection() error {
	args := []string{"reset"}

	_, err := c.runPythonCommand(args)
	if err != nil {
		return fmt.Errorf("reset command failed: %w", err)
	}

	return nil
}

// Helper functions for parsing Python script output

func extractNumber(line string) (int, error) {
	// Extract number from a line like "Total: 42"
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid format")
	}

	numStr := strings.TrimSpace(parts[1])
	return strconv.Atoi(numStr)
}

func parseSearchResults(output string) []ChromaDBSearchResult {
	var results []ChromaDBSearchResult
	lines := strings.Split(output, "\n")

	var currentResult *ChromaDBSearchResult
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "🎯 Result") {
			// Start of a new result
			if currentResult != nil {
				results = append(results, *currentResult)
			}
			currentResult = &ChromaDBSearchResult{
				Metadata: make(map[string]interface{}),
			}

			// Extract similarity from the line
			if idx := strings.Index(line, "("); idx != -1 {
				if endIdx := strings.Index(line[idx:], "%"); endIdx != -1 {
					simStr := line[idx+1 : idx+endIdx]
					if sim, err := strconv.ParseFloat(simStr, 64); err == nil {
						currentResult.Similarity = sim / 100.0
					}
				}
			}
		} else if currentResult != nil {
			if strings.HasPrefix(line, "ID:") {
				if id, err := extractNumber(line); err == nil {
					currentResult.CaptureID = id
				}
			} else if strings.HasPrefix(line, "Created:") {
				currentResult.Metadata["created_at"] = strings.TrimSpace(strings.TrimPrefix(line, "Created:"))
			} else if strings.HasPrefix(line, "Project:") {
				currentResult.Metadata["project"] = strings.TrimSpace(strings.TrimPrefix(line, "Project:"))
			} else if strings.HasPrefix(line, "Tags:") {
				currentResult.Metadata["tags"] = strings.TrimSpace(strings.TrimPrefix(line, "Tags:"))
			} else if line != "" && !strings.HasPrefix(line, "🎯") && !strings.HasPrefix(line, "🔍") {
				// This might be content
				if currentResult.Content == "" {
					currentResult.Content = line
				}
			}
		}
	}

	// Don't forget the last result
	if currentResult != nil {
		results = append(results, *currentResult)
	}

	return results
}

func parseStatsOutput(output string) *ChromaDBStats {
	stats := &ChromaDBStats{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "Total Captures:") {
			if val, err := extractNumber(line); err == nil {
				stats.TotalCaptures = val
			}
		} else if strings.Contains(line, "Embedded:") {
			if val, err := extractNumber(line); err == nil {
				stats.EmbeddedCaptures = val
			}
		} else if strings.Contains(line, "Coverage:") {
			// Extract percentage from "Coverage: 95.5%"
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				percentStr := strings.TrimSpace(strings.TrimSuffix(parts[1], "%"))
				if percent, err := strconv.ParseFloat(percentStr, 64); err == nil {
					stats.CoveragePercent = percent
				}
			}
		} else if strings.Contains(line, "Model:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				stats.EmbedModel = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "ChromaDB Path:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				stats.ChromaDBPath = strings.TrimSpace(parts[1])
			}
		}
	}

	return stats
}

func parseTestOutput(output string) *ChromaDBConnectionTest {
	result := &ChromaDBConnectionTest{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "ChromaDB:") {
			result.ChromaDB = strings.Contains(line, "✅")
		} else if strings.Contains(line, "Ollama:") {
			result.Ollama = strings.Contains(line, "✅")
		} else if strings.Contains(line, "Uroboro DB:") {
			result.UroboroDB = strings.Contains(line, "✅")
		}
	}

	return result
}

// IsAvailable checks if ChromaDB integration is available
func (c *ChromaDBBridge) IsAvailable() bool {
	if !c.enabled {
		return false
	}

	// Check if Python script exists
	if _, err := exec.LookPath(c.pythonPath); err != nil {
		return false
	}

	// Try a simple test command
	_, err := c.runPythonCommand([]string{"--help"})
	return err == nil
}

// findPythonExecutable attempts to find a Python executable
func findPythonExecutable() (string, error) {
	// List of possible Python executables to try
	candidates := []string{"python3", "python", "python3.9", "python3.8", "python3.10", "python3.11"}

	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no Python executable found in PATH")
}
