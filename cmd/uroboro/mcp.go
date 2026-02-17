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

	urocontext "github.com/QRY91/uroboro/internal/context"
	"github.com/QRY91/uroboro/internal/database"
	"github.com/QRY91/uroboro/internal/distill"
	"github.com/QRY91/uroboro/internal/promptprofile"
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
					"timestamp":    map[string]string{"type": "string", "description": "ISO timestamp for retroactive logging (e.g., 2024-01-15T14:30:00)"},
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
					"timestamp":  map[string]string{"type": "string", "description": "ISO timestamp for retroactive logging (e.g., 2024-01-15T14:30:00)"},
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
					"question":  map[string]string{"type": "string", "description": "The open question"},
					"timestamp": map[string]string{"type": "string", "description": "ISO timestamp for retroactive logging (e.g., 2024-01-15T14:30:00)"},
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
					"content":   map[string]string{"type": "string", "description": "What to capture"},
					"tags":      map[string]string{"type": "string", "description": "Comma-separated tags"},
					"timestamp": map[string]string{"type": "string", "description": "ISO timestamp for retroactive logging (e.g., 2024-01-15T14:30:00)"},
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
					"days":   map[string]interface{}{"type": "integer", "description": "Days to look back (default: 7)"},
					"branch": map[string]string{"type": "string", "description": "Filter by git branch"},
				},
			},
		},
		{
			"name":        "uro_search",
			"description": "Search past captures by keyword.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":  map[string]string{"type": "string", "description": "Search keywords"},
					"days":   map[string]interface{}{"type": "integer", "description": "Days to search (default: 30)"},
					"branch": map[string]string{"type": "string", "description": "Filter by git branch"},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "uro_distill",
			"description": "Extract style signal data from git commits and uroboro captures. Writes JSONL to file and returns file path with summary counts.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source":    map[string]string{"type": "string", "description": "What to extract: git, uro, or all (default: all)"},
					"repo":      map[string]string{"type": "string", "description": "Path to git repository (default: current directory)"},
					"project":   map[string]string{"type": "string", "description": "Filter uroboro captures by project"},
					"days":      map[string]interface{}{"type": "integer", "description": "Limit to last N days"},
					"since":     map[string]string{"type": "string", "description": "Limit to after date (YYYY-MM-DD)"},
					"correlate": map[string]interface{}{"type": "boolean", "description": "Join git↔uro captures by ±30min window"},
				},
			},
		},
		{
			"name":        "uro_prompt_profile",
			"description": "Analyze Claude Code prompting patterns. Returns statistics summary or writes JSONL extract to file.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project": map[string]string{"type": "string", "description": "Filter by project name"},
					"days":    map[string]interface{}{"type": "integer", "description": "Only include sessions modified within last N days"},
					"extract": map[string]interface{}{"type": "boolean", "description": "Output JSONL extract to file instead of stats (returns file path)"},
				},
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
			Timestamp    string `json:"timestamp"`
		}
		json.Unmarshal(params.Arguments, &args)

		content := args.Decision
		if args.Reasoning != "" {
			content += " — " + args.Reasoning
		}
		if args.Alternatives != "" {
			content += " (considered: " + args.Alternatives + ")"
		}

		err = mcpCapture(content, []string{"decision"}, args.Timestamp)
		resultText = "Decision recorded"

	case "uro_blocker":
		var args struct {
			Blocker   string `json:"blocker"`
			WaitingOn string `json:"waiting_on"`
			Timestamp string `json:"timestamp"`
		}
		json.Unmarshal(params.Arguments, &args)

		content := args.Blocker
		if args.WaitingOn != "" {
			content += " (waiting on: " + args.WaitingOn + ")"
		}

		err = mcpCapture(content, []string{"blocker"}, args.Timestamp)
		resultText = "Blocker recorded"

	case "uro_question":
		var args struct {
			Question  string `json:"question"`
			Timestamp string `json:"timestamp"`
		}
		json.Unmarshal(params.Arguments, &args)

		err = mcpCapture(args.Question, []string{"question"}, args.Timestamp)
		resultText = "Question recorded"

	case "uro_capture":
		var args struct {
			Content   string `json:"content"`
			Tags      string `json:"tags"`
			Timestamp string `json:"timestamp"`
		}
		json.Unmarshal(params.Arguments, &args)

		var tags []string
		if args.Tags != "" {
			for _, t := range strings.Split(args.Tags, ",") {
				tags = append(tags, strings.TrimSpace(t))
			}
		}

		err = mcpCapture(args.Content, tags, args.Timestamp)
		resultText = "Captured"

	case "uro_recap":
		var args struct {
			Days   int    `json:"days"`
			Branch string `json:"branch"`
		}
		json.Unmarshal(params.Arguments, &args)

		if args.Days == 0 {
			args.Days = 7
		}

		resultText, err = mcpRecap(args.Days, args.Branch)

	case "uro_search":
		var args struct {
			Query  string `json:"query"`
			Days   int    `json:"days"`
			Branch string `json:"branch"`
		}
		json.Unmarshal(params.Arguments, &args)

		if args.Days == 0 {
			args.Days = 30
		}

		resultText, err = mcpSearch(args.Query, args.Days, args.Branch)

	case "uro_distill":
		var args struct {
			Source    string `json:"source"`
			Repo     string `json:"repo"`
			Project  string `json:"project"`
			Days     int    `json:"days"`
			Since    string `json:"since"`
			Correlate bool  `json:"correlate"`
		}
		json.Unmarshal(params.Arguments, &args)

		resultText, err = mcpDistill(args.Source, args.Repo, args.Project, args.Days, args.Since, args.Correlate)

	case "uro_prompt_profile":
		var args struct {
			Project string `json:"project"`
			Days    int    `json:"days"`
			Extract bool   `json:"extract"`
		}
		json.Unmarshal(params.Arguments, &args)

		resultText, err = mcpPromptProfile(args.Project, args.Days, args.Extract)

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

