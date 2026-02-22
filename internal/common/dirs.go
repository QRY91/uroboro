package common

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetDataDir returns the appropriate data directory for the current OS
// Linux/macOS: ~/.local/share/uroboro/daily (XDG compliant)
// Windows: %APPDATA%\uroboro\daily
func GetDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if we can't get home
		return filepath.Join(".", "uroboro", "daily")
	}

	if runtime.GOOS == "windows" {
		// Windows: Use AppData/Roaming
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "uroboro", "daily")
		}
		// Fallback if APPDATA not set
		return filepath.Join(homeDir, "AppData", "Roaming", "uroboro", "daily")
	}

	// Linux/macOS: XDG compliant
	return filepath.Join(homeDir, ".local", "share", "uroboro", "daily")
}

// GetConfigDir returns the cross-platform config directory
// Linux/macOS: ~/.config/uroboro (XDG compliant)
// Windows: %APPDATA%/uroboro
func GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if we can't get home
		return filepath.Join(".", "uroboro")
	}

	if runtime.GOOS == "windows" {
		// Windows: Use AppData/Roaming
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "uroboro")
		}
		// Fallback if APPDATA not set
		return filepath.Join(homeDir, "AppData", "Roaming", "uroboro")
	}

	// Linux/macOS: XDG compliant
	return filepath.Join(homeDir, ".config", "uroboro")
}

// GetBackupDir returns the default backup directory
// Linux/macOS: ~/.local/share/uroboro/backups
// Windows: %APPDATA%/uroboro/backups
func GetBackupDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "uroboro", "backups")
	}

	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "uroboro", "backups")
		}
		return filepath.Join(homeDir, "AppData", "Roaming", "uroboro", "backups")
	}

	return filepath.Join(homeDir, ".local", "share", "uroboro", "backups")
}

// GetDefaultDBPath returns the default database path
// Linux/macOS: ~/.local/share/uroboro/uroboro.sqlite
// Windows: %APPDATA%/uroboro/uroboro.sqlite
func GetDefaultDBPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if we can't get home
		return filepath.Join(".", "uroboro.sqlite")
	}

	if runtime.GOOS == "windows" {
		// Windows: Use AppData/Roaming
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "uroboro", "uroboro.sqlite")
		}
		// Fallback if APPDATA not set
		return filepath.Join(homeDir, "AppData", "Roaming", "uroboro", "uroboro.sqlite")
	}

	// Linux/macOS: XDG compliant
	return filepath.Join(homeDir, ".local", "share", "uroboro", "uroboro.sqlite")
}
