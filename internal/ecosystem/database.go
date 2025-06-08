package ecosystem

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// EcosystemDB wraps database operations for QRY ecosystem intelligence
type EcosystemDB struct {
	*sql.DB
	dbPath   string
	isShared bool
}

// DatabaseConfig holds configuration for ecosystem database discovery
type DatabaseConfig struct {
	ToolName     string
	FallbackPath string
	ForceLocal   bool
}

// SharedDatabasePath returns the standard ecosystem database path
func SharedDatabasePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	
	dataDir := filepath.Join(homeDir, ".local", "share", "qry")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create ecosystem data directory: %w", err)
	}
	
	return filepath.Join(dataDir, "ecosystem.sqlite"), nil
}

// NewEcosystemDB creates a new ecosystem database connection with discovery logic
func NewEcosystemDB(config DatabaseConfig) (*EcosystemDB, error) {
	var dbPath string
	var isShared bool
	
	if !config.ForceLocal {
		// Try to connect to shared ecosystem database
		sharedPath, err := SharedDatabasePath()
		if err == nil {
			// Check if shared database exists or can be created
			if _, err := os.Stat(sharedPath); err == nil || !os.IsNotExist(err) {
				dbPath = sharedPath
				isShared = true
			}
		}
	}
	
	// Fall back to tool-specific database
	if dbPath == "" {
		dbPath = config.FallbackPath
		isShared = false
	}
	
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}
	
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	// Enable optimizations
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}
	
	edb := &EcosystemDB{
		DB:       db,
		dbPath:   dbPath,
		isShared: isShared,
	}
	
	// Run migrations
	if err := edb.migrate(config.ToolName); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	
	return edb, nil
}

// IsShared returns true if this is a shared ecosystem database
func (edb *EcosystemDB) IsShared() bool {
	return edb.isShared
}

// DatabasePath returns the path to the database file
func (edb *EcosystemDB) DatabasePath() string {
	return edb.dbPath
}

// migrate runs database migrations for ecosystem tables
func (edb *EcosystemDB) migrate(toolName string) error {
	// Check if migrations table exists
	var tableExists bool
	err := edb.QueryRow(`
		SELECT COUNT(*) > 0 FROM sqlite_master 
		WHERE type='table' AND name='schema_migrations'
	`).Scan(&tableExists)
	
	if err != nil {
		return fmt.Errorf("failed to check migrations table: %w", err)
	}
	
	if !tableExists {
		// Create initial ecosystem schema
		if err := edb.createEcosystemSchema(); err != nil {
			return fmt.Errorf("failed to create ecosystem schema: %w", err)
		}
	}
	
	// Run tool-specific migrations
	switch toolName {
	case "wherewasi":
		if err := edb.migrateWherewasiTables(); err != nil {
			return fmt.Errorf("failed to migrate wherewasi tables: %w", err)
		}
	case "uroboro":
		if err := edb.migrateUroboroTables(); err != nil {
			return fmt.Errorf("failed to migrate uroboro tables: %w", err)
		}
	case "examinator":
		if err := edb.migrateExaminatorTables(); err != nil {
			return fmt.Errorf("failed to migrate examinator tables: %w", err)
		}
	}
	
	return nil
}

