package feast

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/database"
)

// FeastEngine handles archiving old captures to prevent cognitive overload
type FeastEngine struct {
	db     *database.DB
	config *FeastConfig
}

// FeastConfig holds configuration for feast behavior
type FeastConfig struct {
	AutoEnabled      bool
	AgeThresholdDays int
	ShowDigest       bool
	MaxDigestItems   int
	RescueEnabled    bool
	SilentMode       bool
}

// DigestResult holds the result of showing digest to user
type DigestResult struct {
	RescuedItems []database.Capture
	ArchiveAll   bool
	Skip         bool
}

// DefaultFeastConfig returns sensible defaults
func DefaultFeastConfig() *FeastConfig {
	return &FeastConfig{
		AutoEnabled:      true,
		AgeThresholdDays: 30,
		ShowDigest:       true,
		MaxDigestItems:   10,
		RescueEnabled:    true,
		SilentMode:       false,
	}
}

// NewFeastEngine creates a new feast engine
func NewFeastEngine(db *database.DB, config *FeastConfig) *FeastEngine {
	if config == nil {
		config = DefaultFeastConfig()
	}
	return &FeastEngine{
		db:     db,
		config: config,
	}
}

// isTerminal checks if we're running in an interactive terminal
func isTerminal() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	// Check if stdin is a character device (terminal)
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// colorize adds ANSI color codes if terminal supports it
func colorize(text, color string) string {
	if !isTerminal() {
		return text
	}
	colors := map[string]string{
		"green":  "\033[32m",
		"yellow": "\033[33m",
		"blue":   "\033[34m",
		"red":    "\033[31m",
		"bold":   "\033[1m",
		"reset":  "\033[0m",
	}
	if code, exists := colors[color]; exists {
		return code + text + colors["reset"]
	}
	return text
}

// EnsureArchiveTable creates the archived_captures table if it doesn't exist
func (f *FeastEngine) EnsureArchiveTable() error {
	// Check if table exists and has foreign key constraints
	var hasTable bool
	err := f.db.QueryRow(`
		SELECT COUNT(*) > 0 FROM sqlite_master
		WHERE type='table' AND name='archived_captures'
	`).Scan(&hasTable)

	if err != nil {
		return fmt.Errorf("failed to check if archive table exists: %w", err)
	}

	if hasTable {
		// Check if table has foreign key constraints by examining schema
		var tableSchema string
		err := f.db.QueryRow(`
			SELECT sql FROM sqlite_master
			WHERE type='table' AND name='archived_captures'
		`).Scan(&tableSchema)

		if err != nil {
			return fmt.Errorf("failed to get table schema: %w", err)
		}

		// If the table has foreign key constraints, recreate it
		if strings.Contains(tableSchema, "FOREIGN KEY") {
			// Backup existing data
			_, err = f.db.Exec(`
				CREATE TEMPORARY TABLE archived_captures_backup AS
				SELECT * FROM archived_captures
			`)
			if err != nil {
				return fmt.Errorf("failed to backup archive table: %w", err)
			}

			// Drop the table
			_, err = f.db.Exec("DROP TABLE archived_captures")
			if err != nil {
				return fmt.Errorf("failed to drop archive table: %w", err)
			}

			// Create new table without foreign key
			schema := `
			CREATE TABLE archived_captures (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				original_id INTEGER NOT NULL,
				timestamp DATETIME NOT NULL,
				content TEXT NOT NULL,
				project TEXT,
				tags TEXT,
				source_tool TEXT,
				metadata TEXT,
				archived_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				archive_reason TEXT DEFAULT 'auto_feast'
			);

			CREATE INDEX idx_archived_captures_archived_at ON archived_captures(archived_at);
			CREATE INDEX idx_archived_captures_project ON archived_captures(project);
			CREATE INDEX idx_archived_captures_original_id ON archived_captures(original_id);
			`

			_, err = f.db.Exec(schema)
			if err != nil {
				return fmt.Errorf("failed to recreate archive table: %w", err)
			}

			// Restore data
			_, err = f.db.Exec(`
				INSERT INTO archived_captures
				SELECT * FROM archived_captures_backup
			`)
			if err != nil {
				return fmt.Errorf("failed to restore archive data: %w", err)
			}

			// Drop backup table
			_, err = f.db.Exec("DROP TABLE archived_captures_backup")
			if err != nil {
				return fmt.Errorf("failed to drop backup table: %w", err)
			}
		}
	} else {
		// Create table from scratch
		schema := `
		CREATE TABLE archived_captures (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			original_id INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			content TEXT NOT NULL,
			project TEXT,
			tags TEXT,
			source_tool TEXT,
			metadata TEXT,
			archived_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			archive_reason TEXT DEFAULT 'auto_feast'
		);

		CREATE INDEX idx_archived_captures_archived_at ON archived_captures(archived_at);
		CREATE INDEX idx_archived_captures_project ON archived_captures(project);
		CREATE INDEX idx_archived_captures_original_id ON archived_captures(original_id);
		`

		_, err = f.db.Exec(schema)
		if err != nil {
			return fmt.Errorf("failed to create archive table: %w", err)
		}
	}

	return nil
}

