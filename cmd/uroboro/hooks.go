package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/QRY91/uroboro/internal/hooks"
)

// --- Enforcement config types ---

type EnforcementConfig struct {
	PreCompact    HookConfig  `json:"pre_compact"`
	PostToolNudge NudgeConfig `json:"post_tool_nudge"`
}

type HookConfig struct {
	Enabled bool `json:"enabled"`
}

type NudgeConfig struct {
	Enabled   bool `json:"enabled"`
	Threshold int  `json:"threshold"`
}

func defaultEnforcementConfig() EnforcementConfig {
	return EnforcementConfig{
		PreCompact:    HookConfig{Enabled: false},
		PostToolNudge: NudgeConfig{Enabled: false, Threshold: 15},
	}
}

func getEnforcementConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "uroboro", "enforcement.json")
}

func loadEnforcementConfig() EnforcementConfig {
	path := getEnforcementConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultEnforcementConfig()
	}
	var cfg EnforcementConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultEnforcementConfig()
	}
	if cfg.PostToolNudge.Threshold == 0 {
		cfg.PostToolNudge.Threshold = 15
	}
	return cfg
}

func saveEnforcementConfig(cfg EnforcementConfig) error {
	path := getEnforcementConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if cfg.PostToolNudge.Threshold == 0 {
		cfg.PostToolNudge.Threshold = 15
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

// --- Hook commands ---

func handleHooks(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, `Usage: uroboro hooks <command>

Commands:
  install      Install enforcement hooks into Claude Code
  uninstall    Remove enforcement hooks from Claude Code
  status       Check if hooks are installed`)
		os.Exit(1)
	}

	switch args[0] {
	case "install":
		installHooks()
	case "uninstall":
		uninstallHooks()
	case "status":
		hooksStatus()
	default:
		fmt.Fprintf(os.Stderr, "Unknown hooks command: %s\n", args[0])
		os.Exit(1)
	}
}

func getClaudeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine home directory: %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(home, ".claude")
}

func getHooksDir() string {
	return filepath.Join(getClaudeDir(), "hooks", "uroboro")
}

func getSettingsPath() string {
	return filepath.Join(getClaudeDir(), "settings.json")
}

func installHooks() {
	hooksDir := getHooksDir()
	settingsPath := getSettingsPath()

	// 1. Write hook scripts
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create hooks directory: %v\n", err)
		os.Exit(1)
	}

	scriptMap := map[string][]byte{
		"session-start.sh":   hooks.SessionStartScript,
		"session-audit.sh":   hooks.SessionAuditScript,
		"pre-compact.sh":     hooks.PreCompactScript,
		"post-tool-nudge.sh": hooks.PostToolNudgeScript,
	}

	for name, script := range scriptMap {
		path := filepath.Join(hooksDir, name)
		if err := os.WriteFile(path, script, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s\n", path)
	}

	// 2. Patch settings.json
	settings := loadSettings(settingsPath)
	hooksMap := getOrCreateMap(settings, "hooks")

	startPath := filepath.Join(hooksDir, "session-start.sh")
	auditPath := filepath.Join(hooksDir, "session-audit.sh")
	preCompactPath := filepath.Join(hooksDir, "pre-compact.sh")
	postToolPath := filepath.Join(hooksDir, "post-tool-nudge.sh")

	// Core hooks: catch-all matchers
	patchHookEvent(hooksMap, "SessionStart", startPath, "")
	patchHookEvent(hooksMap, "Stop", auditPath, "")

	// Enforcement hooks: specific matchers
	patchHookEvent(hooksMap, "PreCompact", preCompactPath, "")
	patchHookEvent(hooksMap, "PostToolUse", postToolPath, "Edit|Write|Bash")

	writeSettings(settingsPath, settings)
	fmt.Printf("patched %s\n", settingsPath)

	// 3. Write default enforcement config (if not exists)
	cfgPath := getEnforcementConfigPath()
	if !fileExists(cfgPath) {
		cfg := defaultEnforcementConfig()
		if err := saveEnforcementConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not write enforcement config: %v\n", err)
		} else {
			fmt.Printf("wrote %s\n", cfgPath)
		}
	}

	fmt.Println("\nuroboro hooks installed.")
	fmt.Println("  pre-compact and post-tool-nudge are opt-in (disabled by default).")
	fmt.Println("  Enable via MCP: uro_enforcement(action: \"enable\", hook: \"pre_compact\")")
}

