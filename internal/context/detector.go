package context

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProjectDetector handles smart project detection logic
type ProjectDetector struct{}

// NewProjectDetector creates a new project detector
func NewProjectDetector() *ProjectDetector {
	return &ProjectDetector{}
}

// DetectProject attempts to automatically detect the current project name
func (pd *ProjectDetector) DetectProject() string {
	// Try different detection methods in order of preference
	if project := pd.detectFromGit(); project != "" {
		return project
	}

	if project := pd.detectFromProjectFiles(); project != "" {
		return project
	}

	if project := pd.detectFromDirectory(); project != "" {
		return project
	}

	return ""
}

// detectFromGit tries to get project name from git repository
func (pd *ProjectDetector) detectFromGit() string {
	// Get git remote origin URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	url := strings.TrimSpace(string(output))
	if url == "" {
		return ""
	}

	// Extract project name from various URL formats
	// SSH: git@github.com:user/repo.git
	// HTTPS: https://github.com/user/repo.git
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Get the last part after / or :
	parts := strings.FieldsFunc(url, func(c rune) bool {
		return c == '/' || c == ':'
	})

	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}

// detectFromProjectFiles looks for common project files and extracts names
func (pd *ProjectDetector) detectFromProjectFiles() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Check for package.json (Node.js projects)
	if project := pd.extractFromPackageJSON(cwd); project != "" {
		return project
	}

	// Check for go.mod (Go projects)
	if project := pd.extractFromGoMod(cwd); project != "" {
		return project
	}

	// Check for Cargo.toml (Rust projects)
	if project := pd.extractFromCargoToml(cwd); project != "" {
		return project
	}

	return ""
}

// detectFromDirectory uses directory name as fallback
func (pd *ProjectDetector) detectFromDirectory() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	dirName := filepath.Base(cwd)

	// Clean up common prefixes/suffixes
	dirName = strings.TrimPrefix(dirName, "project-")
	dirName = strings.TrimPrefix(dirName, "app-")
	dirName = strings.TrimSuffix(dirName, "-app")
	dirName = strings.TrimSuffix(dirName, "-project")

	// Only return if it looks like a meaningful project name
	if len(dirName) > 2 && dirName != "src" && dirName != "app" {
		return dirName
	}

	return ""
}

// extractFromPackageJSON reads project name from package.json
func (pd *ProjectDetector) extractFromPackageJSON(dir string) string {
	packagePath := filepath.Join(dir, "package.json")
	content, err := os.ReadFile(packagePath)
	if err != nil {
		return ""
	}

	// Simple extraction - look for "name": "project-name"
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, `"name":`) {
			// Extract name value
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				name := strings.TrimSpace(parts[1])
				name = strings.TrimSuffix(name, ",")
				name = strings.Trim(name, `"`)
				return name
			}
		}
	}

	return ""
}

// extractFromGoMod reads module name from go.mod
func (pd *ProjectDetector) extractFromGoMod(dir string) string {
	goModPath := filepath.Join(dir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			moduleName := strings.TrimPrefix(line, "module ")
			moduleName = strings.TrimSpace(moduleName)

			// Extract just the project name from full module path
			parts := strings.Split(moduleName, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
	}

	return ""
}

// extractFromCargoToml reads project name from Cargo.toml
func (pd *ProjectDetector) extractFromCargoToml(dir string) string {
	cargoPath := filepath.Join(dir, "Cargo.toml")
	content, err := os.ReadFile(cargoPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	inPackageSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "[package]" {
			inPackageSection = true
			continue
		}

		if strings.HasPrefix(line, "[") && line != "[package]" {
			inPackageSection = false
			continue
		}

		if inPackageSection && strings.HasPrefix(line, "name = ") {
			name := strings.TrimPrefix(line, "name = ")
			name = strings.Trim(name, `"`)
			return name
		}
	}

	return ""
}

// IsGitRepository checks if current directory is a git repository
func (pd *ProjectDetector) IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// GetWorkingDirectory returns the current working directory name
func (pd *ProjectDetector) GetWorkingDirectory() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(cwd)
}
