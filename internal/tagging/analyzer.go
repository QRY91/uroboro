package tagging

import (
	"regexp"
	"strings"
)

// TagAnalyzer handles smart content-based tag generation
type TagAnalyzer struct {
	actionPatterns     map[string][]string
	technologyPatterns map[string][]string
	domainPatterns     map[string][]string
}

// NewTagAnalyzer creates a new tag analyzer with predefined patterns
func NewTagAnalyzer() *TagAnalyzer {
	return &TagAnalyzer{
		actionPatterns: map[string][]string{
			"bugfix":        {"fixed", "fix", "bug", "error", "issue", "problem", "crash", "broken"},
			"feature":       {"implemented", "implement", "added", "add", "created", "create", "built", "build"},
			"refactoring":   {"refactored", "refactor", "cleaned", "clean", "reorganized", "restructured", "improved"},
			"optimization":  {"optimized", "optimize", "performance", "faster", "speed", "efficient", "reduced"},
			"testing":       {"test", "tested", "unit test", "integration test", "spec", "coverage"},
			"debugging":     {"debug", "debugged", "investigated", "traced", "logging", "console"},
			"documentation": {"documented", "comment", "readme", "docs", "documentation"},
			"setup":         {"setup", "configured", "installed", "initialized", "scaffold"},
		},
		technologyPatterns: map[string][]string{
			"react":      {"react", "jsx", "component", "hook", "usestate", "useeffect"},
			"go":         {"golang", "go", "goroutine", "channel", "interface", "struct"},
			"javascript": {"javascript", "js", "node", "npm", "yarn", "webpack"},
			"python":     {"python", "django", "flask", "pip", "conda", "virtualenv"},
			"rust":       {"rust", "cargo", "crate", "trait", "impl", "ownership"},
			"database":   {"sql", "database", "db", "mysql", "postgres", "sqlite", "mongodb"},
			"docker":     {"docker", "container", "dockerfile", "compose", "image"},
			"git":        {"git", "commit", "merge", "branch", "pull request", "pr"},
			"api":        {"api", "rest", "graphql", "endpoint", "route", "handler"},
			"websocket":  {"websocket", "ws", "socket", "realtime", "connection"},
			"auth":       {"auth", "authentication", "authorization", "login", "token", "jwt", "oauth"},
			"css":        {"css", "style", "styling", "sass", "scss", "less", "tailwind"},
			"ci":         {"ci", "cd", "pipeline", "github actions", "jenkins", "deploy"},
			"kubernetes": {"k8s", "kubernetes", "kubectl", "pod", "deployment", "service"},
		},
		domainPatterns: map[string][]string{
			"frontend":        {"ui", "frontend", "client", "browser", "dom", "rendering"},
			"backend":         {"backend", "server", "api", "service", "microservice"},
			"security":        {"security", "vulnerability", "encrypt", "decrypt", "hash", "secure"},
			"performance":     {"performance", "speed", "latency", "memory", "cpu", "optimization"},
			"networking":      {"network", "http", "https", "tcp", "udp", "socket", "connection"},
			"architecture":    {"architecture", "design", "pattern", "structure", "organization"},
			"monitoring":      {"monitoring", "logging", "metrics", "observability", "alerting"},
			"deployment":      {"deployment", "deploy", "release", "production", "staging"},
			"user-management": {"user", "users", "account", "profile", "permissions", "roles"},
			"data-processing": {"data", "processing", "pipeline", "etl", "transform", "parse"},
		},
	}
}

// AnalyzeTags generates smart tags based on content analysis
func (ta *TagAnalyzer) AnalyzeTags(content string) []string {
	content = strings.ToLower(content)
	var tags []string

	// Analyze for action patterns
	tags = append(tags, ta.findMatches(content, ta.actionPatterns)...)

	// Analyze for technology patterns
	tags = append(tags, ta.findMatches(content, ta.technologyPatterns)...)

	// Analyze for domain patterns
	tags = append(tags, ta.findMatches(content, ta.domainPatterns)...)

	// Remove duplicates and return
	return ta.removeDuplicates(tags)
}

// findMatches searches for pattern matches in content
func (ta *TagAnalyzer) findMatches(content string, patterns map[string][]string) []string {
	var matches []string

	for tag, keywords := range patterns {
		for _, keyword := range keywords {
			// Use word boundaries to avoid partial matches
			pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
			matched, _ := regexp.MatchString(pattern, content)
			if matched {
				matches = append(matches, tag)
				break // Only add tag once per category
			}
		}
	}

	return matches
}

// removeDuplicates removes duplicate tags from slice
func (ta *TagAnalyzer) removeDuplicates(tags []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, tag := range tags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	return result
}

// EnhanceTags adds smart tags to existing tags string
func (ta *TagAnalyzer) EnhanceTags(content, existingTags string) string {
	suggestedTags := ta.AnalyzeTags(content)

	if len(suggestedTags) == 0 {
		return existingTags
	}

	// Combine existing and suggested tags
	var allTags []string

	// Add existing tags if any
	if existingTags != "" {
		existingList := strings.Split(existingTags, ",")
		for _, tag := range existingList {
			allTags = append(allTags, strings.TrimSpace(tag))
		}
	}

	// Add suggested tags that aren't already present
	for _, suggested := range suggestedTags {
		found := false
		for _, existing := range allTags {
			if strings.ToLower(existing) == strings.ToLower(suggested) {
				found = true
				break
			}
		}
		if !found {
			allTags = append(allTags, suggested)
		}
	}

	return strings.Join(allTags, ",")
}

// GetSuggestedTags returns only the suggested tags for display purposes
func (ta *TagAnalyzer) GetSuggestedTags(content string) string {
	tags := ta.AnalyzeTags(content)
	if len(tags) == 0 {
		return ""
	}
	return strings.Join(tags, ",")
}
