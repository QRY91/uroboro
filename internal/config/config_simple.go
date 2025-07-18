package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/QRY91/uroboro/internal/common"
)

// Config holds simple configuration for uroboro
type Config struct {
	DefaultDBPath string
	UseDatabase   bool
	AutoCapture   bool // Future: for daemon mode
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return filepath.Join(common.GetConfigDir(), "config")
}

// LoadConfig loads configuration from a simple key=value file
func LoadConfig() (*Config, error) {
	config := &Config{
		UseDatabase: true,  // Default to database if available
		AutoCapture: false, // Default to manual capture
	}

	configPath := GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil // Return defaults if no config file
	}

	file, err := os.Open(configPath)
	if err != nil {
		return config, nil // Return defaults on error
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "default_db_path":
			config.DefaultDBPath = value
		case "use_database":
			config.UseDatabase = (value == "true")
		case "auto_capture":
			config.AutoCapture = (value == "true")
		}
	}

	return config, nil
}

// SaveConfig saves configuration to a simple key=value file
func SaveConfig(config *Config) error {
	configDir := common.GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := GetConfigPath()
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Write simple config format
	fmt.Fprintf(file, "# uroboro configuration\n")
	fmt.Fprintf(file, "# Simple key=value format\n\n")

	if config.DefaultDBPath != "" {
		fmt.Fprintf(file, "default_db_path=%s\n", config.DefaultDBPath)
	}
	fmt.Fprintf(file, "use_database=%t\n", config.UseDatabase)
	fmt.Fprintf(file, "auto_capture=%t\n", config.AutoCapture)

	return nil
}

// GetDefaultDBPath returns the default database path
func GetDefaultDBPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".uroboro", "uroboro.db")
}

// GetOrCreateConfig loads config and applies defaults
func GetOrCreateConfig() *Config {
	config, err := LoadConfig()
	if err != nil {
		// Return safe defaults on any error
		return &Config{
			DefaultDBPath: GetDefaultDBPath(),
			UseDatabase:   true,
			AutoCapture:   false,
		}
	}

	// Apply default database path if not set
	if config.DefaultDBPath == "" {
		config.DefaultDBPath = GetDefaultDBPath()
	}

	return config
}
