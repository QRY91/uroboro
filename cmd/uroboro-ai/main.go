package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/QRY91/uroboro/ai"
)

func main() {
	fmt.Println("🤖 uroboro AI Features Test Program")
	fmt.Println("===================================")

	// Get command line arguments
	args := os.Args[1:]

	if len(args) == 0 {
		showUsage()
		return
	}

	// Handle AI commands
	if args[0] == "ai" && len(args) > 1 {
		if err := ai.RunAICommand(args[1:]); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle other test commands
	switch args[0] {
	case "test-full":
		runFullTest()
	case "test-embedding":
		testEmbedding()
	case "test-search":
		testSearch(args[1:])
	case "demo":
		runDemo()
	case "setup-test-db":
		setupTestDatabase()
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		showUsage()
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Println(`Usage:
  uroboro-ai ai <command>          Run AI commands (setup, embed, search, etc.)
  uroboro-ai test-full             Run comprehensive AI test suite
  uroboro-ai test-embedding        Test embedding generation
  uroboro-ai test-search <query>   Test semantic search
  uroboro-ai demo                  Run interactive demo
  uroboro-ai setup-test-db         Create test database with sample data

AI Commands:
  uroboro-ai ai setup              Initialize AI features
  uroboro-ai ai embed              Embed all captures
  uroboro-ai ai search <query>     Search captures semantically
  uroboro-ai ai insights [period]  Generate insights (day/week/month)
  uroboro-ai ai stats              Show statistics
  uroboro-ai ai test               Test AI connection

Examples:
  uroboro-ai ai setup
  uroboro-ai ai search "PostHog integration"
  uroboro-ai demo`)
}

func runFullTest() {
	fmt.Println("🧪 Running comprehensive AI tests...")

	config := ai.DefaultCLIConfig()

	// Override database path for testing
	homeDir, _ := os.UserHomeDir()
	testDbPath := filepath.Join(homeDir, ".local/share/uroboro/test_uroboro.sqlite")
	config.DatabasePath = testDbPath

	fmt.Printf("Using test database: %s\n", testDbPath)

	// Test 1: AI Manager initialization
	fmt.Print("1. Testing AI Manager initialization... ")
	aiManager, err := ai.NewAIManager(config.ToAIConfig())
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		return
	}
	fmt.Println("✅ PASSED")

	// Test 2: Ollama connection
	fmt.Print("2. Testing Ollama connection... ")
	if err := aiManager.TestConnection(); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		fmt.Println("   Make sure Ollama is running and nomic-embed-text is installed:")
		fmt.Println("   ollama pull nomic-embed-text")
		aiManager.Close()
		return
	}
	fmt.Println("✅ PASSED")

	// Test 3: Embedding generation
	fmt.Print("3. Testing embedding generation... ")
	embedding, err := aiManager.GetEmbedding("test embedding for uroboro AI integration")
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		aiManager.Close()
		return
	}
	if len(embedding) != 768 {
		fmt.Printf("❌ FAILED: Expected 768 dimensions, got %d\n", len(embedding))
		aiManager.Close()
		return
	}
	fmt.Println("✅ PASSED")

	// Test 4: Statistics
	fmt.Print("4. Testing statistics retrieval... ")
	stats, err := aiManager.GetEmbeddingStats()
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		aiManager.Close()
		return
	}
	fmt.Printf("✅ PASSED (%.1f%% coverage)\n", stats["embedding_coverage_percent"])

	aiManager.Close()

	fmt.Println("\n🎉 All tests passed! AI features are working correctly.")
	fmt.Println("\nNext steps:")
	fmt.Println("  uroboro-ai ai embed     # Embed your actual captures")
	fmt.Println("  uroboro-ai ai search    # Try semantic search")
}