// createEcosystemSchema creates the shared ecosystem database schema
func (edb *EcosystemDB) createEcosystemSchema() error {
	schema := `
	-- QRY Ecosystem Shared Database Schema
	-- Each tool can function independently, but when pointed to shared DB,
	-- they gain ecosystem intelligence through overlapping tables
	
	-- === SHARED CORE TABLES ===
	
	-- Projects tracked across the ecosystem
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		description TEXT,
		path TEXT,
		git_repo BOOLEAN DEFAULT FALSE,
		last_activity DATETIME,
		primary_tool TEXT, -- which tool primarily manages this project
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Cross-tool communication
	CREATE TABLE IF NOT EXISTS tool_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_tool TEXT NOT NULL,
		to_tool TEXT NOT NULL,
		message_type TEXT NOT NULL,
		data TEXT NOT NULL,
		processed BOOLEAN DEFAULT FALSE,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		processed_at DATETIME
	);
	
	-- Usage analytics across ecosystem
	CREATE TABLE IF NOT EXISTS usage_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tool TEXT NOT NULL,
		command TEXT NOT NULL,
		project TEXT,
		duration_ms INTEGER,
		success BOOLEAN DEFAULT TRUE,
		error_message TEXT,
		session_id TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Cross-tool insights and recommendations
	CREATE TABLE IF NOT EXISTS ecosystem_insights (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		insight_type TEXT NOT NULL,
		source_tool TEXT NOT NULL,
		target_tool TEXT,
		project TEXT,
		confidence REAL DEFAULT 0.0,
		data TEXT NOT NULL, -- JSON
		applied BOOLEAN DEFAULT FALSE,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		applied_at DATETIME
	);
	
	-- === SHARED INDEXES ===
	CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);
	CREATE INDEX IF NOT EXISTS idx_projects_last_activity ON projects(last_activity);
	CREATE INDEX IF NOT EXISTS idx_tool_messages_to_tool ON tool_messages(to_tool, processed);
	CREATE INDEX IF NOT EXISTS idx_tool_messages_from_tool ON tool_messages(from_tool);
	CREATE INDEX IF NOT EXISTS idx_usage_stats_tool ON usage_stats(tool);
	CREATE INDEX IF NOT EXISTS idx_usage_stats_project ON usage_stats(project);
	CREATE INDEX IF NOT EXISTS idx_ecosystem_insights_target ON ecosystem_insights(target_tool, applied);
	CREATE INDEX IF NOT EXISTS idx_ecosystem_insights_project ON ecosystem_insights(project);
	
	-- Migration tracking
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		tool TEXT NOT NULL,
		description TEXT NOT NULL,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Initial migration record
	INSERT OR IGNORE INTO schema_migrations (version, tool, description) 
	VALUES (1, 'ecosystem', 'Initial shared ecosystem schema');
	`
	
	_, err := edb.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create ecosystem schema: %w", err)
	}
	
	return nil
}

// migrateWherewasiTables creates wherewasi-specific tables
func (edb *EcosystemDB) migrateWherewasiTables() error {
	schema := `
	-- === WHEREWASI TABLES ===
	
	-- Context sessions for ripcord deployments
	CREATE TABLE IF NOT EXISTS context_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project TEXT NOT NULL,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		context_data TEXT NOT NULL,
		session_info TEXT,
		keywords TEXT,
		git_branch TEXT,
		git_commit TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Wherewasi-specific indexes
	CREATE INDEX IF NOT EXISTS idx_context_sessions_project ON context_sessions(project);
	CREATE INDEX IF NOT EXISTS idx_context_sessions_timestamp ON context_sessions(timestamp);
	CREATE INDEX IF NOT EXISTS idx_context_sessions_keywords ON context_sessions(keywords);
	CREATE INDEX IF NOT EXISTS idx_context_sessions_git_branch ON context_sessions(git_branch);
	
	-- Migration record
	INSERT OR IGNORE INTO schema_migrations (version, tool, description) 
	VALUES (2, 'wherewasi', 'Wherewasi context sessions and project tracking');
	`
	
	_, err := edb.Exec(schema)
	return err
}

// migrateUroboroTables creates uroboro-specific tables
func (edb *EcosystemDB) migrateUroboroTables() error {
	schema := `
	-- === UROBORO TABLES ===
	
	-- Content captures
	CREATE TABLE IF NOT EXISTS captures (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		content TEXT NOT NULL,
		project TEXT,
		tags TEXT,
		source_tool TEXT DEFAULT 'uroboro',
		metadata TEXT, -- JSON
		context_session_id INTEGER,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (context_session_id) REFERENCES context_sessions(id)
	);
	
	-- Published content
	CREATE TABLE IF NOT EXISTS publications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		format TEXT NOT NULL,
		type TEXT NOT NULL,
		source_captures TEXT, -- JSON array of capture IDs
		project TEXT,
		target_path TEXT,
		context_session_id INTEGER,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (context_session_id) REFERENCES context_sessions(id)
	);
	
	-- Uroboro-specific indexes
	CREATE INDEX IF NOT EXISTS idx_captures_timestamp ON captures(timestamp);
	CREATE INDEX IF NOT EXISTS idx_captures_project ON captures(project);
	CREATE INDEX IF NOT EXISTS idx_captures_source_tool ON captures(source_tool);
	CREATE INDEX IF NOT EXISTS idx_captures_context_session ON captures(context_session_id);
	CREATE INDEX IF NOT EXISTS idx_publications_type ON publications(type);
	CREATE INDEX IF NOT EXISTS idx_publications_project ON publications(project);
	CREATE INDEX IF NOT EXISTS idx_publications_context_session ON publications(context_session_id);
	
	-- Migration record
	INSERT OR IGNORE INTO schema_migrations (version, tool, description) 
	VALUES (3, 'uroboro', 'Uroboro captures and publications with context linking');
	`
	
	_, err := edb.Exec(schema)
	return err
}