// GetItemsForArchive returns captures older than specified days
func (f *FeastEngine) GetItemsForArchive(days int) ([]database.Capture, error) {
	cutoffDate := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT id, timestamp, content, project, tags, source_tool, metadata, created_at, updated_at
		FROM captures
		WHERE timestamp < ?
		ORDER BY timestamp DESC
	`

	rows, err := f.db.Query(query, cutoffDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query old captures: %w", err)
	}
	defer rows.Close()

	var captures []database.Capture
	for rows.Next() {
		var c database.Capture
		var tags, metadata sql.NullString

		err := rows.Scan(
			&c.ID, &c.Timestamp, &c.Content, &c.Project,
			&tags, &c.SourceTool, &metadata, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan capture: %w", err)
		}

		c.Tags = tags
		c.Metadata = metadata
		captures = append(captures, c)
	}

	return captures, nil
}

// ShowDigest displays items ready for archiving and handles user interaction
func (f *FeastEngine) ShowDigest(items []database.Capture) (DigestResult, error) {
	if len(items) == 0 {
		return DigestResult{ArchiveAll: true}, nil
	}

	if f.config.SilentMode {
		return DigestResult{ArchiveAll: true}, nil
	}

	// Check if we're in an interactive terminal
	if !isTerminal() {
		fmt.Printf("🐍 Auto-feast: %d items ready for archive (non-interactive mode)\n", len(items))
		fmt.Println("   Defaulting to archive all items")
		return DigestResult{ArchiveAll: true}, nil
	}

	// Show digest header with colors
	fmt.Printf("%s\n", colorize("🐍 Auto-feast digest", "bold"))
	fmt.Printf("   %s items ready for archive\n\n", colorize(fmt.Sprintf("%d", len(items)), "yellow"))

	// Show up to MaxDigestItems
	displayCount := len(items)
	if displayCount > f.config.MaxDigestItems {
		displayCount = f.config.MaxDigestItems
	}

	for i := 0; i < displayCount; i++ {
		item := items[i]
		timeAgo := formatTimeAgo(item.Timestamp)
		project := "no project"
		if item.Project.Valid && item.Project.String != "" {
			project = item.Project.String
		}

		fmt.Printf("   ✨ %s: \"%s\" %s\n",
			colorize(timeAgo, "blue"),
			truncateString(item.Content, 60),
			colorize("["+project+"]", "green"))
	}

	if len(items) > displayCount {
		fmt.Printf("   %s\n", colorize(fmt.Sprintf("... and %d more items", len(items)-displayCount), "yellow"))
	}

	fmt.Println()

	if !f.config.RescueEnabled {
		fmt.Printf("%s ", colorize("Press Enter to archive all, 's' to skip:", "bold"))
		return f.handleSimpleInput(items)
	}

	fmt.Printf("%s ", colorize("Press 'r' to rescue items, 's' to skip digest, Enter to archive all:", "bold"))
	return f.handleDigestInput(items)
}

// handleSimpleInput handles input when rescue is disabled
func (f *FeastEngine) handleSimpleInput(items []database.Capture) (DigestResult, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return DigestResult{}, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "s", "skip":
		fmt.Printf("%s\n", colorize("🐍 Feast skipped", "yellow"))
		return DigestResult{Skip: true}, nil
	default:
		return DigestResult{ArchiveAll: true}, nil
	}
}

// handleDigestInput handles user input for digest interaction
func (f *FeastEngine) handleDigestInput(items []database.Capture) (DigestResult, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return DigestResult{}, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "r", "rescue":
		return f.handleRescue(items)
	case "s", "skip":
		fmt.Printf("%s\n", colorize("🐍 Feast skipped", "yellow"))
		return DigestResult{Skip: true}, nil
	default:
		return DigestResult{ArchiveAll: true}, nil
	}
}

// handleRescue allows user to select items to rescue from archiving
func (f *FeastEngine) handleRescue(items []database.Capture) (DigestResult, error) {
	fmt.Printf("\n%s\n", colorize("Items available for rescue:", "bold"))
	displayCount := len(items)
	if displayCount > f.config.MaxDigestItems {
		displayCount = f.config.MaxDigestItems
	}

	for i := 0; i < displayCount; i++ {
		item := items[i]
		timeAgo := formatTimeAgo(item.Timestamp)
		project := "no project"
		if item.Project.Valid && item.Project.String != "" {
			project = item.Project.String
		}

		fmt.Printf("   %s %s: \"%s\" %s\n",
			colorize(fmt.Sprintf("%d.", i+1), "yellow"),
			colorize(timeAgo, "blue"),
			truncateString(item.Content, 60),
			colorize("["+project+"]", "green"))
	}

	fmt.Printf("\n%s ", colorize("Enter numbers to rescue (comma-separated, e.g., 1,3,5):", "bold"))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return DigestResult{}, fmt.Errorf("failed to read rescue input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return DigestResult{ArchiveAll: true}, nil
	}

	// Parse rescue numbers
	var rescuedItems []database.Capture
	numbers := strings.Split(input, ",")

	for _, numStr := range numbers {
		numStr = strings.TrimSpace(numStr)
		num, err := strconv.Atoi(numStr)
		if err != nil || num < 1 || num > displayCount {
			fmt.Printf("%s: %s\n", colorize("Ignoring invalid number", "red"), numStr)
			continue
		}

		rescuedItems = append(rescuedItems, items[num-1])
	}

	if len(rescuedItems) > 0 {
		fmt.Printf("%s %d items\n", colorize("✅ Rescued:", "green"), len(rescuedItems))
	}

	return DigestResult{
		RescuedItems: rescuedItems,
		ArchiveAll:   false,
	}, nil
}

// ArchiveCaptures moves captures to archived_captures table
func (f *FeastEngine) ArchiveCaptures(captures []database.Capture, reason string) error {
	if len(captures) == 0 {
		return nil
	}

	// Begin transaction - note: we'll do individual operations for now
	// TODO: Implement proper transaction support in database wrapper
	return f.archiveIndividually(captures, reason)

}

// archiveIndividually archives captures one by one (fallback)
func (f *FeastEngine) archiveIndividually(captures []database.Capture, reason string) error {
	for _, capture := range captures {
		// Handle nullable fields
		var project, tags, metadata interface{}
		if capture.Project.Valid {
			project = capture.Project.String
		} else {
			project = nil
		}
		if capture.Tags.Valid {
			tags = capture.Tags.String
		} else {
			tags = nil
		}
		if capture.Metadata.Valid {
			metadata = capture.Metadata.String
		} else {
			metadata = nil
		}

		// Insert into archive
		_, err := f.db.Exec(`
			INSERT INTO archived_captures (
				original_id, timestamp, content, project, tags,
				source_tool, metadata, archive_reason
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			capture.ID, capture.Timestamp, capture.Content, project,
			tags, capture.SourceTool, metadata, reason,
		)
		if err != nil {
			return fmt.Errorf("failed to archive capture %d: %w", capture.ID, err)
		}

		// Delete from active captures
		_, err = f.db.Exec("DELETE FROM captures WHERE id = ?", capture.ID)
		if err != nil {
			// Try to clean up the archive entry
			f.db.Exec("DELETE FROM archived_captures WHERE original_id = ?", capture.ID)
			return fmt.Errorf("failed to delete capture %d: %w", capture.ID, err)
		}
	}
	return nil
}

