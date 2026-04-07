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
	"github.com/QRY91/uroboro/internal/conventions"
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
					"project":      map[string]string{"type": "string", "description": "Project name (overrides auto-detection from git)"},
					"tags":         map[string]string{"type": "string", "description": "Comma-separated additional tags (decision tag always added)"},
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
					"project":    map[string]string{"type": "string", "description": "Project name (overrides auto-detection from git)"},
					"tags":       map[string]string{"type": "string", "description": "Comma-separated additional tags (blocker tag always added)"},
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
					"project":   map[string]string{"type": "string", "description": "Project name (overrides auto-detection from git)"},
					"tags":      map[string]string{"type": "string", "description": "Comma-separated additional tags (question tag always added)"},
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
					"project":   map[string]string{"type": "string", "description": "Project name (overrides auto-detection from git)"},
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
					"days":    map[string]interface{}{"type": "integer", "description": "Days to look back (default: 7)"},
					"since":   map[string]string{"type": "string", "description": "Start date (YYYY-MM-DD or ISO timestamp)"},
					"until":   map[string]string{"type": "string", "description": "End date (YYYY-MM-DD or ISO timestamp)"},
					"project": map[string]string{"type": "string", "description": "Filter by project name (overrides auto-detection)"},
					"branch":  map[string]string{"type": "string", "description": "Filter by git branch"},
				},
			},
		},
		{
			"name":        "uro_search",
			"description": "Search past captures by keyword.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":   map[string]string{"type": "string", "description": "Search keywords"},
					"days":    map[string]interface{}{"type": "integer", "description": "Days to search (default: 0 = all time)"},
					"since":   map[string]string{"type": "string", "description": "Start date (YYYY-MM-DD or ISO timestamp)"},
					"until":   map[string]string{"type": "string", "description": "End date (YYYY-MM-DD or ISO timestamp)"},
					"tags":    map[string]string{"type": "string", "description": "Comma-separated tag filter (matches any)"},
					"project": map[string]string{"type": "string", "description": "Filter by project name"},
					"branch":  map[string]string{"type": "string", "description": "Filter by git branch"},
					"limit":   map[string]interface{}{"type": "integer", "description": "Max results (default: 50, max: 200)"},
				},
			},
		},
		{
			"name":        "uro_stats",
			"description": "Aggregate statistics about captures. Modes: 'tags' (tag usage and first/last seen), 'activity' (capture frequency over time), 'projects' (per-project breakdown).",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"mode":     map[string]string{"type": "string", "description": "One of: tags, activity, projects"},
					"since":    map[string]string{"type": "string", "description": "Start date (YYYY-MM-DD or ISO timestamp)"},
					"until":    map[string]string{"type": "string", "description": "End date (YYYY-MM-DD or ISO timestamp)"},
					"days":     map[string]interface{}{"type": "integer", "description": "Days to look back (default: 0 = all time)"},
					"project":  map[string]string{"type": "string", "description": "Filter by project"},
					"branch":   map[string]string{"type": "string", "description": "Filter by git branch"},
					"interval": map[string]string{"type": "string", "description": "For activity mode: day, week, or month (default: week)"},
				},
				"required": []string{"mode"},
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
		{
			"name":        "uro_conventions",
			"description": "Extract coding conventions from git history. Auto-scopes decisions to the repos being analyzed by default. Provide repos, scan_dir, or both.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repos":         map[string]string{"type": "string", "description": "Comma-separated absolute paths to git repositories"},
					"scan_dir":      map[string]string{"type": "string", "description": "Discover all git repos in this directory (combines with repos)"},
					"days":          map[string]interface{}{"type": "integer", "description": "Limit to last N days (default: 180)"},
					"since":         map[string]string{"type": "string", "description": "Limit to after date (YYYY-MM-DD)"},
					"correlate":     map[string]interface{}{"type": "boolean", "description": "Join git↔uro captures by ±30min window (default: true)"},
					"audit_dir":     map[string]string{"type": "string", "description": "Path to workspace audit markdown directory for supplementary context"},
					"all_decisions": map[string]interface{}{"type": "boolean", "description": "Include decisions from all projects instead of auto-scoping to repo names (default: false)"},
					"all_commits":   map[string]interface{}{"type": "boolean", "description": "Extract all commits, not just style-signal ones — recommended for richer analysis (default: false)"},
					"max_per_repo":  map[string]interface{}{"type": "integer", "description": "Cap commits per repo to prevent large repos dominating (default: 50; 0 = unlimited)"},
				},
			},
		},
		{
			"name":        "uro_enforcement",
			"description": "Configure uroboro enforcement hooks (pre-compact checkpoint, post-tool-use nudge). These hooks inject capture reminders during active work. Both are opt-in since they add tokens to context.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action":    map[string]string{"type": "string", "description": "Action: status, enable, disable, configure"},
					"hook":      map[string]string{"type": "string", "description": "Hook name: pre_compact, post_tool_nudge (required for enable/disable/configure)"},
					"threshold": map[string]interface{}{"type": "integer", "description": "For post_tool_nudge: number of Edit/Write/Bash calls between nudges (default: 15)"},
				},
				"required": []string{"action"},
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
			Project      string `json:"project"`
			Tags         string `json:"tags"`
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

		tags := []string{"decision"}
		if args.Tags != "" {
			for _, t := range strings.Split(args.Tags, ",") {
				if tt := strings.TrimSpace(t); tt != "" && tt != "decision" {
					tags = append(tags, tt)
				}
			}
		}

		err = mcpCapture(content, tags, args.Timestamp, args.Project)
		resultText = "Decision recorded"

	case "uro_blocker":
		var args struct {
			Blocker   string `json:"blocker"`
			WaitingOn string `json:"waiting_on"`
			Project   string `json:"project"`
			Tags      string `json:"tags"`
			Timestamp string `json:"timestamp"`
		}
		json.Unmarshal(params.Arguments, &args)

		content := args.Blocker
		if args.WaitingOn != "" {
			content += " (waiting on: " + args.WaitingOn + ")"
		}

		blockerTags := []string{"blocker"}
		if args.Tags != "" {
			for _, t := range strings.Split(args.Tags, ",") {
				if tt := strings.TrimSpace(t); tt != "" && tt != "blocker" {
					blockerTags = append(blockerTags, tt)
				}
			}
		}

		err = mcpCapture(content, blockerTags, args.Timestamp, args.Project)
		resultText = "Blocker recorded"

	case "uro_question":
		var args struct {
			Question  string `json:"question"`
			Project   string `json:"project"`
			Tags      string `json:"tags"`
			Timestamp string `json:"timestamp"`
		}
		json.Unmarshal(params.Arguments, &args)

		questionTags := []string{"question"}
		if args.Tags != "" {
			for _, t := range strings.Split(args.Tags, ",") {
				if tt := strings.TrimSpace(t); tt != "" && tt != "question" {
					questionTags = append(questionTags, tt)
				}
			}
		}

		err = mcpCapture(args.Question, questionTags, args.Timestamp, args.Project)
		resultText = "Question recorded"

	case "uro_capture":
		var args struct {
			Content   string `json:"content"`
			Tags      string `json:"tags"`
			Project   string `json:"project"`
			Timestamp string `json:"timestamp"`
		}
		json.Unmarshal(params.Arguments, &args)

		var tags []string
		if args.Tags != "" {
			for _, t := range strings.Split(args.Tags, ",") {
				tags = append(tags, strings.TrimSpace(t))
			}
		}

		err = mcpCapture(args.Content, tags, args.Timestamp, args.Project)
		resultText = "Captured"

	case "uro_recap":
		var args struct {
			Days    int    `json:"days"`
			Since   string `json:"since"`
			Until   string `json:"until"`
			Project string `json:"project"`
			Branch  string `json:"branch"`
		}
		json.Unmarshal(params.Arguments, &args)

		resultText, err = mcpRecapV2(args.Days, args.Since, args.Until, args.Project, args.Branch)

	case "uro_search":
		var args struct {
			Query   string `json:"query"`
			Days    int    `json:"days"`
			Since   string `json:"since"`
			Until   string `json:"until"`
			Tags    string `json:"tags"`
			Project string `json:"project"`
			Branch  string `json:"branch"`
			Limit   int    `json:"limit"`
		}
		json.Unmarshal(params.Arguments, &args)

		resultText, err = mcpSearchV2(args)

	case "uro_stats":
		var args struct {
			Mode     string `json:"mode"`
			Since    string `json:"since"`
			Until    string `json:"until"`
			Days     int    `json:"days"`
			Project  string `json:"project"`
			Branch   string `json:"branch"`
			Interval string `json:"interval"`
		}
		json.Unmarshal(params.Arguments, &args)

		resultText, err = mcpStats(args.Mode, args.Since, args.Until, args.Days, args.Project, args.Branch, args.Interval)

	case "uro_conventions":
		var args struct {
			Repos        string `json:"repos"`
			ScanDir      string `json:"scan_dir"`
			Days         int    `json:"days"`
			Since        string `json:"since"`
			Correlate    *bool  `json:"correlate"`
			AuditDir     string `json:"audit_dir"`
			AllDecisions bool   `json:"all_decisions"`
			AllCommits   bool   `json:"all_commits"`
			MaxPerRepo   int    `json:"max_per_repo"`
		}
		json.Unmarshal(params.Arguments, &args)

		resultText, err = mcpConventions(args.Repos, args.ScanDir, args.Days, args.Since, args.Correlate, args.AuditDir, args.AllDecisions, args.AllCommits, args.MaxPerRepo)

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

	case "uro_enforcement":
		var args struct {
			Action    string `json:"action"`
			Hook      string `json:"hook"`
			Threshold int    `json:"threshold"`
		}
		json.Unmarshal(params.Arguments, &args)

		resultText, err = mcpEnforcement(args.Action, args.Hook, args.Threshold)

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

