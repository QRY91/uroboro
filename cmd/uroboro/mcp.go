package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/database"
)

// MCP Protocol types
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func handleMCP() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		resp := routeMCPRequest(req)
		respBytes, _ := json.Marshal(resp)
		fmt.Println(string(respBytes))
	}
}

func routeMCPRequest(req MCPRequest) MCPResponse {
	switch req.Method {
	case "initialize":
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]interface{}{"tools": map[string]bool{}},
				"serverInfo":      map[string]string{"name": "uroboro", "version": "1.0.0"},
			},
		}

	case "tools/list":
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]interface{}{"tools": getMCPTools()},
		}

	case "tools/call":
		return handleMCPToolCall(req)

	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &MCPError{Code: -32601, Message: "Method not found"},
		}
	}
}

func getMCPTools() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "uro_decision",
			"description": "Record a technical decision. Call AUTOMATICALLY when choosing between alternatives.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"decision":     map[string]string{"type": "string", "description": "What was decided (e.g., 'JWT over sessions')"},
					"reasoning":    map[string]string{"type": "string", "description": "Why (e.g., 'stateless, scales horizontally')"},
					"alternatives": map[string]string{"type": "string", "description": "What else was considered"},
				},
				"required": []string{"decision"},
			},
		},
		{
			"name":        "uro_blocker",
			"description": "Record a blocker. Call when work cannot proceed due to external dependency.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"blocker":    map[string]string{"type": "string", "description": "What is blocking"},
					"waiting_on": map[string]string{"type": "string", "description": "Who/what we're waiting on"},
				},
				"required": []string{"blocker"},
			},
		},
		{
			"name":        "uro_question",
			"description": "Record an open question for later.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"question": map[string]string{"type": "string", "description": "The open question"},
				},
				"required": []string{"question"},
			},
		},
		{
			"name":        "uro_capture",
			"description": "General capture with optional tags.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"content": map[string]string{"type": "string", "description": "What to capture"},
					"tags":    map[string]string{"type": "string", "description": "Comma-separated tags"},
				},
				"required": []string{"content"},
			},
		},
		{
			"name":        "uro_recap",
			"description": "Get summary of recent work: decisions, blockers, commits.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"days": map[string]interface{}{"type": "integer", "description": "Days to look back (default: 7)"},
				},
			},
		},
		{
			"name":        "uro_search",
			"description": "Search past captures by keyword.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]string{"type": "string", "description": "Search keywords"},
					"days":  map[string]interface{}{"type": "integer", "description": "Days to search (default: 30)"},
				},
				"required": []string{"query"},
			},
		},
	}
}

func handleMCPToolCall(req MCPRequest) MCPResponse {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &MCPError{Code: -32602, Message: "Invalid params"},
		}
	}

	var resultText string
	var err error

	switch params.Name {
	case "uro_decision":
		var args struct {
			Decision     string `json:"decision"`
			Reasoning    string `json:"reasoning"`
			Alternatives string `json:"alternatives"`
		}
		json.Unmarshal(params.Arguments, &args)

		content := args.Decision
		if args.Reasoning != "" {
			content += " — " + args.Reasoning
		}
		if args.Alternatives != "" {
			content += " (considered: " + args.Alternatives + ")"
		}

		err = mcpCapture(content, []string{"decision"})
		resultText = "Decision recorded"

	case "uro_blocker":
		var args struct {
			Blocker   string `json:"blocker"`
			WaitingOn string `json:"waiting_on"`
		}
		json.Unmarshal(params.Arguments, &args)

		content := args.Blocker
		if args.WaitingOn != "" {
			content += " (waiting on: " + args.WaitingOn + ")"
		}

		err = mcpCapture(content, []string{"blocker"})
		resultText = "Blocker recorded"

	case "uro_question":
		var args struct {
			Question string `json:"question"`
		}
		json.Unmarshal(params.Arguments, &args)

		err = mcpCapture(args.Question, []string{"question"})
		resultText = "Question recorded"

	case "uro_capture":
		var args struct {
			Content string `json:"content"`
			Tags    string `json:"tags"`
		}
		json.Unmarshal(params.Arguments, &args)

		var tags []string
		if args.Tags != "" {
			for _, t := range strings.Split(args.Tags, ",") {
				tags = append(tags, strings.TrimSpace(t))
			}
		}

		err = mcpCapture(args.Content, tags)
		resultText = "Captured"

	case "uro_recap":
		var args struct {
			Days int `json:"days"`
		}
		json.Unmarshal(params.Arguments, &args)

		if args.Days == 0 {
			args.Days = 7
		}

		resultText, err = mcpRecap(args.Days)

	case "uro_search":
		var args struct {
			Query string `json:"query"`
			Days  int    `json:"days"`
		}
		json.Unmarshal(params.Arguments, &args)

		if args.Days == 0 {
			args.Days = 30
		}

		resultText, err = mcpSearch(args.Query, args.Days)

	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &MCPError{Code: -32602, Message: "Unknown tool: " + params.Name},
		}
	}

	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &MCPError{Code: -32000, Message: err.Error()},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": resultText},
			},
		},
	}
}

