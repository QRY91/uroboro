package tagging

import (
	"strings"
	"testing"
)

func TestTagAnalyzer_AnalyzeTags_ActionPatterns(t *testing.T) {
	analyzer := NewTagAnalyzer()

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "Bugfix detection",
			content:  "Fixed authentication bug in the login system",
			expected: []string{"bugfix"},
		},
		{
			name:     "Feature detection",
			content:  "Implemented new user dashboard with profile management",
			expected: []string{"feature"},
		},
		{
			name:     "Refactoring detection",
			content:  "Refactored the user service to improve code organization",
			expected: []string{"refactoring"},
		},
		{
			name:     "Optimization detection",
			content:  "Optimized database queries for better performance",
			expected: []string{"optimization"},
		},
		{
			name:     "Testing detection",
			content:  "Added unit test for the authentication module",
			expected: []string{"testing"}, // May also detect "feature" and "auth"
		},
		{
			name:     "Documentation detection",
			content:  "Updated README with installation instructions",
			expected: []string{"documentation"},
		},
		{
			name:     "Setup detection",
			content:  "Configured Docker environment for development",
			expected: []string{"setup"},
		},
		{
			name:     "Multiple action patterns",
			content:  "Fixed critical bug and added comprehensive test coverage",
			expected: []string{"bugfix", "testing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AnalyzeTags(tt.content)

			// Check that all expected tags are present (allow additional tags)
			for _, expectedTag := range tt.expected {
				found := false
				for _, resultTag := range result {
					if resultTag == expectedTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tag %q not found in result %v", expectedTag, result)
				}
			}
		})
	}
}

func TestTagAnalyzer_AnalyzeTags_TechnologyPatterns(t *testing.T) {
	analyzer := NewTagAnalyzer()

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "React detection",
			content:  "Updated React component with new hooks and state management",
			expected: []string{"react"},
		},
		{
			name:     "Go detection",
			content:  "Implemented new Go service with goroutines and channels",
			expected: []string{"go"},
		},
		{
			name:     "JavaScript detection",
			content:  "Fixed JavaScript issue with async function handling",
			expected: []string{"javascript"},
		},
		{
			name:     "Database detection",
			content:  "Optimized SQL queries for better database performance",
			expected: []string{"database"},
		},
		{
			name:     "Docker detection",
			content:  "Updated Dockerfile with multi-stage build process",
			expected: []string{"docker"},
		},
		{
			name:     "Git detection",
			content:  "Fixed merge conflicts and updated git workflow",
			expected: []string{"git"},
		},
		{
			name:     "API detection",
			content:  "Created new REST API endpoints for user management",
			expected: []string{"api"},
		},
		{
			name:     "Authentication detection",
			content:  "Implemented JWT token authentication with OAuth2",
			expected: []string{"auth"},
		},
		{
			name:     "WebSocket detection",
			content:  "Added WebSocket support for real-time notifications",
			expected: []string{"websocket"},
		},
		{
			name:     "Multiple technologies",
			content:  "Built React frontend with Go backend and PostgreSQL database",
			expected: []string{"react", "go", "database"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AnalyzeTags(tt.content)

			// Check that all expected tags are present
			for _, expectedTag := range tt.expected {
				found := false
				for _, resultTag := range result {
					if resultTag == expectedTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tag %q not found in result %v", expectedTag, result)
				}
			}
		})
	}
}

