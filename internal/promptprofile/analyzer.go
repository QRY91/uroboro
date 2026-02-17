package promptprofile

import (
	"fmt"
	"strings"
)

var imperativeVerbs = []string{
	"fix", "add", "implement", "create", "update", "remove", "delete",
	"refactor", "change", "make", "build", "write", "run", "test",
	"check", "move", "rename", "extract", "clean", "simplify",
	"merge", "deploy", "configure", "setup", "install", "upgrade",
	"migrate", "convert", "replace", "review", "debug", "optimize",
	"search", "find", "list", "show", "help", "explain", "design",
	"plan", "analyze", "summarize", "generate", "format", "validate",
	"read", "open", "start", "stop", "set", "get", "use", "try",
	"apply", "commit", "push", "pull", "revert", "reset", "look",
}

var directivePhrases = []string{
	"we need to", "let's", "can you", "could you", "please",
	"go ahead", "i want to", "i need", "i'd like", "we should",
}

var questionStarters = []string{
	"how", "what", "why", "when", "where", "which",
	"can", "could", "should", "would", "is", "are",
	"do", "does", "did", "will", "has", "have",
}

// classify populates the classification fields on a UserPrompt.
func classify(p *UserPrompt) {
	lower := strings.ToLower(p.Text)

	p.IsImperative = isImperative(lower)
	p.IsQuestion = isQuestion(lower)
	p.HasFilePath = hasFilePath(p.Text)
	p.HasCodeBlock = strings.Contains(p.Text, "```")
}

func isImperative(lower string) bool {
	firstWord := strings.Fields(lower)
	if len(firstWord) == 0 {
		return false
	}
	w := firstWord[0]
	for _, v := range imperativeVerbs {
		if w == v {
			return true
		}
	}
	// Check directive phrases
	for _, phrase := range directivePhrases {
		if strings.HasPrefix(lower, phrase) {
			return true
		}
	}
	return false
}

func isQuestion(lower string) bool {
	if strings.Contains(lower, "?") {
		return true
	}
	firstWord := strings.Fields(lower)
	if len(firstWord) == 0 {
		return false
	}
	w := firstWord[0]
	for _, q := range questionStarters {
		if w == q {
			return true
		}
	}
	return false
}

func hasFilePath(text string) bool {
	// Check for common path patterns
	for _, word := range strings.Fields(text) {
		if strings.HasPrefix(word, "/") && strings.Contains(word, "/") && len(word) > 2 {
			return true
		}
		if strings.HasPrefix(word, "./") || strings.HasPrefix(word, "../") {
			return true
		}
		// filename.ext pattern (at least one dot with extension)
		if strings.Contains(word, ".") && !strings.HasPrefix(word, "http") {
			parts := strings.Split(word, ".")
			ext := parts[len(parts)-1]
			ext = strings.TrimRight(ext, ",:;)\"'")
			switch ext {
			case "go", "py", "ts", "tsx", "js", "jsx", "rs", "rb", "java",
				"c", "cpp", "h", "cs", "sh", "sql", "html", "css",
				"yaml", "yml", "json", "toml", "md", "txt", "cfg",
				"jsonl", "csv", "xml", "env", "lock", "mod", "sum":
				return true
			}
		}
	}
	return false
}

// Analyze computes aggregate statistics from a slice of prompts.
func Analyze(prompts []UserPrompt) *PromptStats {
	stats := &PromptStats{
		ProjectCounts: make(map[string]int),
		HourCounts:    make(map[int]int),
		WeekdayCounts: make(map[string]int),
	}

	sessionSeen := make(map[string]bool)

	for _, p := range prompts {
		stats.TotalPrompts++
		stats.TotalWords += p.WordCount

		if !sessionSeen[p.SessionID] {
			sessionSeen[p.SessionID] = true
			stats.TotalSessions++
		}

		// Date range
		if stats.EarliestDate.IsZero() || p.Timestamp.Before(stats.EarliestDate) {
			stats.EarliestDate = p.Timestamp
		}
		if p.Timestamp.After(stats.LatestDate) {
			stats.LatestDate = p.Timestamp
		}

		// Length distribution
		switch {
		case p.WordCount < 20:
			stats.ShortPrompts++
		case p.WordCount < 100:
			stats.MediumPrompts++
		case p.WordCount < 500:
			stats.LongPrompts++
		default:
			stats.VeryLong++
		}

		// Classification counts
		if p.IsImperative {
			stats.ImperativeCount++
		}
		if p.IsQuestion {
			stats.QuestionCount++
		}
		if p.HasFilePath {
			stats.FileRefCount++
		}
		if p.HasCodeBlock {
			stats.CodeBlockCount++
		}
		if p.IsFirstMsg && p.WordCount > 100 {
			stats.ContextOpeners++
		}

		// Project distribution
		stats.ProjectCounts[p.Project]++

		// Temporal
		stats.HourCounts[p.Timestamp.Hour()]++
		stats.WeekdayCounts[p.Timestamp.Weekday().String()]++
	}

	// Averages
	if stats.TotalPrompts > 0 {
		stats.AvgWordsPerPrompt = float64(stats.TotalWords) / float64(stats.TotalPrompts)
	}
	if stats.TotalSessions > 0 {
		stats.AvgPromptsPerSession = float64(stats.TotalPrompts) / float64(stats.TotalSessions)
	}

	return stats
}