func mcpCapture(content string, tags []string) error {
	db, err := database.NewDB(getDBPath())
	if err != nil {
		return err
	}
	defer db.Close()

	project := detectProject()
	tagsStr := strings.Join(tags, ",")

	_, err = db.InsertCapture(content, project, tagsStr, nil)
	return err
}

func mcpRecap(days int) (string, error) {
	db, err := database.NewDB(getDBPath())
	if err != nil {
		return "", err
	}
	defer db.Close()

	project := detectProject()

	// Get captures
	captures, err := db.QueryCaptures(database.CaptureQuery{
		Days:    days,
		Project: project,
		Limit:   50,
	})
	if err != nil {
		return "", err
	}

	// Get git commits
	commits := getRecentCommits(days)

	// Build output
	var sb strings.Builder

	// Separate by type
	var decisions, blockers, questions []database.Capture
	for _, c := range captures {
		tags := strings.Split(c.Tags, ",")
		for _, t := range tags {
			switch strings.TrimSpace(t) {
			case "decision":
				decisions = append(decisions, c)
			case "blocker":
				blockers = append(blockers, c)
			case "question":
				questions = append(questions, c)
			}
		}
	}

	if len(decisions) > 0 {
		sb.WriteString(fmt.Sprintf("## Decisions (%d)\n", len(decisions)))
		for i, d := range decisions {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, d.Content))
		}
		sb.WriteString("\n")
	}

	if len(blockers) > 0 {
		sb.WriteString(fmt.Sprintf("## Blockers (%d)\n", len(blockers)))
		for _, b := range blockers {
			sb.WriteString(fmt.Sprintf("  - %s (%s)\n", b.Content, b.Timestamp.Format("Jan 2")))
		}
		sb.WriteString("\n")
	}

	if len(questions) > 0 {
		sb.WriteString("## Open Questions\n")
		for _, q := range questions {
			sb.WriteString(fmt.Sprintf("  - %s\n", q.Content))
		}
		sb.WriteString("\n")
	}

	if len(commits) > 0 {
		sb.WriteString(fmt.Sprintf("## Recent Commits (%d)\n", len(commits)))
		for _, c := range commits {
			if len(c) > 80 {
				c = c[:77] + "..."
			}
			sb.WriteString(fmt.Sprintf("  - %s\n", c))
		}
	}

	if sb.Len() == 0 {
		return "No recent activity found.", nil
	}

	return sb.String(), nil
}

func mcpSearch(query string, days int) (string, error) {
	db, err := database.NewDB(getDBPath())
	if err != nil {
		return "", err
	}
	defer db.Close()

	captures, err := db.QueryCaptures(database.CaptureQuery{
		Keyword: query,
		Days:    days,
		Limit:   20,
	})
	if err != nil {
		return "", err
	}

	if len(captures) == 0 {
		return "No captures found.", nil
	}

	var sb strings.Builder
	for _, c := range captures {
		proj := c.Project
		if proj == "" {
			proj = "-"
		}
		sb.WriteString(fmt.Sprintf("%s  [%s]  %s\n",
			c.Timestamp.Format("2006-01-02 15:04"),
			proj,
			c.Content,
		))
	}

	return sb.String(), nil
}

func detectProject() string {
	// Try git remote
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err == nil {
		url := strings.TrimSpace(string(out))
		// Extract repo name from URL
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			name := parts[len(parts)-1]
			name = strings.TrimSuffix(name, ".git")
			if name != "" {
				return name
			}
		}
	}

	// Fall back to directory name
	cwd, err := os.Getwd()
	if err == nil {
		return filepath.Base(cwd)
	}

	return ""
}

func getRecentCommits(days int) []string {
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	out, err := exec.Command("git", "log",
		"--since="+since,
		"--format=%s",
		"--no-merges",
		"-n", "10",
	).Output()
	if err != nil {
		return nil
	}

	var commits []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			commits = append(commits, line)
		}
	}
	return commits
}