// migrateExaminatorTables creates examinator-specific tables
func (edb *EcosystemDB) migrateExaminatorTables() error {
	schema := `
	-- === EXAMINATOR TABLES ===
	
	-- Flashcards for spaced repetition
	CREATE TABLE IF NOT EXISTS flashcards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		question TEXT NOT NULL,
		answer TEXT NOT NULL,
		category TEXT,
		difficulty INTEGER DEFAULT 1,
		source_capture_id INTEGER,
		context_session_id INTEGER,
		project TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_reviewed DATETIME,
		next_review DATETIME,
		ease_factor REAL DEFAULT 2.5,
		review_count INTEGER DEFAULT 0,
		correct_streak INTEGER DEFAULT 0,
		FOREIGN KEY (source_capture_id) REFERENCES captures(id),
		FOREIGN KEY (context_session_id) REFERENCES context_sessions(id)
	);
	
	-- Study sessions
	CREATE TABLE IF NOT EXISTS study_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project TEXT,
		flashcards_reviewed INTEGER DEFAULT 0,
		correct_answers INTEGER DEFAULT 0,
		duration_minutes INTEGER,
		context_session_id INTEGER,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (context_session_id) REFERENCES context_sessions(id)
	);
	
	-- Flashcard reviews (detailed tracking)
	CREATE TABLE IF NOT EXISTS flashcard_reviews (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		flashcard_id INTEGER NOT NULL,
		study_session_id INTEGER,
		response_quality INTEGER, -- 0-5 scale
		response_time_ms INTEGER,
		was_correct BOOLEAN,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (flashcard_id) REFERENCES flashcards(id),
		FOREIGN KEY (study_session_id) REFERENCES study_sessions(id)
	);
	
	-- Examinator-specific indexes
	CREATE INDEX IF NOT EXISTS idx_flashcards_project ON flashcards(project);
	CREATE INDEX IF NOT EXISTS idx_flashcards_next_review ON flashcards(next_review);
	CREATE INDEX IF NOT EXISTS idx_flashcards_source_capture ON flashcards(source_capture_id);
	CREATE INDEX IF NOT EXISTS idx_flashcards_context_session ON flashcards(context_session_id);
	CREATE INDEX IF NOT EXISTS idx_study_sessions_project ON study_sessions(project);
	CREATE INDEX IF NOT EXISTS idx_flashcard_reviews_flashcard ON flashcard_reviews(flashcard_id);
	CREATE INDEX IF NOT EXISTS idx_flashcard_reviews_session ON flashcard_reviews(study_session_id);
	
	-- Migration record
	INSERT OR IGNORE INTO schema_migrations (version, tool, description) 
	VALUES (4, 'examinator', 'Examinator flashcards and study tracking with ecosystem links');
	`
	
	_, err := edb.Exec(schema)
	return err
}

// Cross-tool communication methods

// SendToolMessage sends a message to another ecosystem tool
func (edb *EcosystemDB) SendToolMessage(fromTool, toTool, messageType, data string) error {
	query := `
		INSERT INTO tool_messages (from_tool, to_tool, message_type, data)
		VALUES (?, ?, ?, ?)
	`
	
	_, err := edb.Exec(query, fromTool, toTool, messageType, data)
	if err != nil {
		return fmt.Errorf("failed to send tool message: %w", err)
	}
	
	return nil
}

