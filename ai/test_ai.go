package ai

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestAIIntegration runs a comprehensive test of AI features
func TestAIIntegration(t *testing.T) {
	// Create temporary database for testing
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_uroboro.sqlite")

	// Initialize test database with sample captures
	if err := setupTestDatabase(dbPath); err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Create AI config for testing
	config := AIConfig{
		DatabasePath: dbPath,
		OllamaURL:    "http://localhost:11434",
		EmbedModel:   "nomic-embed-text",
		ChatModel:    "llama3.2:1b",
		Enabled:      true,
	}

	// Test AI manager initialization
	t.Run("AIManagerInit", func(t *testing.T) {
		ai, err := NewAIManager(config)
		if err != nil {
			t.Skipf("Skipping AI tests - failed to initialize: %v", err)
		}
		defer ai.Close()

		// Test connection
		if err := ai.TestConnection(); err != nil {
			t.Skipf("Skipping AI tests - Ollama not available: %v", err)
		}
	})

	// Test embedding generation
	t.Run("EmbeddingGeneration", func(t *testing.T) {
		ai, err := NewAIManager(config)
		if err != nil {
			t.Skip("AI manager not available")
		}
		defer ai.Close()

		embedding, err := ai.GetEmbedding("test capture content")
		if err != nil {
			t.Fatalf("Failed to generate embedding: %v", err)
		}

		if len(embedding) != 768 {
			t.Errorf("Expected embedding dimension 768, got %d", len(embedding))
		}
	})

	// Test capture embedding
	t.Run("CaptureEmbedding", func(t *testing.T) {
		ai, err := NewAIManager(config)
		if err != nil {
			t.Skip("AI manager not available")
		}
		defer ai.Close()

		err = ai.EmbedCapture(1, "PostHog integration testing with local AI")
		if err != nil {
			t.Fatalf("Failed to embed capture: %v", err)
		}
	})

	// Test semantic search
	t.Run("SemanticSearch", func(t *testing.T) {
		ai, err := NewAIManager(config)
		if err != nil {
			t.Skip("AI manager not available")
		}
		defer ai.Close()

		// First embed some test captures
		testCases := []struct {
			id      int
			content string
		}{
			{1, "PostHog integration with local AI and vector search"},
			{2, "ESP32 development for hardware projects"},
			{3, "SQLite vector database implementation"},
		}

		for _, tc := range testCases {
			if err := ai.EmbedCapture(tc.id, tc.content); err != nil {
				t.Fatalf("Failed to embed test capture %d: %v", tc.id, err)
			}
		}

		// Test search
		results, err := ai.SemanticSearch("PostHog AI integration", 5)
		if err != nil {
			t.Fatalf("Semantic search failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected search results, got none")
		}

		// Verify result structure
		if len(results) > 0 {
			result := results[0]
			if result.CaptureID == 0 {
				t.Error("Expected non-zero capture ID")
			}
			if result.Content == "" {
				t.Error("Expected non-empty content")
			}
		}
	})

	// Test insights generation
	t.Run("InsightsGeneration", func(t *testing.T) {
		ai, err := NewAIManager(config)
		if err != nil {
			t.Skip("AI manager not available")
		}
		defer ai.Close()

		insights, err := ai.GetInsights("day")
		if err != nil {
			t.Fatalf("Failed to generate insights: %v", err)
		}

		if insights == "" {
			t.Error("Expected non-empty insights")
		}
	})

	// Test suggestions
	t.Run("CapturesSuggestions", func(t *testing.T) {
		ai, err := NewAIManager(config)
		if err != nil {
			t.Skip("AI manager not available")
		}
		defer ai.Close()

		suggestions, err := ai.SuggestCaptures("working on PostHog integration")
		if err != nil {
			t.Fatalf("Failed to get suggestions: %v", err)
		}

		if len(suggestions) == 0 {
			t.Error("Expected at least one suggestion")
		}
	})

	// Test statistics
	t.Run("AIStats", func(t *testing.T) {
		ai, err := NewAIManager(config)
		if err != nil {
			t.Skip("AI manager not available")
		}
		defer ai.Close()

		stats, err := ai.GetEmbeddingStats()
		if err != nil {
			t.Fatalf("Failed to get stats: %v", err)
		}

		// Verify expected keys exist
		expectedKeys := []string{"total_embeddings", "total_captures", "embedding_coverage_percent", "ai_enabled", "embed_model"}
		for _, key := range expectedKeys {
			if _, exists := stats[key]; !exists {
				t.Errorf("Expected stats key %s not found", key)
			}
		}
	})
}