func mcpCapture(content string, tags []string, timestampStr string) error {
	db, err := database.NewDB(getDBPath())
	if err != nil {
		return err
	}
	defer db.Close()

	detector := urocontext.NewProjectDetector()
	project := detectProject()
	branch := detector.DetectBranch()
	tagsStr := strings.Join(tags, ",")

	var ts *time.Time
	if timestampStr != "" {
		parsed, err := mcpParseTimestamp(timestampStr)
		if err != nil {
			return fmt.Errorf("invalid timestamp: %w", err)
		}
		ts = &parsed
	}

	_, err = db.InsertCapture(content, project, tagsStr, branch, ts)
	return err
}

func mcpParseTimestamp(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse '%s'", s)
}

func mcpRecap(days int, branch string) (string, error) {
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
		Branch:  branch,
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

func mcpSearch(query string, days int, branch string) (string, error) {
	db, err := database.NewDB(getDBPath())
	if err != nil {
		return "", err
	}
	defer db.Close()

	captures, err := db.QueryCaptures(database.CaptureQuery{
		Keyword: query,
		Days:    days,
		Branch:  branch,
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
		branchInfo := ""
		if c.Branch != "" {
			branchInfo = " @" + c.Branch
		}
		sb.WriteString(fmt.Sprintf("%s  [%s%s]  %s\n",
			c.Timestamp.Format("2006-01-02 15:04"),
			proj,
			branchInfo,
			c.Content,
		))
	}

	return sb.String(), nil
}

func detectProject() string {
	// Use shared project detection logic (guards against $HOME dotfiles repos)
	detector := urocontext.NewProjectDetector()
	if project := detector.DetectProject(); project != "" {
		return project
	}

	// Fall back to directory name
	cwd, err := os.Getwd()
	if err == nil {
		return filepath.Base(cwd)
	}

	return ""
}

func mcpDistill(source, repo, project string, days int, since string, correlate bool) (string, error) {
	if source == "" {
		source = "all"
	}
	if repo == "" {
		repo = "."
	}

	repoPath, err := filepath.Abs(repo)
	if err != nil {
		return "", fmt.Errorf("resolve repo path: %w", err)
	}

	// Parse time filters
	var sinceTime *time.Time
	if since != "" {
		t, err := mcpParseTimestamp(since)
		if err != nil {
			return "", fmt.Errorf("invalid since: %w", err)
		}
		sinceTime = &t
	}
	if days > 0 && sinceTime == nil {
		t := time.Now().AddDate(0, 0, -days)
		sinceTime = &t
	}

	// Extract
	var gitExtracts []distill.GitExtract
	var uroExtracts []distill.UroExtract

	if source == "git" || source == "all" {
		ext := distill.NewGitExtractor(repoPath, sinceTime)
		gitExtracts, err = ext.Extract()
		if err != nil {
			return "", fmt.Errorf("git extraction: %w", err)
		}
	}

	if source == "uro" || source == "all" {
		db, err := database.NewDB(getDBPath())
		if err != nil {
			return "", fmt.Errorf("open db: %w", err)
		}
		defer db.Close()
		ext := distill.NewUroExtractor(db, project, days, sinceTime)
		uroExtracts, err = ext.Extract()
		if err != nil {
			return "", fmt.Errorf("uro extraction: %w", err)
		}
	}

	if correlate && source == "all" {
		distill.Correlate(gitExtracts, uroExtracts)
	}

	// Write to file
	homeDir, _ := os.UserHomeDir()
	outDir := filepath.Join(homeDir, ".local", "share", "uroboro", "style-data")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	outPath := filepath.Join(outDir, fmt.Sprintf("distill-%s.jsonl", time.Now().Format("2006-01-02-150405")))
	f, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, g := range gitExtracts {
		enc.Encode(g)
	}
	for _, u := range uroExtracts {
		enc.Encode(u)
	}

	// Build summary
	correlated := 0
	if correlate {
		for _, u := range uroExtracts {
			if u.CorrelatedGitHash != "" {
				correlated++
			}
		}
	}

	summary := fmt.Sprintf("Distill complete: %d git + %d uro records\nOutput: %s", len(gitExtracts), len(uroExtracts), outPath)
	if correlate {
		summary += fmt.Sprintf("\nCorrelated: %d", correlated)
	}
	return summary, nil
}

func mcpPromptProfile(project string, days int, extract bool) (string, error) {
	projects, err := promptprofile.DiscoverProjects("")
	if err != nil {
		return "", fmt.Errorf("discover projects: %w", err)
	}

	// Filter by project name
	if project != "" {
		lower := strings.ToLower(project)
		var filtered []promptprofile.ProjectInfo
		for _, p := range projects {
			if strings.ToLower(p.ProjectName) == lower {
				filtered = append(filtered, p)
			}
		}
		projects = filtered
	}

	if len(projects) == 0 {
		return "No matching projects found.", nil
	}

	// Extract prompts
	var allPrompts []promptprofile.UserPrompt
	for _, p := range projects {
		prompts, err := promptprofile.ExtractFromProject(p, days)
		if err != nil {
			continue
		}
		allPrompts = append(allPrompts, prompts...)
	}

	if extract {
		// Write JSONL to file
		homeDir, _ := os.UserHomeDir()
		outDir := filepath.Join(homeDir, ".local", "share", "uroboro", "style-data")
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return "", fmt.Errorf("create output dir: %w", err)
		}

		outPath := filepath.Join(outDir, fmt.Sprintf("prompts-%s.jsonl", time.Now().Format("2006-01-02-150405")))
		f, err := os.Create(outPath)
		if err != nil {
			return "", fmt.Errorf("create output file: %w", err)
		}
		defer f.Close()

		enc := json.NewEncoder(f)
		for _, p := range allPrompts {
			enc.Encode(p)
		}

		return fmt.Sprintf("Extracted %d prompts from %d projects\nOutput: %s", len(allPrompts), len(projects), outPath), nil
	}

	// Default: return stats summary
	stats := promptprofile.Analyze(allPrompts)
	return promptprofile.FormatStats(stats), nil
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
