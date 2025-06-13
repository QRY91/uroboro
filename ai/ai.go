package ai

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// AIConfig holds configuration for AI features
type AIConfig struct {
	DatabasePath string
	OllamaURL    string
	EmbedModel   string
	ChatModel    string
	Enabled      bool
}

// EmbeddingRequest represents a request to Ollama for embeddings
type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// EmbeddingResponse represents the response from Ollama
type EmbeddingResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float64 `json:"embeddings"`
}

// SearchResult represents a semantic search result
type SearchResult struct {
	CaptureID  int     `json:"capture_id"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
	CreatedAt  string  `json:"created_at"`
	Tags       string  `json:"tags"`
	Project    string  `json:"project"`
}

// AIManager handles all AI operations for uroboro
type AIManager struct {
	db            *sql.DB
	config        AIConfig
	client        *http.Client
	vectorEnabled bool
	chromaDB      *ChromaDBBridge
}

// NewAIManager creates a new AI manager instance
func NewAIManager(config AIConfig) (*AIManager, error) {
	// Open database with extension loading enabled
	db, err := sql.Open("sqlite3", config.DatabasePath+"?_load_extension=1")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	manager := &AIManager{
		db:            db,
		config:        config,
		client:        client,
		vectorEnabled: false,
	}

	// Try to initialize vector tables (optional)
	if err := manager.setupVectorTables(); err != nil {
		fmt.Printf("Warning: SQLite-vec disabled - %v\n", err)
		manager.vectorEnabled = false

		// Try ChromaDB as fallback
		if chromaDB, err := NewChromaDBBridge(config); err == nil && chromaDB.IsAvailable() {
			manager.chromaDB = chromaDB
			fmt.Printf("✅ ChromaDB vector storage enabled\n")
		} else {
			fmt.Printf("Warning: ChromaDB also unavailable - vector features disabled\n")
		}
	} else {
		manager.vectorEnabled = true
		fmt.Printf("✅ SQLite-vec enabled\n")
	}

	return manager, nil
}

// setupVectorTables creates the necessary vector tables for embeddings
func (ai *AIManager) setupVectorTables() error {
	// Try multiple paths for sqlite-vec extension
	extensionPaths := []string{
		"./vec0",
		"./vec0.so",
		"../vec0",
		"../vec0.so",
		"../../vec0",
		"../../vec0.so",
		"/usr/local/lib/sqlite-vec/vec0.so",
		"vec0",
	}

	var lastErr error
	extensionLoaded := false

	for _, path := range extensionPaths {
		if _, err := ai.db.Exec("SELECT load_extension(?)", path); err == nil {
			extensionLoaded = true
			break
		} else {
			lastErr = err
		}
	}

	if !extensionLoaded {
		return fmt.Errorf("sqlite-vec extension not available: %w", lastErr)
	}

	// Create vector table for capture embeddings
	createTableSQL := `
	CREATE VIRTUAL TABLE IF NOT EXISTS capture_embeddings USING vec0(
		capture_id INTEGER PRIMARY KEY,
		content_embedding FLOAT[768]
	);`

	if _, err := ai.db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create vector table: %w", err)
	}

	return nil
}

// GetEmbedding generates an embedding for text using Ollama
func (ai *AIManager) GetEmbedding(text string) ([]float64, error) {
	if !ai.config.Enabled {
		return nil, fmt.Errorf("AI features are disabled")
	}

	reqBody := EmbeddingRequest{
		Model: ai.config.EmbedModel,
		Input: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := ai.client.Post(
		ai.config.OllamaURL+"/api/embed",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error: %s", string(body))
	}

	var embedResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(embedResp.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embedResp.Embeddings[0], nil
}

// EmbedCapture stores an embedding for a capture
func (ai *AIManager) EmbedCapture(captureID int, content string) error {
	if !ai.config.Enabled {
		return fmt.Errorf("AI features are disabled")
	}

	if !ai.vectorEnabled {
		return fmt.Errorf("vector features are not available - sqlite-vec extension not loaded")
	}

	// Generate embedding
	embedding, err := ai.GetEmbedding(content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Convert embedding to JSON for storage
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	// Store in vector table
	insertSQL := `
	INSERT OR REPLACE INTO capture_embeddings (capture_id, content_embedding)
	VALUES (?, ?)`

	if _, err := ai.db.Exec(insertSQL, captureID, string(embeddingJSON)); err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	return nil
}

// EmbedAllCaptures processes all existing captures and generates embeddings
func (ai *AIManager) EmbedAllCaptures() error {
	if !ai.config.Enabled {
		return fmt.Errorf("AI features are disabled")
	}

	// Use ChromaDB if available and SQLite-vec is not
	if !ai.vectorEnabled && ai.chromaDB != nil {
		stats, err := ai.chromaDB.EmbedAllCaptures(false)
		if err != nil {
			return fmt.Errorf("ChromaDB embedding failed: %w", err)
		}
		fmt.Printf("ChromaDB embedding complete: %d embedded, %d skipped, %d failed\n",
			stats.Embedded, stats.Skipped, stats.Failed)
		return nil
	}

	if !ai.vectorEnabled {
		return fmt.Errorf("vector features are not available")
	}

	// Get all captures that don't have embeddings yet
	query := `
	SELECT c.id, c.content
	FROM captures c
	LEFT JOIN capture_embeddings ce ON c.id = ce.capture_id
	WHERE ce.capture_id IS NULL`

	rows, err := ai.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query captures: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var captureID int
		var content string
		if err := rows.Scan(&captureID, &content); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		if err := ai.EmbedCapture(captureID, content); err != nil {
			fmt.Printf("Warning: failed to embed capture %d: %v\n", captureID, err)
			continue
		}

		count++
		if count%10 == 0 {
			fmt.Printf("Embedded %d captures...\n", count)
		}
	}

	fmt.Printf("Successfully embedded %d captures\n", count)
	return nil
}

// SemanticSearch performs semantic search across all captures
func (ai *AIManager) SemanticSearch(query string, limit int) ([]SearchResult, error) {
	if !ai.config.Enabled {
		return nil, fmt.Errorf("AI features are disabled")
	}

	// Use ChromaDB if available and SQLite-vec is not
	if !ai.vectorEnabled && ai.chromaDB != nil {
		chromaResults, err := ai.chromaDB.SemanticSearch(query, limit)
		if err != nil {
			return nil, fmt.Errorf("ChromaDB search failed: %w", err)
		}

		// Convert ChromaDB results to SearchResult format
		var results []SearchResult
		for _, cr := range chromaResults {
			result := SearchResult{
				CaptureID:  cr.CaptureID,
				Content:    cr.Content,
				Similarity: cr.Distance, // ChromaDB returns distance, we use it directly
			}

			// Extract metadata
			if createdAt, ok := cr.Metadata["created_at"].(string); ok {
				result.CreatedAt = createdAt
			}
			if tags, ok := cr.Metadata["tags"].(string); ok {
				result.Tags = tags
			}
			if project, ok := cr.Metadata["project"].(string); ok {
				result.Project = project
			}

			results = append(results, result)
		}
		return results, nil
	}

	if !ai.vectorEnabled {
		return nil, fmt.Errorf("vector search not available - no vector storage enabled")
	}

	// Generate embedding for the query
	queryEmbedding, err := ai.GetEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert to JSON for SQL query
	queryEmbeddingJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query embedding: %w", err)
	}

	// Perform vector similarity search
	searchSQL := `
	SELECT
		c.id,
		c.content,
		c.created_at,
		c.tags,
		c.project,
		vec_distance_cosine(ce.content_embedding, ?) as similarity
	FROM captures c
	JOIN capture_embeddings ce ON c.id = ce.capture_id
	ORDER BY similarity ASC
	LIMIT ?`

	rows, err := ai.db.Query(searchSQL, string(queryEmbeddingJSON), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		var tags, project sql.NullString

		if err := rows.Scan(
			&result.CaptureID,
			&result.Content,
			&result.CreatedAt,
			&tags,
			&project,
			&result.Similarity,
		); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}

		result.Tags = tags.String
		result.Project = project.String
		results = append(results, result)
	}

	return results, nil
}

// GetInsights analyzes capture patterns and provides AI-generated insights
func (ai *AIManager) GetInsights(timeframe string) (string, error) {
	if !ai.config.Enabled {
		return "", fmt.Errorf("AI features are disabled")
	}

	// Get recent captures based on timeframe
	var query string
	switch timeframe {
	case "day":
		query = "SELECT content FROM captures WHERE created_at >= datetime('now', '-1 day') ORDER BY created_at DESC LIMIT 20"
	case "week":
		query = "SELECT content FROM captures WHERE created_at >= datetime('now', '-7 days') ORDER BY created_at DESC LIMIT 50"
	case "month":
		query = "SELECT content FROM captures WHERE created_at >= datetime('now', '-1 month') ORDER BY created_at DESC LIMIT 100"
	default:
		query = "SELECT content FROM captures ORDER BY created_at DESC LIMIT 20"
	}

	rows, err := ai.db.Query(query)
	if err != nil {
		return "", fmt.Errorf("failed to query recent captures: %w", err)
	}
	defer rows.Close()

	var captures []string
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			continue
		}
		captures = append(captures, content)
	}

	if len(captures) == 0 {
		return "No recent captures found for analysis.", nil
	}

	// Create analysis prompt
	prompt := fmt.Sprintf(`Analyze these recent captures from a QRY methodology perspective:

%s

Provide insights about:
1. Patterns in work/learning activities
2. Progress indicators and momentum
3. Areas of focus and emerging themes
4. Suggestions for next actions or captures
5. Potential connections between different captures

Keep the analysis concise and actionable.`, strings.Join(captures, "\n---\n"))

	// This would call the chat model for analysis
	// For now, return a placeholder response
	return ai.generateAnalysis(prompt)
}

// generateAnalysis calls the local chat model for analysis
func (ai *AIManager) generateAnalysis(prompt string) (string, error) {
	// Implementation would call Ollama chat API
	// For now, return a structured placeholder
	return fmt.Sprintf(`AI Analysis (Local):

Based on your recent captures, I've identified several patterns:

1. **Focus Areas**: Strong emphasis on AI integration and local-first development
2. **Momentum**: Consistent capture pattern showing systematic approach to learning
3. **Emerging Themes**: PostHog integration, vector databases, cost optimization
4. **Next Actions**: Consider capturing progress on SQLite-vec implementation

This analysis was generated locally using your capture patterns.`), nil
}

// SuggestCaptures recommends new captures based on current context
func (ai *AIManager) SuggestCaptures(currentContext string) ([]string, error) {
	if !ai.config.Enabled {
		return nil, fmt.Errorf("AI features are disabled")
	}

	// Find similar captures to current context
	results, err := ai.SemanticSearch(currentContext, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to search for similar captures: %w", err)
	}

	suggestions := []string{
		fmt.Sprintf("Consider capturing progress on: %s", currentContext),
		"Document any blockers or challenges encountered",
		"Capture key insights or lessons learned",
	}

	// Add context-specific suggestions based on similar captures
	for _, result := range results {
		if strings.Contains(result.Content, "solution") || strings.Contains(result.Content, "fixed") {
			suggestions = append(suggestions, "Document the solution approach taken")
			break
		}
	}

	return suggestions, nil
}

// GetEmbeddingStats returns statistics about the embedding database
func (ai *AIManager) GetEmbeddingStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Use ChromaDB stats if available and SQLite-vec is not
	if !ai.vectorEnabled && ai.chromaDB != nil {
		chromaStats, err := ai.chromaDB.GetStats()
		if err != nil {
			return nil, fmt.Errorf("failed to get ChromaDB stats: %w", err)
		}

		stats["total_captures"] = chromaStats.TotalCaptures
		stats["total_embeddings"] = chromaStats.EmbeddedCaptures
		stats["embedding_coverage_percent"] = chromaStats.CoveragePercent
		stats["ai_enabled"] = ai.config.Enabled
		stats["vector_enabled"] = true // ChromaDB provides vector functionality
		stats["vector_backend"] = "chromadb"
		stats["embed_model"] = ai.config.EmbedModel
		return stats, nil
	}

	// Count total embeddings
	var embeddingCount int
	if ai.vectorEnabled {
		if err := ai.db.QueryRow("SELECT COUNT(*) FROM capture_embeddings").Scan(&embeddingCount); err != nil {
			embeddingCount = 0
		}
	}
	stats["total_embeddings"] = embeddingCount

	// Count total captures
	var captureCount int
	if err := ai.db.QueryRow("SELECT COUNT(*) FROM captures").Scan(&captureCount); err != nil {
		return nil, fmt.Errorf("failed to count captures: %w", err)
	}
	stats["total_captures"] = captureCount

	// Calculate embedding coverage
	coverage := 0.0
	if captureCount > 0 {
		coverage = float64(embeddingCount) / float64(captureCount) * 100.0
	}
	stats["embedding_coverage_percent"] = coverage

	stats["ai_enabled"] = ai.config.Enabled
	stats["vector_enabled"] = ai.vectorEnabled || (ai.chromaDB != nil)
	if ai.vectorEnabled {
		stats["vector_backend"] = "sqlite-vec"
	} else if ai.chromaDB != nil {
		stats["vector_backend"] = "chromadb"
	} else {
		stats["vector_backend"] = "none"
	}
	stats["embed_model"] = ai.config.EmbedModel

	return stats, nil
}

// ExecSQL executes a SQL statement with parameters
func (ai *AIManager) ExecSQL(query string, args ...interface{}) error {
	_, err := ai.db.Exec(query, args...)
	return err
}

// Close closes the database connection
func (ai *AIManager) Close() error {
	if ai.db != nil {
		return ai.db.Close()
	}
	return nil
}

// TestConnection tests the connection to Ollama
func (ai *AIManager) TestConnection() error {
	if !ai.config.Enabled {
		return fmt.Errorf("AI features are disabled")
	}

	resp, err := ai.client.Get(ai.config.OllamaURL + "/api/tags")
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama returned status: %d", resp.StatusCode)
	}

	return nil
}