func mcpCapture(content string, tags []string, timestampStr string, projectOverride ...string) error {
	db, err := database.NewDB(getDBPath())
	if err != nil {
		return err
	}
	defer db.Close()

	detector := urocontext.NewProjectDetector()
	project := detectProject()
	if len(projectOverride) > 0 && projectOverride[0] != "" {
		project = projectOverride[0]
	}
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

	_, err = db.InsertCapture(content, project, tagsStr, branch, localMachineHostname(), ts)
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

func mcpRecapV2(days int, sinceStr, untilStr, project, branch string) (string, error) {
	db, err := database.NewDB(getDBPath())
	if err != nil {
		return "", err
	}
	defer db.Close()

	if project == "" {
		project = detectProject()
	}

	q := database.CaptureQuery{
		Project: project,
		Branch:  branch,
		Limit:   50,
	}

	// Date range: since/until override days
	var gitSince, gitUntil string
	if sinceStr != "" {
		t, err := mcpParseTimestamp(sinceStr)
		if err != nil {
			return "", fmt.Errorf("invalid since: %w", err)
		}
		q.Since = &t
		gitSince = t.Format("2006-01-02")
	} else {
		if days == 0 {
			days = 7
		}
		q.Days = days
		gitSince = time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	}
	if untilStr != "" {
		t, err := mcpParseTimestamp(untilStr)
		if err != nil {
			return "", fmt.Errorf("invalid until: %w", err)
		}
		q.Until = &t
		gitUntil = t.Format("2006-01-02")
	}

	// Get captures
	captures, err := db.QueryCaptures(q)
	if err != nil {
		return "", err
	}

	// Get git commits
	commits := getCommitsBetween(gitSince, gitUntil, 30)

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

type mcpSearchArgs struct {
	Query   string `json:"query"`
	Days    int    `json:"days"`
	Since   string `json:"since"`
	Until   string `json:"until"`
	Tags    string `json:"tags"`
	Project string `json:"project"`
	Branch  string `json:"branch"`
	Limit   int    `json:"limit"`
}

func mcpSearchV2(args mcpSearchArgs) (string, error) {
	db, err := database.NewDB(getDBPath())
	if err != nil {
		return "", err
	}
	defer db.Close()

	q := database.CaptureQuery{
		Keyword: args.Query,
		Days:    args.Days,
		Branch:  args.Branch,
		Project: args.Project,
	}

	// Parse since/until
	if args.Since != "" {
		t, err := mcpParseTimestamp(args.Since)
		if err != nil {
			return "", fmt.Errorf("invalid since: %w", err)
		}
		q.Since = &t
	}
	if args.Until != "" {
		t, err := mcpParseTimestamp(args.Until)
		if err != nil {
			return "", fmt.Errorf("invalid until: %w", err)
		}
		q.Until = &t
	}

	// Parse tags
	if args.Tags != "" {
		for _, t := range strings.Split(args.Tags, ",") {
			if tag := strings.TrimSpace(t); tag != "" {
				q.Tags = append(q.Tags, tag)
			}
		}
	}

	// Limit: default 50, cap 200
	q.Limit = args.Limit
	if q.Limit <= 0 {
		q.Limit = 50
	}
	if q.Limit > 200 {
		q.Limit = 200
	}

	captures, err := db.QueryCaptures(q)
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
		tagsInfo := ""
		if c.Tags != "" {
			tagsInfo = "  {" + c.Tags + "}"
		}
		sb.WriteString(fmt.Sprintf("%s  [%s%s]%s  %s\n",
			c.Timestamp.Format("2006-01-02 15:04"),
			proj,
			branchInfo,
			tagsInfo,
			c.Content,
		))
	}

	return sb.String(), nil
}

func mcpStats(mode, sinceStr, untilStr string, days int, project, branch, interval string) (string, error) {
	db, err := database.NewDB(getDBPath())
	if err != nil {
		return "", err
	}
	defer db.Close()

	q := database.CaptureQuery{
		Days:    days,
		Project: project,
		Branch:  branch,
		Limit:   5000,
	}
	if sinceStr != "" {
		t, err := mcpParseTimestamp(sinceStr)
		if err != nil {
			return "", fmt.Errorf("invalid since: %w", err)
		}
		q.Since = &t
	}
	if untilStr != "" {
		t, err := mcpParseTimestamp(untilStr)
		if err != nil {
			return "", fmt.Errorf("invalid until: %w", err)
		}
		q.Until = &t
	}

	captures, err := db.QueryCaptures(q)
	if err != nil {
		return "", err
	}

	if len(captures) == 0 {
		return "No captures found.", nil
	}

	switch mode {
	case "tags":
		return statsTags(captures), nil
	case "activity":
		if interval == "" {
			interval = "week"
		}
		return statsActivity(captures, interval), nil
	case "projects":
		return statsProjects(captures), nil
	default:
		return "", fmt.Errorf("unknown mode: %s (use tags, activity, or projects)", mode)
	}
}

type tagStats struct {
	count     int
	firstSeen time.Time
	lastSeen  time.Time
}

func statsTags(captures []database.Capture) string {
	stats := make(map[string]*tagStats)

	for _, c := range captures {
		if c.Tags == "" {
			continue
		}
		for _, tag := range strings.Split(c.Tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			s, ok := stats[tag]
			if !ok {
				stats[tag] = &tagStats{count: 1, firstSeen: c.Timestamp, lastSeen: c.Timestamp}
			} else {
				s.count++
				if c.Timestamp.Before(s.firstSeen) {
					s.firstSeen = c.Timestamp
				}
				if c.Timestamp.After(s.lastSeen) {
					s.lastSeen = c.Timestamp
				}
			}
		}
	}

	// Sort by count descending
	type tagEntry struct {
		tag   string
		stats *tagStats
	}
	var entries []tagEntry
	for tag, s := range stats {
		entries = append(entries, tagEntry{tag, s})
	}
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].stats.count > entries[i].stats.count {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Tag usage (%d captures):\n\n", len(captures)))
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("  %-20s %4d  first: %s  last: %s\n",
			e.tag, e.stats.count,
			e.stats.firstSeen.Format("2006-01-02"),
			e.stats.lastSeen.Format("2006-01-02"),
		))
	}
	sb.WriteString(fmt.Sprintf("\nTotal: %d distinct tags\n", len(entries)))
	return sb.String()
}