func uninstallHooks() {
	hooksDir := getHooksDir()
	settingsPath := getSettingsPath()

	allScripts := []string{
		"session-start.sh",
		"session-audit.sh",
		"pre-compact.sh",
		"post-tool-nudge.sh",
	}

	// Event → script mapping for removal
	eventMap := map[string]string{
		"SessionStart": "session-start.sh",
		"Stop":         "session-audit.sh",
		"PreCompact":   "pre-compact.sh",
		"PostToolUse":  "post-tool-nudge.sh",
	}

	// 1. Remove from settings.json
	if _, err := os.Stat(settingsPath); err == nil {
		settings := loadSettings(settingsPath)
		if hooksMap, ok := settings["hooks"].(map[string]interface{}); ok {
			for event, script := range eventMap {
				removeHookCommand(hooksMap, event, filepath.Join(hooksDir, script))
			}
			writeSettings(settingsPath, settings)
			fmt.Printf("cleaned %s\n", settingsPath)
		}
	}

	// 2. Remove hook scripts
	_ = allScripts // used for documentation; we remove the whole dir
	if err := os.RemoveAll(hooksDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to remove hooks directory: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("removed %s\n", hooksDir)
	fmt.Println("\nuroboro hooks uninstalled.")
}

func hooksStatus() {
	hooksDir := getHooksDir()
	settingsPath := getSettingsPath()

	type hookInfo struct {
		name     string
		script   string
		event    string
		optIn    bool
		enabled  bool
	}

	cfg := loadEnforcementConfig()

	allHooks := []hookInfo{
		{"session-start", "session-start.sh", "SessionStart", false, true},
		{"session-audit", "session-audit.sh", "Stop", false, true},
		{"pre-compact", "pre-compact.sh", "PreCompact", true, cfg.PreCompact.Enabled},
		{"post-tool-nudge", "post-tool-nudge.sh", "PostToolUse", true, cfg.PostToolNudge.Enabled},
	}

	var settings map[string]interface{}
	if _, err := os.Stat(settingsPath); err == nil {
		settings = loadSettings(settingsPath)
	}

	fmt.Println("uroboro hooks status:")
	allInstalled := true
	for _, h := range allHooks {
		scriptPath := filepath.Join(hooksDir, h.script)
		exists := fileExists(scriptPath)

		registered := false
		if settings != nil {
			if hooksMap, ok := settings["hooks"].(map[string]interface{}); ok {
				registered = hookCommandExists(hooksMap, h.event, scriptPath)
			}
		}

		fileStatus := "missing"
		if exists {
			fileStatus = "installed"
		}
		regStatus := "not registered"
		if registered {
			regStatus = "registered"
		}

		extra := ""
		if h.optIn {
			if h.enabled {
				extra = ", enabled"
			} else {
				extra = ", disabled (opt-in)"
			}
		}

		fmt.Printf("  %-20s %s, %s%s\n", h.name+":", fileStatus, regStatus, extra)

		if !exists || !registered {
			allInstalled = false
		}
	}

	if cfg.PostToolNudge.Enabled {
		fmt.Printf("\n  post-tool-nudge threshold: every %d tool calls\n", cfg.PostToolNudge.Threshold)
	}

	if allInstalled {
		fmt.Println("\nAll hooks installed and registered.")
	} else {
		anyInstalled := false
		for _, h := range allHooks {
			if fileExists(filepath.Join(hooksDir, h.script)) {
				anyInstalled = true
				break
			}
		}
		if anyInstalled {
			fmt.Println("\nPartially installed. Run: uroboro hooks install")
		} else {
			fmt.Println("\nNot installed. Run: uroboro hooks install")
		}
	}
}

// --- settings.json helpers ---

func loadSettings(path string) map[string]interface{} {
	data, err := os.ReadFile(path)
	if err != nil {
		// No settings file yet — start fresh
		return map[string]interface{}{}
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not parse %s, creating new: %v\n", path, err)
		return map[string]interface{}{}
	}
	return settings
}

func writeSettings(path string, settings map[string]interface{}) {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal settings: %v\n", err)
		os.Exit(1)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write settings: %v\n", err)
		os.Exit(1)
	}
}

