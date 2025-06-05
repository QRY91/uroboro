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
	DefaultDBPath string `json:"default_db_path"`
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

	// Simple implementation - just save the default DB path for now
	content := fmt.Sprintf("default_db_path=%s\n", config.DefaultDBPath)

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// LoadDefaultDBPath loads the default database path from config
func LoadDefaultDBPath() (string, error) {
	configPath := GetConfigPath()

	// If config file doesn't exist, return empty string
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", nil
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	// Simple parsing - look for default_db_path=
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "default_db_path=") {
			return strings.TrimPrefix(line, "default_db_path="), nil
		}
	}

	return "", nil
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
