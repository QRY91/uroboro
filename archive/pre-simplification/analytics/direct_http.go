package analytics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// DirectHTTPClient sends events directly to PostHog HTTP API
type DirectHTTPClient struct {
	apiKey string
	host   string
	userID string
	client *http.Client
	debug  bool
}

// DirectEvent represents a PostHog event for direct HTTP sending
type DirectEvent struct {
	APIKey     string                 `json:"api_key"`
	Event      string                 `json:"event"`
	DistinctID string                 `json:"distinct_id"`
	Properties map[string]interface{} `json:"properties"`
	Timestamp  string                 `json:"timestamp,omitempty"`
}

// DirectBatchRequest represents a batch of events
type DirectBatchRequest struct {
	APIKey              string        `json:"api_key"`
	Batch               []DirectEvent `json:"batch"`
	HistoricalMigration bool          `json:"historical_migration"`
	SentAt              string        `json:"sentAt"`
}

// NewDirectHTTPClient creates a new direct HTTP client
func NewDirectHTTPClient(apiKey, host, userID string, debug bool) *DirectHTTPClient {
	return &DirectHTTPClient{
		apiKey: apiKey,
		host:   host,
		userID: userID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		debug: debug,
	}
}

// SendEvent sends a single event directly to PostHog
func (d *DirectHTTPClient) SendEvent(eventName string, properties map[string]interface{}) error {
	if d.debug {
		log.Printf("🔍 DIRECT: Sending event %s to %s", eventName, d.host)
	}

	// Create the event payload exactly like the Python test
	event := DirectEvent{
		APIKey:     d.apiKey,
		Event:      eventName,
		DistinctID: d.userID,
		Properties: properties,
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
	}

	// Send to capture endpoint
	return d.sendToCapture(event)
}

// SendBatchEvent sends an event as part of a batch (like the Go SDK does)
func (d *DirectHTTPClient) SendBatchEvent(eventName string, properties map[string]interface{}) error {
	if d.debug {
		log.Printf("🔍 DIRECT: Sending batch event %s to %s", eventName, d.host)
	}

	// Create event without API key (batch format)
	event := DirectEvent{
		Event:      eventName,
		DistinctID: d.userID,
		Properties: properties,
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
	}

	// Create batch request
	batch := DirectBatchRequest{
		APIKey:              d.apiKey,
		Batch:               []DirectEvent{event},
		HistoricalMigration: false,
		SentAt:              time.Now().UTC().Format(time.RFC3339Nano),
	}

	return d.sendToBatch(batch)
}

func (d *DirectHTTPClient) sendToCapture(event DirectEvent) error {
	url := fmt.Sprintf("%s/capture/", d.host)

	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if d.debug {
		log.Printf("🔍 DIRECT: POST %s", url)
		log.Printf("🔍 DIRECT: Payload: %s", string(jsonData))
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "uroboro-direct-http/1.0")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if d.debug {
		log.Printf("🔍 DIRECT: Response status: %d", resp.StatusCode)
		log.Printf("🔍 DIRECT: Response body: %s", string(body))
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("PostHog returned status %d: %s", resp.StatusCode, string(body))
	}

	if d.debug {
		log.Printf("✅ DIRECT: Event sent successfully via capture endpoint")
	}

	return nil
}

func (d *DirectHTTPClient) sendToBatch(batch DirectBatchRequest) error {
	url := fmt.Sprintf("%s/batch/", d.host)

	jsonData, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to marshal batch: %w", err)
	}

	if d.debug {
		log.Printf("🔍 DIRECT: POST %s", url)
		log.Printf("🔍 DIRECT: Batch payload: %s", string(jsonData))
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "uroboro-direct-http/1.0")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if d.debug {
		log.Printf("🔍 DIRECT: Response status: %d", resp.StatusCode)
		log.Printf("🔍 DIRECT: Response body: %s", string(body))
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("PostHog returned status %d: %s", resp.StatusCode, string(body))
	}

	if d.debug {
		log.Printf("✅ DIRECT: Batch sent successfully via batch endpoint")
	}

	return nil
}

// TestConnection tests the connection with a simple event
func (d *DirectHTTPClient) TestConnection() error {
	testProperties := map[string]interface{}{
		"test":            true,
		"method":          "direct_http",
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
		"connection_test": true,
		"library":         "uroboro_direct_http",
	}

	log.Printf("🧪 DIRECT: Testing connection to PostHog...")

	// Try both capture and batch methods
	if err := d.SendEvent("uroboro_direct_http_test", testProperties); err != nil {
		log.Printf("❌ DIRECT: Capture method failed: %v", err)

		// Try batch method
		if err := d.SendBatchEvent("uroboro_direct_http_test_batch", testProperties); err != nil {
			return fmt.Errorf("both capture and batch methods failed: %w", err)
		}
		log.Printf("✅ DIRECT: Batch method succeeded")
		return nil
	}

	log.Printf("✅ DIRECT: Capture method succeeded")
	return nil
}
