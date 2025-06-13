package main

import (
	"log"
	"os"

	"github.com/QRY91/uroboro/internal/analytics"
)

func main() {
	log.Printf("🧪 Starting direct HTTP PostHog test...")

	apiKey := os.Getenv("POSTHOG_API_KEY")
	host := os.Getenv("POSTHOG_HOST")

	if apiKey == "" {
		log.Fatal("❌ POSTHOG_API_KEY environment variable not set")
	}

	if host == "" {
		host = "https://us.posthog.com"
	}

	log.Printf("🔍 Testing with host: %s", host)
	log.Printf("🔍 API Key: %s...%s", apiKey[:8], apiKey[len(apiKey)-4:])

	// Create direct HTTP client
	client := analytics.NewDirectHTTPClient(apiKey, host, "uroboro_user_direct_test", true)

	// Test connection
	if err := client.TestConnection(); err != nil {
		log.Fatalf("❌ Direct HTTP test failed: %v", err)
	}

	log.Printf("✅ Direct HTTP test completed successfully!")
	log.Printf("🎯 Check your PostHog dashboard for events:")
	log.Printf("   - uroboro_direct_http_test")
	log.Printf("   - uroboro_direct_http_test_batch (if capture failed)")
}
