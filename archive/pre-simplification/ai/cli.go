package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CLIConfig holds configuration for AI CLI commands
type CLIConfig struct {
	DatabasePath string
	OllamaURL    string
	EmbedModel   string
	ChatModel    string
	Enabled      bool
}

// ToAIConfig converts CLIConfig to AIConfig
func (c CLIConfig) ToAIConfig() AIConfig {
	return AIConfig{
		DatabasePath: c.DatabasePath,
		OllamaURL:    c.OllamaURL,
		EmbedModel:   c.EmbedModel,
		ChatModel:    c.ChatModel,
		Enabled:      c.Enabled,
	}
}

// DefaultCLIConfig returns default configuration for AI features
func DefaultCLIConfig() CLIConfig {
	return CLIConfig{
		DatabasePath: os.ExpandEnv("$HOME/.local/share/uroboro/uroboro.sqlite"),
		OllamaURL:    "http://localhost:11434",
		EmbedModel:   "nomic-embed-text",
		ChatModel:    "llama3.2:1b",
		Enabled:      true,
	}
}

// RunAICommand handles AI-related CLI commands
func RunAICommand(args []string) error {
	if len(args) == 0 {
		return showAIHelp()
	}

	config := DefaultCLIConfig()

	switch args[0] {
	case "setup":
		return setupAI(config)
	case "embed":
		return embedCaptures(config, args[1:])
	case "search":
		return semanticSearch(config, args[1:])
	case "insights":
		return getInsights(config, args[1:])
	case "suggest":
		return getSuggestions(config, args[1:])
	case "stats":
		return showAIStats(config)
	case "test":
		return testAIConnection(config)
	case "help":
		return showAIHelp()
	default:
		fmt.Printf("Unknown AI command: %s\n", args[0])
		return showAIHelp()
	}
}

// showAIHelp displays help for AI commands
func showAIHelp() error {
	fmt.Println(`🤖 uroboro AI Commands

SETUP:
  uroboro ai setup              Initialize AI features and vector database
  uroboro ai test               Test connection to Ollama

EMBEDDING:
  uroboro ai embed              Embed all captures for semantic search
  uroboro ai embed --new        Embed only new captures (not yet embedded)

SEARCH & ANALYSIS:
  uroboro ai search <query>     Semantic search across all captures
  uroboro ai insights [period]  AI-powered insights (day/week/month)
  uroboro ai suggest <context>  Get capture suggestions based on context

INFORMATION:
  uroboro ai stats              Show AI feature statistics
  uroboro ai help               Show this help message

EXAMPLES:
  uroboro ai search "PostHog integration"
  uroboro ai insights week
  uroboro ai suggest "working on ESP32 project"

NOTE: Requires Ollama running locally with nomic-embed-text model installed.
Install with: ollama pull nomic-embed-text`)
	return nil
}

// setupAI initializes AI features
func setupAI(config CLIConfig) error {
	fmt.Println("🤖 Setting up uroboro AI features...")

	// Test Ollama connection first
	ai, err := NewAIManager(config.ToAIConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize AI manager: %w", err)
	}
	defer ai.Close()

	if err := ai.TestConnection(); err != nil {
		fmt.Printf("❌ Ollama connection failed: %v\n", err)
		fmt.Println("\nTo fix this:")
		fmt.Println("1. Install Ollama: curl -fsSL https://ollama.ai/install.sh | sh")
		fmt.Println("2. Pull embedding model: ollama pull nomic-embed-text")
		fmt.Println("3. Ensure Ollama is running: ollama serve")
		return err
	}

	fmt.Println("✅ Ollama connection successful")
	fmt.Println("✅ Vector database initialized")
	fmt.Println("✅ AI features ready!")
	fmt.Println("\nNext steps:")
	fmt.Println("  uroboro ai embed     # Embed existing captures")
	fmt.Println("  uroboro ai search    # Try semantic search")

	return nil
}

// embedCaptures embeds captures for semantic search
func embedCaptures(config CLIConfig, args []string) error {
	fmt.Println("🔄 Embedding captures for semantic search...")

	ai, err := NewAIManager(config.ToAIConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize AI manager: %w", err)
	}
	defer ai.Close()

	// Check if we should only embed new captures
	newOnly := len(args) > 0 && args[0] == "--new"

	if newOnly {
		fmt.Println("Embedding only new captures...")
	} else {
		fmt.Println("Embedding all captures...")
	}

	if err := ai.EmbedAllCaptures(); err != nil {
		return fmt.Errorf("failed to embed captures: %w", err)
	}

	// Show stats after embedding
	stats, err := ai.GetEmbeddingStats()
	if err != nil {
		fmt.Println("✅ Embedding completed (stats unavailable)")
		return nil
	}

	fmt.Printf("✅ Embedding completed!\n")
	fmt.Printf("   Total captures: %v\n", stats["total_captures"])
	fmt.Printf("   Embedded: %v (%.1f%% coverage)\n",
		stats["total_embeddings"],
		stats["embedding_coverage_percent"])

	return nil
}

