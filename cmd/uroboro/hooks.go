package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/QRY91/uroboro/internal/hooks"
)

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

	startPath := filepath.Join(hooksDir, "session-start.sh")
	auditPath := filepath.Join(hooksDir, "session-audit.sh")

	if err := os.WriteFile(startPath, hooks.SessionStartScript, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write session-start.sh: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(auditPath, hooks.SessionAuditScript, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write session-audit.sh: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s\n", startPath)
	fmt.Printf("wrote %s\n", auditPath)

	// 2. Patch settings.json
	settings := loadSettings(settingsPath)

	hooks := getOrCreateMap(settings, "hooks")
	patchHookEvent(hooks, "SessionStart", startPath)
	patchHookEvent(hooks, "Stop", auditPath)

	writeSettings(settingsPath, settings)
	fmt.Printf("patched %s\n", settingsPath)
	fmt.Println("\nuroboro hooks installed.")
}

func uninstallHooks() {
	hooksDir := getHooksDir()
	settingsPath := getSettingsPath()

	startPath := filepath.Join(hooksDir, "session-start.sh")
	auditPath := filepath.Join(hooksDir, "session-audit.sh")

	// 1. Remove from settings.json
	if _, err := os.Stat(settingsPath); err == nil {
		settings := loadSettings(settingsPath)
		if hooks, ok := settings["hooks"].(map[string]interface{}); ok {
			removeHookCommand(hooks, "SessionStart", startPath)
			removeHookCommand(hooks, "Stop", auditPath)
			writeSettings(settingsPath, settings)
			fmt.Printf("cleaned %s\n", settingsPath)
		}
	}

	// 2. Remove hook scripts
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

	startPath := filepath.Join(hooksDir, "session-start.sh")
	auditPath := filepath.Join(hooksDir, "session-audit.sh")

	startExists := fileExists(startPath)
	auditExists := fileExists(auditPath)

	startRegistered := false
	auditRegistered := false
	if _, err := os.Stat(settingsPath); err == nil {
		settings := loadSettings(settingsPath)
		if hooks, ok := settings["hooks"].(map[string]interface{}); ok {
			startRegistered = hookCommandExists(hooks, "SessionStart", startPath)
			auditRegistered = hookCommandExists(hooks, "Stop", auditPath)
		}
	}

	fmt.Println("uroboro hooks status:")
	printStatus("  session-start.sh", startExists, startRegistered)
	printStatus("  session-audit.sh", auditExists, auditRegistered)

	if startExists && auditExists && startRegistered && auditRegistered {
		fmt.Println("\nAll hooks installed and registered.")
	} else if !startExists && !auditExists && !startRegistered && !auditRegistered {
		fmt.Println("\nNot installed. Run: uro hooks install")
	} else {
		fmt.Println("\nPartially installed. Run: uro hooks install")
	}
}

func printStatus(label string, exists, registered bool) {
	fileStatus := "missing"
	if exists {
		fileStatus = "installed"
	}
	regStatus := "not registered"
	if registered {
		regStatus = "registered"
	}
	fmt.Printf("%s: %s, %s\n", label, fileStatus, regStatus)
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
// settings.json structure: hooks.<Event>[{matcher:"", hooks:[{type,command,timeout},...]}]
func patchHookEvent(hooks map[string]interface{}, event, command string) {
	// Get or create the event's matcher array
	var matchers []interface{}
	if v, ok := hooks[event]; ok {
		if arr, ok := v.([]interface{}); ok {
			matchers = arr
		}
	}

	// Find or create the catch-all matcher entry (matcher: "")
	var matcherEntry map[string]interface{}
	for _, m := range matchers {
		if me, ok := m.(map[string]interface{}); ok {
			if matcher, ok := me["matcher"].(string); ok && matcher == "" {
				matcherEntry = me
				break
			}
		}
	}
	if matcherEntry == nil {
		matcherEntry = map[string]interface{}{
			"matcher": "",
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
	hooks[event] = matchers
}

// removeHookCommand removes a specific command from an event's hook arrays.
func removeHookCommand(hooks map[string]interface{}, event, command string) {
	v, ok := hooks[event]
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
func hookCommandExists(hooks map[string]interface{}, event, command string) bool {
	v, ok := hooks[event]
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
