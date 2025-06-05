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
