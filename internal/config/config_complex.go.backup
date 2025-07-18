package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/QRY91/uroboro/internal/common"
)

type Config struct {
	DefaultDBPath    string `json:"default_db_path"`
	AnalyticsEnabled bool   `json:"analytics_enabled"`
	PostHogAPIKey    string `json:"posthog_api_key"`
	PostHogHost      string `json:"posthog_host"`
	PrivacyMode      string `json:"privacy_mode"`
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return filepath.Join(common.GetConfigDir(), "config.json")
}

// LoadConfig loads the configuration from file, creating defaults if needed
func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()

	// If config file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	// For now, return empty config - we'll implement JSON loading later if needed
	// This keeps it simple and focused on the UX improvement
	return &Config{}, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config *Config) error {
	configDir := common.GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := GetConfigPath()

	// Simple implementation - save all config values
	content := fmt.Sprintf("default_db_path=%s\n", config.DefaultDBPath)
	content += fmt.Sprintf("analytics_enabled=%t\n", config.AnalyticsEnabled)
	content += fmt.Sprintf("posthog_api_key=%s\n", config.PostHogAPIKey)
	content += fmt.Sprintf("posthog_host=%s\n", config.PostHogHost)
	content += fmt.Sprintf("privacy_mode=%s\n", config.PrivacyMode)

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// parseConfigFile parses the simple key=value config file format
func parseConfigFile() (*Config, error) {
	configPath := GetConfigPath()

	// If config file doesn't exist, return defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			AnalyticsEnabled: false,
			PostHogHost:      "https://eu.posthog.com",
			PrivacyMode:      "enhanced",
		}, nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := &Config{
		AnalyticsEnabled: false,
		PostHogHost:      "https://eu.posthog.com",
		PrivacyMode:      "enhanced",
	}

	// Simple parsing - look for key=value pairs
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "default_db_path=") {
			config.DefaultDBPath = strings.TrimPrefix(line, "default_db_path=")
		} else if strings.HasPrefix(line, "analytics_enabled=") {
			value := strings.TrimPrefix(line, "analytics_enabled=")
			config.AnalyticsEnabled = value == "true"
		} else if strings.HasPrefix(line, "posthog_api_key=") {
			config.PostHogAPIKey = strings.TrimPrefix(line, "posthog_api_key=")
		} else if strings.HasPrefix(line, "posthog_host=") {
			config.PostHogHost = strings.TrimPrefix(line, "posthog_host=")
		} else if strings.HasPrefix(line, "privacy_mode=") {
			config.PrivacyMode = strings.TrimPrefix(line, "privacy_mode=")
		}
	}

	return config, nil
}

// LoadDefaultDBPath loads the default database path from config
func LoadDefaultDBPath() (string, error) {
	config, err := parseConfigFile()
	if err != nil {
		return "", err
	}
	return config.DefaultDBPath, nil
}

// SaveDefaultDBPath saves the default database path to config
func SaveDefaultDBPath(dbPath string) error {
	config := &Config{DefaultDBPath: dbPath}
	return SaveConfig(config)
}

// PromptForDefaultDB interactively prompts user to set default database path
func PromptForDefaultDB() (string, error) {
	defaultPath := common.GetDefaultDBPath()

	fmt.Printf("🗄️  No default database configured.\n")
	fmt.Printf("   Suggested: %s\n", defaultPath)
	fmt.Printf("   Create default database? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)

	// Default to yes if empty or starts with y/Y
	if input == "" || strings.ToLower(input)[0] == 'y' {
		// Create directory if needed
		dbDir := filepath.Dir(defaultPath)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create database directory: %w", err)
		}

		// Save config
		config := &Config{DefaultDBPath: defaultPath}
		if err := SaveConfig(config); err != nil {
			return "", fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✅ Default database set to: %s\n", defaultPath)
		return defaultPath, nil
	}

	fmt.Println("❌ Database setup cancelled. Use --db=path/to/db.sqlite to specify manually.")
	return "", fmt.Errorf("database setup cancelled")
}

// GetDefaultDBPath returns the configured default database path or prompts to set one
func GetDefaultDBPath() (string, error) {
	// Try to load from config first
	dbPath, err := LoadDefaultDBPath()
	if err != nil {
		return "", err
	}

	// If we have a configured path, return it
	if dbPath != "" {
		return dbPath, nil
	}

	// Otherwise, prompt user to set one
	return PromptForDefaultDB()
}

// LoadAnalyticsConfig loads analytics configuration from config file
func LoadAnalyticsConfig() (*Config, error) {
	return parseConfigFile()
}

// SaveAnalyticsConfig saves analytics configuration
func SaveAnalyticsConfig(enabled bool, apiKey, host, privacyMode string) error {
	// Load existing config to preserve other settings
	config, err := parseConfigFile()
	if err != nil {
		config = &Config{}
	}

	// Update analytics settings
	config.AnalyticsEnabled = enabled
	config.PostHogAPIKey = apiKey
	config.PostHogHost = host
	config.PrivacyMode = privacyMode

	return SaveConfig(config)
}

// IsAnalyticsEnabled checks if analytics is enabled in config
func IsAnalyticsEnabled() bool {
	config, err := parseConfigFile()
	if err != nil {
		return false
	}
	return config.AnalyticsEnabled
}