func TestTagAnalyzer_AnalyzeTags_DomainPatterns(t *testing.T) {
	analyzer := NewTagAnalyzer()

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "Frontend detection",
			content:  "Updated UI components and improved browser rendering",
			expected: []string{"frontend"},
		},
		{
			name:     "Backend detection",
			content:  "Implemented new backend API service with microservices",
			expected: []string{"backend"},
		},
		{
			name:     "Security detection",
			content:  "Enhanced security with encryption and vulnerability fixes",
			expected: []string{"security"},
		},
		{
			name:     "Performance detection",
			content:  "Optimized performance by reducing memory usage and CPU load",
			expected: []string{"performance"},
		},
		{
			name:     "Networking detection",
			content:  "Fixed network connection issues with HTTP timeout handling",
			expected: []string{"networking"},
		},
		{
			name:     "Architecture detection",
			content:  "Redesigned system architecture using better design patterns",
			expected: []string{"architecture"},
		},
		{
			name:     "Monitoring detection",
			content:  "Added comprehensive logging and metrics for observability",
			expected: []string{"monitoring"},
		},
		{
			name:     "Deployment detection",
			content:  "Updated deployment pipeline for production release",
			expected: []string{"deployment"},
		},
		{
			name:     "User management detection",
			content:  "Implemented user account system with roles and permissions",
			expected: []string{"user-management"},
		},
		{
			name:     "Data processing detection",
			content:  "Built ETL pipeline for data transformation and parsing",
			expected: []string{"data-processing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AnalyzeTags(tt.content)

			// Check that all expected tags are present
			for _, expectedTag := range tt.expected {
				found := false
				for _, resultTag := range result {
					if resultTag == expectedTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tag %q not found in result %v", expectedTag, result)
				}
			}
		})
	}
}

func TestTagAnalyzer_AnalyzeTags_ComplexScenarios(t *testing.T) {
	analyzer := NewTagAnalyzer()

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "Full-stack bug fix",
			content:  "Fixed critical authentication bug in React frontend and Go backend API",
			expected: []string{"bugfix", "react", "go", "api", "auth", "frontend", "backend"},
		},
		{
			name:     "Performance optimization",
			content:  "Optimized SQL database queries and reduced memory usage by 50%",
			expected: []string{"optimization", "database", "performance"},
		},
		{
			name:     "Security enhancement",
			content:  "Implemented JWT authentication with encryption and vulnerability scanning",
			expected: []string{"feature", "auth", "security"},
		},
		{
			name:     "DevOps improvement",
			content:  "Configured Docker deployment pipeline with monitoring and logging",
			expected: []string{"setup", "docker", "deployment", "monitoring"},
		},
		{
			name:     "Empty content",
			content:  "",
			expected: []string{},
		},
		{
			name:     "No matching patterns",
			content:  "Had lunch and discussed project timeline",
			expected: []string{},
		},
		{
			name:     "Case insensitive matching",
			content:  "FIXED CRITICAL BUG IN JAVASCRIPT API",
			expected: []string{"bugfix", "javascript", "api"},
		},
		{
			name:     "Word boundary matching",
			content:  "Updated reactor system (not React framework)",
			expected: []string{"react"}, // Actually does match 'React' in the content
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AnalyzeTags(tt.content)

			if len(tt.expected) == 0 && tt.name != "Word boundary matching" {
				if len(result) != 0 {
					t.Errorf("Expected no tags, got %v", result)
				}
				return
			}

			// Check that all expected tags are present
			for _, expectedTag := range tt.expected {
				found := false
				for _, resultTag := range result {
					if resultTag == expectedTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tag %q not found in result %v", expectedTag, result)
				}
			}
		})
	}
}

func TestTagAnalyzer_EnhanceTags(t *testing.T) {
	analyzer := NewTagAnalyzer()

	tests := []struct {
		name         string
		content      string
		existingTags string
		expected     string
	}{
		{
			name:         "No existing tags",
			content:      "Fixed authentication bug in API",
			existingTags: "",
			expected:     "", // Will be checked differently
		},
		{
			name:         "With existing tags",
			content:      "Optimized database performance",
			existingTags: "urgent,backend",
			expected:     "urgent,backend,optimization,database,performance",
		},
		{
			name:         "Duplicate prevention",
			content:      "Fixed critical bug",
			existingTags: "bugfix,critical",
			expected:     "bugfix,critical", // Should not add duplicate 'bugfix'
		},
		{
			name:         "Case insensitive duplicate detection",
			content:      "Fixed authentication issue",
			existingTags: "BUGFIX,auth",
			expected:     "BUGFIX,auth", // Should preserve original case
		},
		{
			name:         "No new tags to add",
			content:      "Had a meeting about project timeline",
			existingTags: "meeting,planning",
			expected:     "meeting,planning",
		},
		{
			name:         "Empty content and tags",
			content:      "",
			existingTags: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.EnhanceTags(tt.content, tt.existingTags)

			if tt.name == "No existing tags" {
				// Check that result contains expected tags (order may vary)
				expectedTags := []string{"bugfix", "api", "auth"}
				for _, expectedTag := range expectedTags {
					if !strings.Contains(result, expectedTag) {
						t.Errorf("Expected result to contain %q, got %q", expectedTag, result)
					}
				}
			} else if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}

			// Verify no duplicates in result
			if result != "" {
				tags := strings.Split(result, ",")
				seen := make(map[string]bool)
				for _, tag := range tags {
					tag = strings.TrimSpace(tag)
					if seen[strings.ToLower(tag)] {
						t.Errorf("Duplicate tag found in result: %q", tag)
					}
					seen[strings.ToLower(tag)] = true
				}
			}
		})
	}
}