// GetUnprocessedMessages retrieves unprocessed tool messages for a specific tool
func (edb *EcosystemDB) GetUnprocessedMessages(toolName string) ([]*ToolMessage, error) {
	query := `
		SELECT id, from_tool, to_tool, message_type, data, processed, created_at, processed_at
		FROM tool_messages 
		WHERE to_tool = ? AND processed = FALSE
		ORDER BY created_at ASC
	`
	
	rows, err := edb.Query(query, toolName)
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

// MarkMessageProcessed marks a tool message as processed
func (edb *EcosystemDB) MarkMessageProcessed(id int64) error {
	query := `
		UPDATE tool_messages 
		SET processed = TRUE, processed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	
	_, err := edb.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to mark tool message as processed: %w", err)
	}
	
	return nil
}

// Project management methods

// TrackProject records project activity in the ecosystem
func (edb *EcosystemDB) TrackProject(name, path, primaryTool string, isGitRepo bool) error {
	query := `
		INSERT INTO projects (name, path, primary_tool, git_repo, last_activity, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(name) DO UPDATE SET
			last_activity = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP,
			path = excluded.path,
			primary_tool = excluded.primary_tool,
			git_repo = excluded.git_repo
	`
	
	_, err := edb.Exec(query, name, path, primaryTool, isGitRepo)
	if err != nil {
		return fmt.Errorf("failed to track project: %w", err)
	}
	
	return nil
}

// GetRecentProjects returns recently active projects
func (edb *EcosystemDB) GetRecentProjects(limit int) ([]*Project, error) {
	query := `
		SELECT id, name, description, path, git_repo, last_activity, primary_tool, created_at, updated_at
		FROM projects 
		ORDER BY last_activity DESC 
		LIMIT ?
	`
	
	rows, err := edb.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent projects: %w", err)
	}
	defer rows.Close()
	
	var projects []*Project
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.Path,
			&project.GitRepo,
			&project.LastActivity,
			&project.PrimaryTool,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}
	
	return projects, nil
}

// Usage tracking

// TrackUsage records tool usage for analytics
func (edb *EcosystemDB) TrackUsage(tool, command, project, sessionID string, durationMs int, success bool, errorMsg string) error {
	query := `
		INSERT INTO usage_stats (tool, command, project, session_id, duration_ms, success, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := edb.Exec(query, tool, command, project, sessionID, durationMs, success, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to track usage: %w", err)
	}
	
	return nil
}

// Wherewasi-specific methods

// SaveContext saves a context session to the database
func (edb *EcosystemDB) SaveContext(project, contextData, sessionInfo, keywords string) (*ContextSession, error) {
	timestamp := time.Now()

	query := `
		INSERT INTO context_sessions (project, context_data, session_info, keywords, timestamp)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := edb.Exec(query, project, contextData, sessionInfo, keywords, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to save context: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get insert ID: %w", err)
	}

	session := &ContextSession{
		ID:          id,
		Project:     project,
		Timestamp:   timestamp,
		ContextData: contextData,
		SessionInfo: sessionInfo,
		Keywords:    keywords,
		CreatedAt:   timestamp,
	}

	return session, nil
}

// GetRecentContexts retrieves recent context sessions for a project
func (edb *EcosystemDB) GetRecentContexts(project string, limit int) ([]ContextSession, error) {
	query := `
		SELECT id, project, timestamp, context_data, session_info, keywords, git_branch, git_commit, created_at
		FROM context_sessions 
		WHERE project = ?
		ORDER BY timestamp DESC 
		LIMIT ?
	`

	rows, err := edb.Query(query, project, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent contexts: %w", err)
	}
	defer rows.Close()

	var sessions []ContextSession
	for rows.Next() {
		var session ContextSession
		err := rows.Scan(
			&session.ID,
			&session.Project,
			&session.Timestamp,
			&session.ContextData,
			&session.SessionInfo,
			&session.Keywords,
			&session.GitBranch,
			&session.GitCommit,
			&session.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan context session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Uroboro-specific methods

// InsertCapture creates a new capture in the database
func (edb *EcosystemDB) InsertCapture(content, project, tags string) (*Capture, error) {
	timestamp := time.Now()

	query := `
		INSERT INTO captures (content, project, tags, timestamp)
		VALUES (?, ?, ?, ?)
	`

	result, err := edb.Exec(query, content, project, tags, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to insert capture: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get insert ID: %w", err)
	}

	capture := &Capture{
		ID:        id,
		Timestamp: timestamp,
		Content:   content,
		Project:   &project,
		Tags:      &tags,
		SourceTool: "uroboro",
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}

	return capture, nil
}

// InsertCaptureWithContext creates a new capture linked to a context session
func (edb *EcosystemDB) InsertCaptureWithContext(content, project, tags string, contextSessionID *int64) (*Capture, error) {
	timestamp := time.Now()

	query := `
		INSERT INTO captures (content, project, tags, timestamp, context_session_id)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := edb.Exec(query, content, project, tags, timestamp, contextSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert capture with context: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get insert ID: %w", err)
	}

	capture := &Capture{
		ID:               id,
		Timestamp:        timestamp,
		Content:          content,
		Project:          &project,
		Tags:             &tags,
		SourceTool:       "uroboro",
		ContextSessionID: contextSessionID,
		CreatedAt:        timestamp,
		UpdatedAt:        timestamp,
	}

	return capture, nil
}

// GetRecentCaptures retrieves recent captures for publishing
func (edb *EcosystemDB) GetRecentCaptures(days int, project string) ([]Capture, error) {
	query := `
		SELECT id, timestamp, content, project, tags, source_tool, metadata, context_session_id, created_at, updated_at
		FROM captures 
		WHERE timestamp >= datetime('now', '-' || ? || ' days')
	`
	args := []interface{}{days}

	if project != "" {
		query += " AND project = ?"
		args = append(args, project)
	}

	query += " ORDER BY timestamp DESC"

	rows, err := edb.Query(query, args...)
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
			&capture.ContextSessionID,
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

// InsertPublication creates a publication record
func (edb *EcosystemDB) InsertPublication(title, content, format, pubType, project, targetPath string, sourceCaptureIDs []int64) (*Publication, error) {
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
	err := edb.QueryRow(query, title, content, format, pubType, sourceCaptures, project, targetPath).Scan(
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

// ProcessUroboroCaptures processes messages from other ecosystem tools
func (edb *EcosystemDB) ProcessUroboroCaptures() error {
	if !edb.IsShared() {
		fmt.Println("📝 Local database mode - no ecosystem messages to process")
		return nil
	}

	// Get unprocessed messages for uroboro
	messages, err := edb.GetUnprocessedMessages(ToolUroboro)
	if err != nil {
		return fmt.Errorf("failed to get unprocessed messages: %w", err)
	}

	var processedCount int
	for _, msg := range messages {
		switch msg.MessageType {
		case MessageTypeContextUpdate:
			// Process context update from wherewasi
			var contextData ContextUpdateMessageData
			if err := msg.ParseMessageData(&contextData); err != nil {
				fmt.Printf("⚠️  Failed to parse context update: %v\n", err)
				continue
			}

			// Create automatic capture from context update
			content := fmt.Sprintf("Context update: %s", contextData.SessionInfo)
			if contextData.Keywords != "" {
				content += fmt.Sprintf(" [%s]", contextData.Keywords)
			}

			_, err := edb.InsertCapture(content, contextData.Project, "auto-context")
			if err != nil {
				fmt.Printf("⚠️  Failed to create capture from context: %v\n", err)
				continue
			}

			fmt.Printf("📝 Auto-captured from wherewasi context: %s\n", contextData.SessionInfo)
			processedCount++
		}

		// Mark message as processed
		if err := edb.MarkMessageProcessed(msg.ID); err != nil {
			fmt.Printf("⚠️  Failed to mark message as processed: %v\n", err)
		}
	}

	if processedCount > 0 {
		fmt.Printf("✅ Processed %d ecosystem messages\n", processedCount)
	} else {
		fmt.Println("📝 No new ecosystem messages to process")
	}

	return nil
}

// SearchStoredContexts searches stored contexts by keyword
func (edb *EcosystemDB) SearchStoredContexts(keyword string) ([]ContextSession, error) {
	query := `
		SELECT id, project, timestamp, context_data, session_info, keywords, git_branch, git_commit, created_at
		FROM context_sessions 
		WHERE context_data LIKE ? OR keywords LIKE ? OR session_info LIKE ?
		ORDER BY timestamp DESC 
		LIMIT 10
	`

	searchTerm := "%" + keyword + "%"
	rows, err := edb.Query(query, searchTerm, searchTerm, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search contexts: %w", err)
	}
	defer rows.Close()

	var sessions []ContextSession
	for rows.Next() {
		var session ContextSession
		err := rows.Scan(
			&session.ID,
			&session.Project,
			&session.Timestamp,
			&session.ContextData,
			&session.SessionInfo,
			&session.Keywords,
			&session.GitBranch,
			&session.GitCommit,
			&session.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan context session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