// AutoFeastCheck performs automatic feast if conditions are met
func (f *FeastEngine) AutoFeastCheck() error {
	if !f.config.AutoEnabled {
		return nil
	}

	// Ensure archive table exists
	if err := f.EnsureArchiveTable(); err != nil {
		return fmt.Errorf("failed to ensure archive table: %w", err)
	}

	items, err := f.GetItemsForArchive(f.config.AgeThresholdDays)
	if err != nil {
		return fmt.Errorf("failed to get items for archive: %w", err)
	}

	if len(items) == 0 {
		return nil // Nothing to archive
	}

	if !f.config.ShowDigest {
		// Silent auto-archive
		return f.ArchiveCaptures(items, "auto_feast")
	}

	// Show digest and handle user interaction
	result, err := f.ShowDigest(items)
	if err != nil {
		return fmt.Errorf("failed to show digest: %w", err)
	}

	if result.Skip {
		return nil
	}

	// Handle rescue logic
	itemsToArchive := items
	if len(result.RescuedItems) > 0 {
		itemsToArchive = f.removeRescuedItems(items, result.RescuedItems)
		fmt.Printf("Rescued: %d items kept active\n", len(result.RescuedItems))
	}

	if len(itemsToArchive) > 0 {
		// Show progress for large operations
		if len(itemsToArchive) > 50 {
			fmt.Printf("🐍 Archiving %d items...\n", len(itemsToArchive))
		}

		err = f.ArchiveCaptures(itemsToArchive, "auto_feast")
		if err != nil {
			return fmt.Errorf("failed to archive captures: %w", err)
		}

		fmt.Printf("%s\n", colorize(fmt.Sprintf("🐍 Feasted on %d items", len(itemsToArchive)), "green"))
		fmt.Printf("   %s\n", colorize("The snake eats its tail", "blue"))
	}

	return nil
}