// semanticSearch performs semantic search
func semanticSearch(config CLIConfig, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("search query required. Usage: uroboro ai search <query>")
	}

	query := strings.Join(args, " ")
	fmt.Printf("🔍 Searching for: %s\n\n", query)

	ai, err := NewAIManager(config.ToAIConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize AI manager: %w", err)
	}
	defer ai.Close()

	results, err := ai.SemanticSearch(query, 10)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No results found. Try:")
		fmt.Println("1. uroboro ai embed  # Ensure captures are embedded")
		fmt.Println("2. Different search terms")
		return nil
	}

	fmt.Printf("Found %d results:\n\n", len(results))
	for i, result := range results {
		fmt.Printf("🎯 Result %d (similarity: %.3f)\n", i+1, 1-result.Similarity)
		fmt.Printf("   ID: %d | Created: %s\n", result.CaptureID, result.CreatedAt)
		if result.Project != "" {
			fmt.Printf("   Project: %s", result.Project)
		}
		if result.Tags != "" {
			fmt.Printf(" | Tags: %s", result.Tags)
		}
		fmt.Println()

		// Truncate long content
		content := result.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		fmt.Printf("   %s\n\n", content)
	}

	return nil
}

// getInsights provides AI-powered insights
func getInsights(config CLIConfig, args []string) error {
	timeframe := "week"
	if len(args) > 0 && (args[0] == "day" || args[0] == "week" || args[0] == "month") {
		timeframe = args[0]
	}

	fmt.Printf("🧠 Generating insights for the last %s...\n\n", timeframe)

	ai, err := NewAIManager(config.ToAIConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize AI manager: %w", err)
	}
	defer ai.Close()

	insights, err := ai.GetInsights(timeframe)
	if err != nil {
		return fmt.Errorf("failed to generate insights: %w", err)
	}

	fmt.Println(insights)
	return nil
}

// getSuggestions provides capture suggestions
func getSuggestions(config CLIConfig, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("context required. Usage: uroboro ai suggest <context>")
	}

	context := strings.Join(args, " ")
	fmt.Printf("💡 Getting suggestions for: %s\n\n", context)

	ai, err := NewAIManager(config.ToAIConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize AI manager: %w", err)
	}
	defer ai.Close()

	suggestions, err := ai.SuggestCaptures(context)
	if err != nil {
		return fmt.Errorf("failed to get suggestions: %w", err)
	}

	fmt.Println("Suggested captures:")
	for i, suggestion := range suggestions {
		fmt.Printf("  %d. %s\n", i+1, suggestion)
	}

	return nil
}

// showAIStats displays AI feature statistics
func showAIStats(config CLIConfig) error {
	fmt.Println("📊 uroboro AI Statistics")

	ai, err := NewAIManager(config.ToAIConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize AI manager: %w", err)
	}
	defer ai.Close()

	stats, err := ai.GetEmbeddingStats()
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	// Format stats nicely
	fmt.Printf("AI Status: %s\n", map[bool]string{true: "✅ Enabled", false: "❌ Disabled"}[stats["ai_enabled"].(bool)])
	fmt.Printf("Embedding Model: %s\n", stats["embed_model"])
	fmt.Printf("Total Captures: %v\n", stats["total_captures"])
	fmt.Printf("Embedded Captures: %v\n", stats["total_embeddings"])
	fmt.Printf("Coverage: %.1f%%\n", stats["embedding_coverage_percent"])

	// Test connection
	fmt.Print("Ollama Connection: ")
	if err := ai.TestConnection(); err != nil {
		fmt.Printf("❌ Failed (%v)\n", err)
	} else {
		fmt.Println("✅ Connected")
	}

	return nil
}

// testAIConnection tests the AI setup
func testAIConnection(config CLIConfig) error {
	fmt.Println("🔧 Testing AI setup...")

	ai, err := NewAIManager(config.ToAIConfig())
	if err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		return err
	}
	defer ai.Close()

	fmt.Println("✅ Database connection successful")

	if err := ai.TestConnection(); err != nil {
		fmt.Printf("❌ Ollama connection failed: %v\n", err)
		fmt.Println("\nTroubleshooting:")
		fmt.Println("1. Check if Ollama is running: ollama list")
		fmt.Println("2. Start Ollama if needed: ollama serve")
		fmt.Println("3. Install models: ollama pull nomic-embed-text")
		return err
	}

	fmt.Println("✅ Ollama connection successful")

	// Test embedding generation
	fmt.Print("Testing embedding generation... ")
	_, err = ai.GetEmbedding("test embedding")
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return err
	}

	fmt.Println("✅ Working")
	fmt.Println("\n🎉 All AI features are working correctly!")

	return nil
}

// FormatSearchResults formats search results as JSON
func FormatSearchResults(results []SearchResult) (string, error) {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format results: %w", err)
	}
	return string(jsonData), nil
}

// ParseLimit parses limit from command line args
func ParseLimit(args []string, defaultLimit int) int {
	for i, arg := range args {
		if arg == "--limit" && i+1 < len(args) {
			if limit, err := strconv.Atoi(args[i+1]); err == nil && limit > 0 {
				return limit
			}
		}
	}
	return defaultLimit
}