func testEmbedding() {
	fmt.Println("🔬 Testing embedding generation...")

	config := ai.DefaultCLIConfig()
	aiManager, err := ai.NewAIManager(config.ToAIConfig())
	if err != nil {
		fmt.Printf("❌ Failed to initialize AI: %v\n", err)
		return
	}
	defer aiManager.Close()

	testTexts := []string{
		"PostHog integration with local AI vector search",
		"ESP32 hardware development for IoT projects",
		"SQLite vector database for semantic search",
		"QRY methodology for systematic development",
		"Local-first AI to reduce cloud costs",
	}

	fmt.Printf("Generating embeddings for %d test texts...\n", len(testTexts))

	for i, text := range testTexts {
		fmt.Printf("  %d. Embedding: '%s'\n", i+1, text)

		embedding, err := aiManager.GetEmbedding(text)
		if err != nil {
			fmt.Printf("     ❌ Failed: %v\n", err)
			continue
		}

		fmt.Printf("     ✅ Generated %d-dimensional embedding\n", len(embedding))

		// Show first few values for verification
		if len(embedding) >= 5 {
			fmt.Printf("     First 5 values: [%.4f, %.4f, %.4f, %.4f, %.4f...]\n",
				embedding[0], embedding[1], embedding[2], embedding[3], embedding[4])
		}
	}

	fmt.Println("\n✅ Embedding test completed!")
}

func testSearch(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Search query required")
		fmt.Println("Usage: uroboro-ai test-search <query>")
		return
	}

	query := args[0]
	fmt.Printf("🔍 Testing semantic search for: '%s'\n", query)

	config := ai.DefaultCLIConfig()
	aiManager, err := ai.NewAIManager(config.ToAIConfig())
	if err != nil {
		fmt.Printf("❌ Failed to initialize AI: %v\n", err)
		return
	}
	defer aiManager.Close()

	results, err := aiManager.SemanticSearch(query, 5)
	if err != nil {
		fmt.Printf("❌ Search failed: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("No results found. Make sure captures are embedded first:")
		fmt.Println("  uroboro-ai ai embed")
		return
	}

	fmt.Printf("Found %d results:\n\n", len(results))
	for i, result := range results {
		similarity := 1 - result.Similarity // Convert distance to similarity
		fmt.Printf("🎯 Result %d (%.1f%% similarity)\n", i+1, similarity*100)
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
		if len(content) > 150 {
			content = content[:150] + "..."
		}
		fmt.Printf("   %s\n\n", content)
	}
}

func runDemo() {
	fmt.Println("🎬 uroboro AI Demo")
	fmt.Println("This demo showcases local-first AI features integrated with uroboro")
	fmt.Println()

	config := ai.DefaultCLIConfig()

	// Show configuration
	fmt.Println("📋 Configuration:")
	fmt.Printf("  Database: %s\n", config.DatabasePath)
	fmt.Printf("  Ollama URL: %s\n", config.OllamaURL)
	fmt.Printf("  Embed Model: %s\n", config.EmbedModel)
	fmt.Printf("  Chat Model: %s\n", config.ChatModel)
	fmt.Println()

	// Initialize AI
	fmt.Print("🤖 Initializing AI manager... ")
	aiManager, err := ai.NewAIManager(config.ToAIConfig())
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	defer aiManager.Close()
	fmt.Println("✅ Ready")

	// Test connection
	fmt.Print("🔗 Testing Ollama connection... ")
	if err := aiManager.TestConnection(); err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		fmt.Println("\nSetup instructions:")
		fmt.Println("1. Install Ollama: curl -fsSL https://ollama.ai/install.sh | sh")
		fmt.Println("2. Pull model: ollama pull nomic-embed-text")
		fmt.Println("3. Start service: ollama serve")
		return
	}
	fmt.Println("✅ Connected")

	// Show stats
	fmt.Print("📊 Getting current statistics... ")
	stats, err := aiManager.GetEmbeddingStats()
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	fmt.Println("✅ Retrieved")

	fmt.Printf("\n📈 Current Stats:\n")
	fmt.Printf("  Total Captures: %v\n", stats["total_captures"])
	fmt.Printf("  Embedded: %v (%.1f%% coverage)\n",
		stats["total_embeddings"],
		stats["embedding_coverage_percent"])

	// Demo embedding
	fmt.Println("\n🧪 Demo: Generating embedding...")
	demoText := "PostHog AI integration with local vector search using SQLite-vec"
	fmt.Printf("Text: '%s'\n", demoText)

	embedding, err := aiManager.GetEmbedding(demoText)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated %d-dimensional embedding\n", len(embedding))
	fmt.Printf("Sample values: [%.4f, %.4f, %.4f, ...]\n",
		embedding[0], embedding[1], embedding[2])

	// Demo search if embeddings exist
	if stats["total_embeddings"].(int) > 0 {
		fmt.Println("\n🔍 Demo: Semantic search...")
		results, err := aiManager.SemanticSearch("AI integration", 3)
		if err != nil {
			fmt.Printf("❌ Search failed: %v\n", err)
		} else if len(results) > 0 {
			fmt.Printf("Found %d similar captures:\n", len(results))
			for i, result := range results {
				similarity := (1 - result.Similarity) * 100
				fmt.Printf("  %d. %.1f%% match - %s\n", i+1, similarity,
					truncateString(result.Content, 60))
			}
		} else {
			fmt.Println("No similar captures found")
		}
	}

	fmt.Println("\n🎉 Demo completed!")
	fmt.Println("\nTry these commands:")
	fmt.Println("  uroboro-ai ai embed              # Embed all captures")
	fmt.Println("  uroboro-ai ai search 'AI'        # Search for AI-related captures")
	fmt.Println("  uroboro-ai ai insights week      # Get weekly insights")
}