// ManualFeast performs manual feast with options
func (f *FeastEngine) ManualFeast(days int, silent bool) error {
	// Ensure archive table exists
	if err := f.EnsureArchiveTable(); err != nil {
		return fmt.Errorf("failed to ensure archive table: %w", err)
	}

	items, err := f.GetItemsForArchive(days)
	if err != nil {
		return fmt.Errorf("failed to get items for archive: %w", err)
	}

	if len(items) == 0 {
		fmt.Println("🐍 Nothing to feast on - the snake rests")
		return nil
	}

	if silent {
		err = f.ArchiveCaptures(items, "manual_feast")
		if err != nil {
			return fmt.Errorf("failed to archive captures: %w", err)
		}
		fmt.Printf("🐍 FEAST: Archived %d items silently\n", len(items))
		return nil
	}

	// Show digest and handle interaction
	tempConfig := *f.config
	tempConfig.SilentMode = false
	f.config = &tempConfig

	result, err := f.ShowDigest(items)
	if err != nil {
		return fmt.Errorf("failed to show digest: %w", err)
	}

	if result.Skip {
		fmt.Println("🐍 Feast skipped")
		return nil
	}

	// Handle rescue logic
	itemsToArchive := items
	if len(result.RescuedItems) > 0 {
		itemsToArchive = f.removeRescuedItems(items, result.RescuedItems)
		fmt.Printf("Rescued: %d items kept active\n", len(result.RescuedItems))
	}

	if len(itemsToArchive) > 0 {
		// Show progress for large operations
		if len(itemsToArchive) > 50 {
			fmt.Printf("🐍 Archiving %d items...\n", len(itemsToArchive))
		}

		err = f.ArchiveCaptures(itemsToArchive, "manual_feast")
		if err != nil {
			return fmt.Errorf("failed to archive captures: %w", err)
		}

		fmt.Printf("%s\n", colorize(fmt.Sprintf("🐍 FEAST: Consumed %d items", len(itemsToArchive)), "green"))
		fmt.Printf("   %s\n", colorize("The snake eats its tail", "blue"))
	} else {
		fmt.Printf("%s\n", colorize("🐍 All items rescued - nothing archived", "yellow"))
	}

	return nil
}

// removeRescuedItems removes rescued items from the list to archive
func (f *FeastEngine) removeRescuedItems(allItems, rescuedItems []database.Capture) []database.Capture {
	rescuedIDs := make(map[int64]bool)
	for _, rescued := range rescuedItems {
		rescuedIDs[rescued.ID] = true
	}

	var result []database.Capture
	for _, item := range allItems {
		if !rescuedIDs[item.ID] {
			result = append(result, item)
		}
	}
	return result
}

// Utility functions

func formatTimeAgo(timestamp time.Time) string {
	now := time.Now()
	diff := now.Sub(timestamp)

	days := int(diff.Hours() / 24)
	if days > 0 {
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	hours := int(diff.Hours())
	if hours > 0 {
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	return "recently"
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