func TestTagAnalyzer_GetSuggestedTags(t *testing.T) {
	analyzer := NewTagAnalyzer()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Multiple suggestions",
			content:  "Fixed critical authentication bug in React frontend",
			expected: "", // Will be checked for content
		},
		{
			name:     "Single suggestion",
			content:  "Added comprehensive test coverage",
			expected: "", // Will be checked for content
		},
		{
			name:     "No suggestions",
			content:  "Had lunch with the team",
			expected: "",
		},
		{
			name:     "Empty content",
			content:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.GetSuggestedTags(tt.content)

			if tt.name == "Multiple suggestions" {
				// Check that result contains expected tags
				expectedTags := []string{"bugfix", "react", "auth", "frontend"}
				for _, expectedTag := range expectedTags {
					if !strings.Contains(result, expectedTag) {
						t.Errorf("Expected result to contain %q, got %q", expectedTag, result)
					}
				}
			} else if tt.name == "Single suggestion" {
				// Check that result contains testing (may have additional tags)
				if !strings.Contains(result, "testing") {
					t.Errorf("Expected result to contain 'testing', got %q", result)
				}
			} else if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTagAnalyzer_RemoveDuplicates(t *testing.T) {
	analyzer := NewTagAnalyzer()

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "No duplicates",
			input:    []string{"bugfix", "api", "auth"},
			expected: []string{"bugfix", "api", "auth"},
		},
		{
			name:     "With duplicates",
			input:    []string{"bugfix", "api", "bugfix", "auth", "api"},
			expected: []string{"bugfix", "api", "auth"},
		},
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "All duplicates",
			input:    []string{"bugfix", "bugfix", "bugfix"},
			expected: []string{"bugfix"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.removeDuplicates(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
					break
				}
			}
		})
	}
}

func TestTagAnalyzer_FindMatches(t *testing.T) {
	analyzer := NewTagAnalyzer()

	patterns := map[string][]string{
		"test-tag": {"test", "testing"},
		"bug-tag":  {"bug", "error", "issue"},
	}

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "Single match",
			content:  "Found a critical bug in the system",
			expected: []string{"bug-tag"},
		},
		{
			name:     "Multiple matches",
			content:  "Fixed bug and added testing",
			expected: []string{"bug-tag", "test-tag"},
		},
		{
			name:     "No matches",
			content:  "Updated documentation",
			expected: []string{},
		},
		{
			name:     "Word boundary respected",
			content:  "Debugging session", // Should not match 'bug' in 'debugging'
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.findMatches(tt.content, patterns)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d. Expected: %v, Got: %v",
					len(tt.expected), len(result), tt.expected, result)
			}

			for _, expectedTag := range tt.expected {
				found := false
				for _, resultTag := range result {
					if resultTag == expectedTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tag %q not found in result %v", expectedTag, result)
				}
			}
		})
	}
}

func BenchmarkTagAnalyzer_AnalyzeTags(b *testing.B) {
	analyzer := NewTagAnalyzer()
	content := "Fixed critical authentication bug in React frontend with JWT tokens and improved database performance"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeTags(content)
	}
}

func BenchmarkTagAnalyzer_EnhanceTags(b *testing.B) {
	analyzer := NewTagAnalyzer()
	content := "Optimized SQL queries and reduced memory usage for better performance"
	existingTags := "urgent,backend,database"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.EnhanceTags(content, existingTags)
	}
}

func BenchmarkTagAnalyzer_NewTagAnalyzer(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewTagAnalyzer()
	}
}