func setupTestDatabase() {
	fmt.Println("🗄️  Setting up test database with sample data...")

	homeDir, _ := os.UserHomeDir()
	testDbPath := filepath.Join(homeDir, ".local/share/uroboro/test_uroboro.sqlite")

	// Create directory if it doesn't exist
	os.MkdirAll(filepath.Dir(testDbPath), 0755)

	// Remove existing test database
	os.Remove(testDbPath)

	fmt.Printf("Creating test database: %s\n", testDbPath)

	// This would normally use the actual uroboro database creation logic
	// For now, we'll create a simple version

	config := ai.AIConfig{
		DatabasePath: testDbPath,
		OllamaURL:    "http://localhost:11434",
		EmbedModel:   "nomic-embed-text",
		ChatModel:    "llama3.2:1b",
		Enabled:      false, // Don't load extension yet
	}

	aiManager, err := ai.NewAIManager(config)
	if err != nil {
		fmt.Printf("❌ Failed to create AI manager: %v\n", err)
		return
	}
	defer aiManager.Close()

	// Create captures table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS captures (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		tags TEXT,
		project TEXT
	);`

	if err := aiManager.ExecSQL(createTableSQL); err != nil {
		fmt.Printf("❌ Failed to create table: %v\n", err)
		return
	}

	// Insert sample data
	sampleCaptures := []struct {
		content string
		tags    string
		project string
	}{
		{"PostHog Max AI architecture analysis complete - Multi-LLM strategy, RAG over fine-tuning insights", "AI,posthog,architecture", "posthog-integration"},
		{"SQLite-vec integration working perfectly - local vector search with 768-dim embeddings", "AI,sqlite,vector,local-first", "uroboro"},
		{"Ollama setup complete - nomic-embed-text model ready for local embeddings", "AI,ollama,embeddings,local", "uroboro"},
		{"ESP32 DeskHog development continues - hardware analytics integration planned", "hardware,esp32,analytics", "hardware-dev"},
		{"Cost analysis: Local AI saves $25k vs OpenAI for 1TB embeddings", "AI,cost,analysis,savings", "strategy"},
		{"QRY methodology applied to AI development - systematic local-first approach", "methodology,AI,qry,systematic", "qry-methodology"},
		{"Vector similarity search accuracy testing - cosine distance working well", "AI,vector,similarity,testing", "uroboro"},
		{"AI-enhanced uroboro features planned: semantic search, insights, suggestions", "AI,features,roadmap,uroboro", "uroboro"},
		{"PostHog application timeline on track - AI integration demonstrates technical depth", "career,posthog,timeline,progress", "career-strategy"},
		{"Local-first AI philosophy: enhance not compete with PostHog", "AI,philosophy,integration,strategy", "posthog-integration"},
	}

	for _, capture := range sampleCaptures {
		insertSQL := `INSERT INTO captures (content, tags, project) VALUES (?, ?, ?)`
		if err := aiManager.ExecSQL(insertSQL, capture.content, capture.tags, capture.project); err != nil {
			fmt.Printf("❌ Failed to insert sample data: %v\n", err)
			return
		}
	}

	fmt.Printf("✅ Test database created with %d sample captures\n", len(sampleCaptures))
	fmt.Println("\nNext steps:")
	fmt.Println("  uroboro-ai ai setup     # Initialize AI features")
	fmt.Println("  uroboro-ai ai embed     # Embed the sample captures")
	fmt.Println("  uroboro-ai demo         # Run interactive demo")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
