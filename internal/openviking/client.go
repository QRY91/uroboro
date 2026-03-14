package openviking

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client talks to an OpenViking HTTP server.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates an OpenViking client.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

// Response is the standard OpenViking API envelope.
type Response struct {
	Status string          `json:"status"`
	Result json.RawMessage `json:"result"`
	Error  *ErrorInfo      `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SessionResult from POST /api/v1/sessions
type SessionResult struct {
	SessionID string `json:"session_id"`
}

// CommitResult from POST /api/v1/sessions/{id}/commit
type CommitResult struct {
	SessionID        string `json:"session_id"`
	MemoriesExtracted int   `json:"memories_extracted"`
	Archived         bool   `json:"archived"`
}

// MessageResult from POST /api/v1/sessions/{id}/messages
type MessageResult struct {
	SessionID    string `json:"session_id"`
	MessageCount int    `json:"message_count"`
}

func (c *Client) do(method, path string, body interface{}) (*Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	var apiResp Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if apiResp.Status == "error" && apiResp.Error != nil {
		return nil, fmt.Errorf("API error [%s]: %s", apiResp.Error.Code, apiResp.Error.Message)
	}

	return &apiResp, nil
}

// Health checks if the server is reachable.
func (c *Client) Health() error {
	_, err := c.do("GET", "/api/v1/observer/system", nil)
	return err
}

// CreateSession starts a new session for memory extraction.
func (c *Client) CreateSession() (*SessionResult, error) {
	resp, err := c.do("POST", "/api/v1/sessions", nil)
	if err != nil {
		return nil, err
	}
	var result SessionResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("parse session result: %w", err)
	}
	return &result, nil
}

// AddMessage adds a message to a session.
func (c *Client) AddMessage(sessionID, role, content string) error {
	body := map[string]string{
		"role":    role,
		"content": content,
	}
	_, err := c.do("POST", "/api/v1/sessions/"+sessionID+"/messages", body)
	return err
}

// AsyncCommitResult from a background commit.
type AsyncCommitResult struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	TaskID    string `json:"task_id"`
}

// TaskResult from GET /api/v1/tasks/{id}
type TaskResult struct {
	TaskID string          `json:"task_id"`
	Status string          `json:"status"` // pending, running, completed, failed
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// CommitSession triggers memory extraction on a session (blocking).
func (c *Client) CommitSession(sessionID string) (*CommitResult, error) {
	resp, err := c.do("POST", "/api/v1/sessions/"+sessionID+"/commit?wait=true", nil)
	if err != nil {
		return nil, err
	}
	var result CommitResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("parse commit result: %w", err)
	}
	return &result, nil
}

// CommitSessionAsync starts memory extraction in the background.
func (c *Client) CommitSessionAsync(sessionID string) (*AsyncCommitResult, error) {
	resp, err := c.do("POST", "/api/v1/sessions/"+sessionID+"/commit?wait=false", nil)
	if err != nil {
		return nil, err
	}
	var result AsyncCommitResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("parse async commit result: %w", err)
	}
	return &result, nil
}

// GetTask checks the status of a background task.
func (c *Client) GetTask(taskID string) (*TaskResult, error) {
	resp, err := c.do("GET", "/api/v1/tasks/"+taskID, nil)
	if err != nil {
		return nil, err
	}
	var result TaskResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("parse task result: %w", err)
	}
	return &result, nil
}

// WaitForTask polls until a background task completes or fails.
func (c *Client) WaitForTask(taskID string, pollInterval, timeout time.Duration) (*TaskResult, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		task, err := c.GetTask(taskID)
		if err != nil {
			return nil, err
		}
		switch task.Status {
		case "completed":
			return task, nil
		case "failed":
			return task, fmt.Errorf("task failed: %s", task.Error)
		}
		time.Sleep(pollInterval)
	}
	return nil, fmt.Errorf("task %s timed out after %v", taskID, timeout)
}

// DeleteSession cleans up a session.
func (c *Client) DeleteSession(sessionID string) error {
	_, err := c.do("DELETE", "/api/v1/sessions/"+sessionID, nil)
	return err
}