func getOrCreateMap(parent map[string]interface{}, key string) map[string]interface{} {
	if v, ok := parent[key]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	m := map[string]interface{}{}
	parent[key] = m
	return m
}

// patchHookEvent adds a uroboro hook command to an event's hook array if not already present.
// settings.json structure: hooks.<Event>[{matcher:"...", hooks:[{type,command,timeout},...]}]
// Use matcher="" for catch-all, or a specific pattern like "Edit|Write|Bash".
func patchHookEvent(hooksMap map[string]interface{}, event, command, matcher string) {
	// Get or create the event's matcher array
	var matchers []interface{}
	if v, ok := hooksMap[event]; ok {
		if arr, ok := v.([]interface{}); ok {
			matchers = arr
		}
	}

	// Find or create the matching matcher entry
	var matcherEntry map[string]interface{}
	for _, m := range matchers {
		if me, ok := m.(map[string]interface{}); ok {
			if mv, ok := me["matcher"].(string); ok && mv == matcher {
				matcherEntry = me
				break
			}
		}
	}
	if matcherEntry == nil {
		matcherEntry = map[string]interface{}{
			"matcher": matcher,
			"hooks":   []interface{}{},
		}
		matchers = append(matchers, matcherEntry)
	}

	// Get the hooks array within the matcher
	var hookList []interface{}
	if v, ok := matcherEntry["hooks"]; ok {
		if arr, ok := v.([]interface{}); ok {
			hookList = arr
		}
	}

	// Check if uroboro hook already registered
	for _, h := range hookList {
		if hm, ok := h.(map[string]interface{}); ok {
			if cmd, ok := hm["command"].(string); ok && cmd == command {
				return // Already present
			}
		}
	}

	// Append uroboro hook
	hookList = append(hookList, map[string]interface{}{
		"type":    "command",
		"command": command,
		"timeout": 10,
	})
	matcherEntry["hooks"] = hookList
	hooksMap[event] = matchers
}

// removeHookCommand removes a specific command from an event's hook arrays.
func removeHookCommand(hooksMap map[string]interface{}, event, command string) {
	v, ok := hooksMap[event]
	if !ok {
		return
	}
	matchers, ok := v.([]interface{})
	if !ok {
		return
	}
	for _, m := range matchers {
		me, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		hookArr, ok := me["hooks"].([]interface{})
		if !ok {
			continue
		}
		filtered := make([]interface{}, 0, len(hookArr))
		for _, h := range hookArr {
			hm, ok := h.(map[string]interface{})
			if !ok {
				filtered = append(filtered, h)
				continue
			}
			if cmd, ok := hm["command"].(string); ok && cmd == command {
				continue // Skip this one
			}
			filtered = append(filtered, h)
		}
		me["hooks"] = filtered
	}
}

// hookCommandExists checks if a specific command is registered for an event.
func hookCommandExists(hooksMap map[string]interface{}, event, command string) bool {
	v, ok := hooksMap[event]
	if !ok {
		return false
	}
	matchers, ok := v.([]interface{})
	if !ok {
		return false
	}
	for _, m := range matchers {
		me, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		hookArr, ok := me["hooks"].([]interface{})
		if !ok {
			continue
		}
		for _, h := range hookArr {
			hm, ok := h.(map[string]interface{})
			if !ok {
				continue
			}
			if cmd, ok := hm["command"].(string); ok && strings.Contains(cmd, command) {
				return true
			}
		}
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