// FormatStats produces a terminal-friendly stats summary.
func FormatStats(stats *PromptStats) string {
	var b strings.Builder

	b.WriteString(strings.Repeat("-", 60) + "\n")
	b.WriteString(fmt.Sprintf("  PROMPT PROFILE\n"))
	b.WriteString(fmt.Sprintf("  %d sessions | %d prompts | since %s\n",
		stats.TotalSessions, stats.TotalPrompts,
		stats.EarliestDate.Format("2006-01-02")))
	b.WriteString(strings.Repeat("-", 60) + "\n\n")

	// Length distribution
	b.WriteString("## Length Distribution\n")
	total := float64(stats.TotalPrompts)
	if total == 0 {
		total = 1
	}
	b.WriteString(fmt.Sprintf("  Short (<20 words):     %4.0f%%  %s\n",
		pct(stats.ShortPrompts, stats.TotalPrompts), bar(stats.ShortPrompts, stats.TotalPrompts)))
	b.WriteString(fmt.Sprintf("  Medium (20-100):       %4.0f%%  %s\n",
		pct(stats.MediumPrompts, stats.TotalPrompts), bar(stats.MediumPrompts, stats.TotalPrompts)))
	b.WriteString(fmt.Sprintf("  Long (100-500):        %4.0f%%  %s\n",
		pct(stats.LongPrompts, stats.TotalPrompts), bar(stats.LongPrompts, stats.TotalPrompts)))
	b.WriteString(fmt.Sprintf("  Very long (500+):      %4.0f%%  %s\n",
		pct(stats.VeryLong, stats.TotalPrompts), bar(stats.VeryLong, stats.TotalPrompts)))
	b.WriteString(fmt.Sprintf("  Average: %.0f words\n\n", stats.AvgWordsPerPrompt))

	// Style
	b.WriteString("## Style\n")
	b.WriteString(fmt.Sprintf("  Imperative:            %4.0f%%  %s\n",
		pct(stats.ImperativeCount, stats.TotalPrompts), bar(stats.ImperativeCount, stats.TotalPrompts)))
	b.WriteString(fmt.Sprintf("  Questions:             %4.0f%%  %s\n",
		pct(stats.QuestionCount, stats.TotalPrompts), bar(stats.QuestionCount, stats.TotalPrompts)))
	b.WriteString("\n")

	// Patterns
	b.WriteString("## Patterns\n")
	b.WriteString(fmt.Sprintf("  File references:       %4.0f%% of prompts\n",
		pct(stats.FileRefCount, stats.TotalPrompts)))
	b.WriteString(fmt.Sprintf("  Code blocks:           %4.0f%% of prompts\n",
		pct(stats.CodeBlockCount, stats.TotalPrompts)))
	if stats.TotalSessions > 0 {
		b.WriteString(fmt.Sprintf("  Context-rich openers:  %4.0f%% of sessions\n",
			pct(stats.ContextOpeners, stats.TotalSessions)))
	}
	b.WriteString(fmt.Sprintf("  Avg prompts/session:   %.1f\n", stats.AvgPromptsPerSession))
	b.WriteString("\n")

	// Top projects
	b.WriteString("## Top Projects\n")
	type kv struct {
		k string
		v int
	}
	var sorted []kv
	for k, v := range stats.ProjectCounts {
		sorted = append(sorted, kv{k, v})
	}
	// Sort by count descending
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].v > sorted[i].v {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	max := 10
	if len(sorted) < max {
		max = len(sorted)
	}
	for i := 0; i < max; i++ {
		b.WriteString(fmt.Sprintf("  %2d. %-30s %d\n", i+1, sorted[i].k, sorted[i].v))
	}
	if len(sorted) > max {
		b.WriteString(fmt.Sprintf("  ...%d more\n", len(sorted)-max))
	}
	b.WriteString("\n")

	// Active times
	b.WriteString("## Active Times\n")
	peakHour, peakCount := 0, 0
	for h, c := range stats.HourCounts {
		if c > peakCount {
			peakHour = h
			peakCount = c
		}
	}
	b.WriteString(fmt.Sprintf("  Peak hour:  %02d:00 (%d prompts)\n", peakHour, peakCount))

	peakDay, peakDayCount := "", 0
	for d, c := range stats.WeekdayCounts {
		if c > peakDayCount {
			peakDay = d
			peakDayCount = c
		}
	}
	b.WriteString(fmt.Sprintf("  Peak day:   %s (%d prompts)\n", peakDay, peakDayCount))

	return b.String()
}

func pct(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total) * 100
}

func bar(n, total int) string {
	if total == 0 {
		return ""
	}
	width := int(float64(n) / float64(total) * 20)
	if width == 0 && n > 0 {
		width = 1
	}
	return strings.Repeat("#", width)
}