func statsActivity(captures []database.Capture, interval string) string {
	type bucket struct {
		label     string
		decisions int
		blockers  int
		questions int
		other     int
	}

	buckets := make(map[string]*bucket)
	var bucketKeys []string

	for _, c := range captures {
		var key, label string
		switch interval {
		case "day":
			key = c.Timestamp.Format("2006-01-02")
			label = key
		case "month":
			key = c.Timestamp.Format("2006-01")
			label = key
		default: // week
			y, w := c.Timestamp.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", y, w)
			// Find Monday of this week for label
			mon := isoWeekStart(y, w)
			sun := mon.AddDate(0, 0, 6)
			label = fmt.Sprintf("%s  %s - %s", key, mon.Format("Jan 2"), sun.Format("Jan 2"))
		}

		b, ok := buckets[key]
		if !ok {
			b = &bucket{label: label}
			buckets[key] = b
			bucketKeys = append(bucketKeys, key)
		}

		// Classify
		classified := false
		for _, tag := range strings.Split(c.Tags, ",") {
			switch strings.TrimSpace(tag) {
			case "decision":
				b.decisions++
				classified = true
			case "blocker":
				b.blockers++
				classified = true
			case "question":
				b.questions++
				classified = true
			}
		}
		if !classified {
			b.other++
		}
	}

	// Sort keys chronologically (they're already sortable strings)
	for i := 0; i < len(bucketKeys); i++ {
		for j := i + 1; j < len(bucketKeys); j++ {
			if bucketKeys[j] < bucketKeys[i] {
				bucketKeys[i], bucketKeys[j] = bucketKeys[j], bucketKeys[i]
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Activity (%s, %d captures):\n\n", interval, len(captures)))
	for _, key := range bucketKeys {
		b := buckets[key]
		total := b.decisions + b.blockers + b.questions + b.other
		var parts []string
		if b.decisions > 0 {
			parts = append(parts, fmt.Sprintf("%d decision", b.decisions))
		}
		if b.blockers > 0 {
			parts = append(parts, fmt.Sprintf("%d blocker", b.blockers))
		}
		if b.questions > 0 {
			parts = append(parts, fmt.Sprintf("%d question", b.questions))
		}
		if b.other > 0 {
			parts = append(parts, fmt.Sprintf("%d other", b.other))
		}
		detail := ""
		if len(parts) > 0 {
			detail = "  (" + strings.Join(parts, ", ") + ")"
		}
		sb.WriteString(fmt.Sprintf("  %-30s  %3d captures%s\n", b.label, total, detail))
	}
	return sb.String()
}

// isoWeekStart returns the Monday of the given ISO week.
func isoWeekStart(year, week int) time.Time {
	// Jan 4 is always in ISO week 1
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.Local)
	// Find Monday of week 1
	weekday := jan4.Weekday()
	if weekday == 0 {
		weekday = 7
	}
	week1Monday := jan4.AddDate(0, 0, -int(weekday-1))
	return week1Monday.AddDate(0, 0, (week-1)*7)
}

func statsProjects(captures []database.Capture) string {
	type projStats struct {
		count     int
		firstSeen time.Time
		lastSeen  time.Time
		decisions int
		blockers  int
	}
	stats := make(map[string]*projStats)

	for _, c := range captures {
		proj := c.Project
		if proj == "" {
			proj = "(none)"
		}
		s, ok := stats[proj]
		if !ok {
			stats[proj] = &projStats{count: 1, firstSeen: c.Timestamp, lastSeen: c.Timestamp}
			s = stats[proj]
		} else {
			s.count++
			if c.Timestamp.Before(s.firstSeen) {
				s.firstSeen = c.Timestamp
			}
			if c.Timestamp.After(s.lastSeen) {
				s.lastSeen = c.Timestamp
			}
		}
		for _, tag := range strings.Split(c.Tags, ",") {
			switch strings.TrimSpace(tag) {
			case "decision":
				s.decisions++
			case "blocker":
				s.blockers++
			}
		}
	}

	// Sort by count descending
	type projEntry struct {
		name  string
		stats *projStats
	}
	var entries []projEntry
	for name, s := range stats {
		entries = append(entries, projEntry{name, s})
	}
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].stats.count > entries[i].stats.count {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Projects (%d captures):\n\n", len(captures)))
	for _, e := range entries {
		var detail string
		var parts []string
		if e.stats.decisions > 0 {
			parts = append(parts, fmt.Sprintf("%d decisions", e.stats.decisions))
		}
		if e.stats.blockers > 0 {
			parts = append(parts, fmt.Sprintf("%d blockers", e.stats.blockers))
		}
		if len(parts) > 0 {
			detail = "  (" + strings.Join(parts, ", ") + ")"
		}
		sb.WriteString(fmt.Sprintf("  %-20s %4d captures  first: %s  last: %s%s\n",
			e.name, e.stats.count,
			e.stats.firstSeen.Format("2006-01-02"),
			e.stats.lastSeen.Format("2006-01-02"),
			detail,
		))
	}
	sb.WriteString(fmt.Sprintf("\nTotal: %d projects\n", len(entries)))
	return sb.String()
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

func mcpConventions(repos, scanDir string, days int, since string, correlate *bool, auditDir string, allDecisions bool, allCommits bool, maxPerRepo int) (string, error) {
	if repos == "" && scanDir == "" {
		return "", fmt.Errorf("provide repos (comma-separated paths) or scan_dir")
	}

	// Split and resolve explicit repo paths
	var repoPaths []string
	for _, r := range strings.Split(repos, ",") {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if strings.HasPrefix(r, "~/") {
			home, _ := os.UserHomeDir()
			r = filepath.Join(home, r[2:])
		}
		abs, err := filepath.Abs(r)
		if err != nil {
			continue
		}
		repoPaths = append(repoPaths, abs)
	}

	// Resolve scan_dir
	if scanDir != "" {
		if strings.HasPrefix(scanDir, "~/") {
			home, _ := os.UserHomeDir()
			scanDir = filepath.Join(home, scanDir[2:])
		}
		abs, err := filepath.Abs(scanDir)
		if err != nil {
			return "", fmt.Errorf("resolve scan_dir: %w", err)
		}
		scanDir = abs
	}

	cor := true
	if correlate != nil {
		cor = *correlate
	}
	if days == 0 {
		days = 180
	}

	var sinceTime *time.Time
	if since != "" {
		t, err := mcpParseTimestamp(since)
		if err != nil {
			return "", fmt.Errorf("invalid since: %w", err)
		}
		sinceTime = &t
	}
	if sinceTime == nil {
		t := time.Now().AddDate(0, 0, -days)
		sinceTime = &t
	}

	opts := conventions.Options{
		Repos:        repoPaths,
		ScanDir:      scanDir,
		Days:         days,
		Since:        sinceTime,
		Correlate:    cor,
		AuditDir:     auditDir,
		AllDecisions: allDecisions,
		AllCommits:   allCommits,
		MaxPerRepo:   maxPerRepo,
	}

	result, err := conventions.Run(opts, getDBPath())
	if err != nil {
		return "", err
	}

	return result.Prompt, nil
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

func mcpEnforcement(action, hook string, threshold int) (string, error) {
	cfg := loadEnforcementConfig()

	switch action {
	case "status":
		preStatus := "disabled"
		if cfg.PreCompact.Enabled {
			preStatus = "enabled"
		}
		postStatus := "disabled"
		if cfg.PostToolNudge.Enabled {
			postStatus = fmt.Sprintf("enabled (threshold: %d)", cfg.PostToolNudge.Threshold)
		}
		return fmt.Sprintf("Enforcement hooks:\n  pre_compact:     %s\n  post_tool_nudge: %s\n\nConfig: %s",
			preStatus, postStatus, getEnforcementConfigPath()), nil

	case "enable":
		if hook == "" {
			return "", fmt.Errorf("hook name required (pre_compact or post_tool_nudge)")
		}
		switch hook {
		case "pre_compact":
			cfg.PreCompact.Enabled = true
		case "post_tool_nudge":
			cfg.PostToolNudge.Enabled = true
			if threshold > 0 {
				cfg.PostToolNudge.Threshold = threshold
			}
		default:
			return "", fmt.Errorf("unknown hook: %s (use pre_compact or post_tool_nudge)", hook)
		}
		if err := saveEnforcementConfig(cfg); err != nil {
			return "", fmt.Errorf("save config: %w", err)
		}
		return fmt.Sprintf("Enabled %s", hook), nil

	case "disable":
		if hook == "" {
			return "", fmt.Errorf("hook name required (pre_compact or post_tool_nudge)")
		}
		switch hook {
		case "pre_compact":
			cfg.PreCompact.Enabled = false
		case "post_tool_nudge":
			cfg.PostToolNudge.Enabled = false
		default:
			return "", fmt.Errorf("unknown hook: %s (use pre_compact or post_tool_nudge)", hook)
		}
		if err := saveEnforcementConfig(cfg); err != nil {
			return "", fmt.Errorf("save config: %w", err)
		}
		return fmt.Sprintf("Disabled %s", hook), nil

	case "configure":
		if hook == "" {
			return "", fmt.Errorf("hook name required")
		}
		switch hook {
		case "post_tool_nudge":
			if threshold > 0 {
				cfg.PostToolNudge.Threshold = threshold
			} else {
				return "", fmt.Errorf("threshold required for post_tool_nudge (e.g., 10, 15, 20)")
			}
		default:
			return "", fmt.Errorf("no configurable options for %s", hook)
		}
		if err := saveEnforcementConfig(cfg); err != nil {
			return "", fmt.Errorf("save config: %w", err)
		}
		return fmt.Sprintf("Updated %s: threshold=%d", hook, cfg.PostToolNudge.Threshold), nil

	default:
		return "", fmt.Errorf("unknown action: %s (use status, enable, disable, configure)", action)
	}
}

func getCommitsBetween(since, until string, limit int) []string {
	args := []string{"log", "--format=%s", "--no-merges"}
	if since != "" {
		args = append(args, "--since="+since)
	}
	if until != "" {
		args = append(args, "--until="+until)
	}
	args = append(args, "-n", fmt.Sprintf("%d", limit))

	out, err := exec.Command("git", args...).Output()
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