// setupTestDatabase creates a test database with sample captures
func setupTestDatabase(dbPath string) error {
	// Create database file
	file, err := os.Create(dbPath)
	if err != nil {
		return err
	}
	file.Close()

	// Initialize with basic schema (simplified version of uroboro schema)
	config := AIConfig{
		DatabasePath: dbPath,
		OllamaURL:    "http://localhost:11434",
		EmbedModel:   "nomic-embed-text",
		ChatModel:    "llama3.2:1b",
		Enabled:      false, // Don't try to load extension yet
	}

	ai, err := NewAIManager(config)
	if err != nil {
		return err
	}
	defer ai.Close()

	// Create captures table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS captures (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		tags TEXT,
		project TEXT
	);`

	if _, err := ai.db.Exec(createTableSQL); err != nil {
		return err
	}

	// Insert test data
	testCaptures := []struct {
		content string
		tags    string
		project string
	}{
		{"PostHog integration analysis complete - excellent architecture insights", "AI,posthog,analysis", "posthog-integration"},
		{"SQLite-vec extension working perfectly for local vector search", "AI,sqlite,vector", "uroboro"},
		{"ESP32 development continues with DeskHog hardware prototyping", "hardware,esp32", "hardware-dev"},
		{"Local-first AI approach saving significant costs vs cloud APIs", "AI,cost,local-first", "uroboro"},
		{"QRY methodology applied to AI development workflow", "methodology,AI,qry", "qry-methodology"},
	}

	for _, capture := range testCaptures {
		insertSQL := `INSERT INTO captures (content, tags, project) VALUES (?, ?, ?)`
		if _, err := ai.db.Exec(insertSQL, capture.content, capture.tags, capture.project); err != nil {
			return err
		}
	}

	return nil
}

// BenchmarkEmbeddingGeneration benchmarks embedding generation performance
func BenchmarkEmbeddingGeneration(b *testing.B) {
	config := AIConfig{
		DatabasePath: ":memory:",
		OllamaURL:    "http://localhost:11434",
		EmbedModel:   "nomic-embed-text",
		ChatModel:    "llama3.2:1b",
		Enabled:      true,
	}

	ai, err := NewAIManager(config)
	if err != nil {
		b.Skipf("AI manager not available: %v", err)
	}
	defer ai.Close()

	testText := "This is a test capture for benchmarking embedding generation performance"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ai.GetEmbedding(testText)
		if err != nil {
			b.Fatalf("Embedding generation failed: %v", err)
		}
	}
}

// TestCLICommands tests the CLI interface
func TestCLICommands(t *testing.T) {
	// Test help command
	t.Run("HelpCommand", func(t *testing.T) {
		err := RunAICommand([]string{"help"})
		if err != nil {
			t.Errorf("Help command failed: %v", err)
		}
	})

	// Test with no arguments
	t.Run("NoArgs", func(t *testing.T) {
		err := RunAICommand([]string{})
		if err != nil {
			t.Errorf("No args command failed: %v", err)
		}
	})

	// Test unknown command
	t.Run("UnknownCommand", func(t *testing.T) {
		err := RunAICommand([]string{"unknown"})
		if err != nil {
			t.Errorf("Unknown command handling failed: %v", err)
		}
	})
}

// TestUtilityFunctions tests utility functions
func TestUtilityFunctions(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultCLIConfig()
		if config.EmbedModel == "" {
			t.Error("Expected non-empty embed model")
		}
		if config.OllamaURL == "" {
			t.Error("Expected non-empty Ollama URL")
		}
	})

	t.Run("ParseLimit", func(t *testing.T) {
		args := []string{"search", "test", "--limit", "20"}
		limit := ParseLimit(args, 10)
		if limit != 20 {
			t.Errorf("Expected limit 20, got %d", limit)
		}

		// Test default
		args = []string{"search", "test"}
		limit = ParseLimit(args, 10)
		if limit != 10 {
			t.Errorf("Expected default limit 10, got %d", limit)
		}
	})

	t.Run("FormatResults", func(t *testing.T) {
		results := []SearchResult{
			{
				CaptureID:  1,
				Content:    "test content",
				Similarity: 0.95,
				CreatedAt:  time.Now().Format(time.RFC3339),
				Tags:       "test",
				Project:    "test-project",
			},
		}

		jsonStr, err := FormatSearchResults(results)
		if err != nil {
			t.Fatalf("Failed to format results: %v", err)
		}

		if jsonStr == "" {
			t.Error("Expected non-empty JSON string")
		}
	})
}

// TestErrorHandling tests error conditions
func TestErrorHandling(t *testing.T) {
	t.Run("InvalidDatabase", func(t *testing.T) {
		config := AIConfig{
			DatabasePath: "/invalid/path/database.sqlite",
			OllamaURL:    "http://localhost:11434",
			EmbedModel:   "nomic-embed-text",
			ChatModel:    "llama3.2:1b",
			Enabled:      true,
		}

		_, err := NewAIManager(config)
		if err == nil {
			t.Error("Expected error for invalid database path")
		}
	})

	t.Run("DisabledAI", func(t *testing.T) {
		config := AIConfig{
			DatabasePath: ":memory:",
			OllamaURL:    "http://localhost:11434",
			EmbedModel:   "nomic-embed-text",
			ChatModel:    "llama3.2:1b",
			Enabled:      false,
		}

		ai, err := NewAIManager(config)
		if err != nil {
			t.Skip("Failed to create AI manager")
		}
		defer ai.Close()

		_, err = ai.GetEmbedding("test")
		if err == nil {
			t.Error("Expected error when AI is disabled")
		}
	})
}
