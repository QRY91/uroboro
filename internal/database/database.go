package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/QRY91/uroboro/internal/common"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

type Capture struct {
	ID         int64          `json:"id"`
	Timestamp  time.Time      `json:"timestamp"`
	Content    string         `json:"content"`
	Project    sql.NullString `json:"project"`
	Tags       sql.NullString `json:"tags"`
	SourceTool string         `json:"source_tool"`
	Metadata   sql.NullString `json:"metadata"` // JSON string
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type Publication struct {
	ID             int64     `json:"id"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Format         string    `json:"format"`
	Type           string    `json:"type"`
	SourceCaptures string    `json:"source_captures"` // JSON array of capture IDs
	Project        string    `json:"project"`
	TargetPath     string    `json:"target_path"`
	CreatedAt      time.Time `json:"created_at"`
}

// Initialize database connection
func NewDB(dbPath string) (*DB, error) {
	// If no path specified, use default location
	if dbPath == "" {
		dataDir := common.GetDataDir()
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
		dbPath = filepath.Join(dataDir, "uroboro.sqlite")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	dbInstance := &DB{db}

	// Run migrations
	if err := dbInstance.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return dbInstance, nil
}

// Run database migrations
func (db *DB) migrate() error {
	// Check if migrations table exists
	var tableExists bool
	err := db.QueryRow(`
		SELECT COUNT(*) > 0 FROM sqlite_master 
		WHERE type='table' AND name='schema_migrations'
	`).Scan(&tableExists)

	if err != nil {
		return fmt.Errorf("failed to check migrations table: %w", err)
	}

	if !tableExists {
		// Create initial schema
		if err := db.createInitialSchema(); err != nil {
			return fmt.Errorf("failed to create initial schema: %w", err)
		}
	}

	return nil
}

// Create the initial database schema
func (db *DB) createInitialSchema() error {
	schema := `
	-- Core captures table
	CREATE TABLE captures (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		content TEXT NOT NULL,
		project TEXT,
		tags TEXT,
		source_tool TEXT DEFAULT 'uroboro',
		metadata TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Published content
	CREATE TABLE publications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		format TEXT NOT NULL,
		type TEXT NOT NULL,
		source_captures TEXT,
		project TEXT,
		target_path TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Projects
	CREATE TABLE projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		description TEXT,
		git_path TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_activity DATETIME
	);

	-- Cross-tool communication
	CREATE TABLE tool_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_tool TEXT NOT NULL,
		to_tool TEXT NOT NULL,
		message_type TEXT NOT NULL,
		data TEXT NOT NULL,
		processed BOOLEAN DEFAULT FALSE,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		processed_at DATETIME
	);

	-- Usage tracking
	CREATE TABLE usage_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		command TEXT NOT NULL,
		project TEXT,
		duration_ms INTEGER,
		success BOOLEAN DEFAULT TRUE,
		error_message TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Indexes
	CREATE INDEX idx_captures_timestamp ON captures(timestamp);
	CREATE INDEX idx_captures_project ON captures(project);
	CREATE INDEX idx_captures_source_tool ON captures(source_tool);
	CREATE INDEX idx_publications_type ON publications(type);
	CREATE INDEX idx_publications_project ON publications(project);
	CREATE INDEX idx_tool_messages_to_tool ON tool_messages(to_tool, processed);
	CREATE INDEX idx_projects_name ON projects(name);

	-- Migration tracking
	CREATE TABLE schema_migrations (
		version INTEGER PRIMARY KEY,
		description TEXT NOT NULL,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Initial migration record
	INSERT INTO schema_migrations (version, description) 
	VALUES (1, 'Initial schema with captures, publications, projects, and cross-tool communication');
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// Insert a new capture
func (db *DB) InsertCapture(content, project, tags string) (*Capture, error) {
	timestamp := time.Now()

	query := `
		INSERT INTO captures (content, project, tags, timestamp)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.Exec(query, content, project, tags, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to insert capture: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get insert ID: %w", err)
	}

	// Create the capture struct with the known values
	capture := &Capture{
		ID:         id,
		Timestamp:  timestamp,
		Content:    content,
		Project:    sql.NullString{String: project, Valid: project != ""},
		Tags:       sql.NullString{String: tags, Valid: tags != ""},
		SourceTool: "uroboro",
		Metadata:   sql.NullString{String: "", Valid: false},
		CreatedAt:  timestamp,
		UpdatedAt:  timestamp,
	}

	return capture, nil
}

// Get recent captures for publishing
func (db *DB) GetRecentCaptures(days int, project string) ([]Capture, error) {
	query := `
		SELECT id, timestamp, content, project, tags, source_tool, metadata, created_at, updated_at
		FROM captures 
		WHERE timestamp >= datetime('now', '-' || ? || ' days')
	`
	args := []interface{}{days}

	if project != "" {
		query += " AND project = ?"
		args = append(args, project)
	}

	query += " ORDER BY timestamp DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query captures: %w", err)
	}
	defer rows.Close()

	var captures []Capture
	for rows.Next() {
		var capture Capture
		err := rows.Scan(
			&capture.ID,
			&capture.Timestamp,
			&capture.Content,
			&capture.Project,
			&capture.Tags,
			&capture.SourceTool,
			&capture.Metadata,
			&capture.CreatedAt,
			&capture.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan capture: %w", err)
		}
		captures = append(captures, capture)
	}

	return captures, nil
}

// Insert a publication record
func (db *DB) InsertPublication(title, content, format, pubType, project, targetPath string, sourceCaptureIDs []int64) (*Publication, error) {
	// Convert capture IDs to JSON string
	sourceCaptures := "[]"
	if len(sourceCaptureIDs) > 0 {
		sourceCaptures = fmt.Sprintf("[%v]", sourceCaptureIDs[0])
		for i := 1; i < len(sourceCaptureIDs); i++ {
			sourceCaptures = sourceCaptures[:len(sourceCaptures)-1] + fmt.Sprintf(",%v]", sourceCaptureIDs[i])
		}
	}

	query := `
		INSERT INTO publications (title, content, format, type, source_captures, project, target_path)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, title, content, format, type, source_captures, project, target_path, created_at
	`

	pub := &Publication{}
	err := db.QueryRow(query, title, content, format, pubType, sourceCaptures, project, targetPath).Scan(
		&pub.ID,
		&pub.Title,
		&pub.Content,
		&pub.Format,
		&pub.Type,
		&pub.SourceCaptures,
		&pub.Project,
		&pub.TargetPath,
		&pub.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert publication: %w", err)
	}

	return pub, nil
}

// ToolMessage represents a cross-tool communication message
type ToolMessage struct {
	ID          int64     `json:"id"`
	FromTool    string    `json:"from_tool"`
	ToTool      string    `json:"to_tool"`
	MessageType string    `json:"message_type"`
	Data        string    `json:"data"`
	Processed   bool      `json:"processed"`
	CreatedAt   time.Time `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at"`
}

// GetUnprocessedToolMessages retrieves unprocessed tool messages for uroboro
func (db *DB) GetUnprocessedToolMessages() ([]*ToolMessage, error) {
	query := `
		SELECT id, from_tool, to_tool, message_type, data, processed, created_at, processed_at
		FROM tool_messages 
		WHERE to_tool = 'uroboro' AND processed = FALSE
		ORDER BY created_at ASC
	`
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tool messages: %w", err)
	}
	defer rows.Close()
	
	var messages []*ToolMessage
	for rows.Next() {
		msg := &ToolMessage{}
		err := rows.Scan(
			&msg.ID,
			&msg.FromTool,
			&msg.ToTool,
			&msg.MessageType,
			&msg.Data,
			&msg.Processed,
			&msg.CreatedAt,
			&msg.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tool message: %w", err)
		}
		messages = append(messages, msg)
	}
	
	return messages, nil
}

// MarkToolMessageProcessed marks a tool message as processed
func (db *DB) MarkToolMessageProcessed(id int64) error {
	query := `
		UPDATE tool_messages 
		SET processed = TRUE, processed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	
	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to mark tool message as processed: %w", err)
	}
	
	return nil
}

// ProcessUroboroCaptures processes uroboro_capture messages from wherewasi database
func (db *DB) ProcessUroboroCaptures() error {
	// Connect to wherewasi database for tool messages
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	
	wherewasiDBPath := filepath.Join(homeDir, ".local", "share", "wherewasi", "context.sqlite")
	if _, err := os.Stat(wherewasiDBPath); os.IsNotExist(err) {
		fmt.Println("📝 No wherewasi database found - no tool messages to process")
		return nil
	}
	
	wherewasiDB, err := sql.Open("sqlite3", wherewasiDBPath)
	if err != nil {
		return fmt.Errorf("failed to open wherewasi database: %w", err)
	}
	defer wherewasiDB.Close()
	
	// Get unprocessed uroboro messages from wherewasi
	query := `
		SELECT id, from_tool, to_tool, message_type, data, processed, created_at, processed_at
		FROM tool_messages 
		WHERE to_tool = 'uroboro' AND processed = FALSE
		ORDER BY created_at ASC
	`
	
	rows, err := wherewasiDB.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query tool messages from wherewasi: %w", err)
	}
	defer rows.Close()
	
	var processedCount int
	for rows.Next() {
		var msg ToolMessage
		err := rows.Scan(
			&msg.ID,
			&msg.FromTool,
			&msg.ToTool,
			&msg.MessageType,
			&msg.Data,
			&msg.Processed,
			&msg.CreatedAt,
			&msg.ProcessedAt,
		)
		if err != nil {
			fmt.Printf("⚠️  Failed to scan tool message: %v\n", err)
			continue
		}
		
		if msg.MessageType == "uroboro_capture" {
			// Parse the JSON data to extract capture content
			var captureData map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Data), &captureData); err != nil {
				fmt.Printf("⚠️  Failed to parse uroboro capture data from %s: %v\n", msg.FromTool, err)
				continue
			}
			
			// Extract fields
			content, _ := captureData["content"].(string)
			project, _ := captureData["project"].(string)
			tags, _ := captureData["tags"].(string)
			
			if content != "" {
				// Create the capture from tool integration
				_, err := db.InsertCapture(content, project, tags)
				if err != nil {
					fmt.Printf("⚠️  Failed to create capture from %s: %v\n", msg.FromTool, err)
					continue
				}
				
				fmt.Printf("📝 Auto-captured from %s: %s\n", msg.FromTool, content)
				processedCount++
			}
			
			// Mark as processed in wherewasi database
			updateQuery := `
				UPDATE tool_messages 
				SET processed = TRUE, processed_at = CURRENT_TIMESTAMP
				WHERE id = ?
			`
			
			_, err = wherewasiDB.Exec(updateQuery, msg.ID)
			if err != nil {
				fmt.Printf("⚠️  Failed to mark message as processed: %v\n", err)
			}
		}
	}
	
	if processedCount > 0 {
		fmt.Printf("✅ Processed %d tool messages from wherewasi\n", processedCount)
	} else {
		fmt.Println("📝 No new tool messages to process")
	}
	
	return nil
}
