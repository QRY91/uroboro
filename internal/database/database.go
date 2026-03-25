package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/common"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

type Capture struct {
	ID        int64
	Timestamp time.Time
	Content   string
	Project   string
	Tags      string
	Branch    string
	Machine   string
}

func NewDB(dbPath string) (*DB, error) {
	if dbPath == "" {
		dataDir := common.GetDataDir()
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("create data dir: %w", err)
		}
		dbPath = filepath.Join(dataDir, "uroboro.sqlite")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	instance := &DB{db}
	if err := instance.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return instance, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

// VacuumInto creates a clean, defragmented backup of the database at destPath.
func (db *DB) VacuumInto(destPath string) error {
	_, err := db.db.Exec(`VACUUM INTO ?`, destPath)
	return err
}

// CaptureCount returns the total number of captures in the database.
func (db *DB) CaptureCount() (int64, error) {
	var count int64
	err := db.db.QueryRow(`SELECT COUNT(*) FROM captures`).Scan(&count)
	return count, err
}

func (db *DB) migrate() error {
	var exists bool
	err := db.db.QueryRow(`
		SELECT COUNT(*) > 0 FROM sqlite_master
		WHERE type='table' AND name='captures'
	`).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		schema := `
		CREATE TABLE captures (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			content TEXT NOT NULL,
			project TEXT,
			tags TEXT,
			branch TEXT,
			machine TEXT
		);
		CREATE INDEX idx_captures_timestamp ON captures(timestamp);
		CREATE INDEX idx_captures_project ON captures(project);
		`
		if _, err := db.db.Exec(schema); err != nil {
			return fmt.Errorf("create schema: %w", err)
		}
	} else {
		// Migrate: add missing columns
		var hasBranch, hasMachine bool
		rows, err := db.db.Query(`PRAGMA table_info(captures)`)
		if err != nil {
			return fmt.Errorf("check columns: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var cid int
			var name, typ string
			var notnull int
			var dflt *string
			var pk int
			if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
				return err
			}
			switch name {
			case "branch":
				hasBranch = true
			case "machine":
				hasMachine = true
			}
		}
		if !hasBranch {
			if _, err := db.db.Exec(`ALTER TABLE captures ADD COLUMN branch TEXT`); err != nil {
				return fmt.Errorf("add branch column: %w", err)
			}
		}
		if !hasMachine {
			if _, err := db.db.Exec(`ALTER TABLE captures ADD COLUMN machine TEXT`); err != nil {
				return fmt.Errorf("add machine column: %w", err)
			}
		}
	}

	return nil
}

func (db *DB) InsertCapture(content, project, tags, branch, machine string, timestamp *time.Time) (*Capture, error) {
	ts := time.Now()
	if timestamp != nil {
		ts = *timestamp
	}
	result, err := db.db.Exec(
		`INSERT INTO captures (content, project, tags, branch, machine, timestamp) VALUES (?, ?, ?, ?, ?, ?)`,
		content, project, tags, branch, machine, ts,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &Capture{ID: id, Timestamp: ts, Content: content, Project: project, Tags: tags, Branch: branch, Machine: machine}, nil
}

// ExistsCapture checks if a capture with the same content and project already exists.
// Used by import to avoid duplicates.
func (db *DB) ExistsCapture(content, project string) (bool, error) {
	var count int
	err := db.db.QueryRow(
		`SELECT COUNT(*) FROM captures WHERE content = ? AND COALESCE(project, '') = ?`,
		content, project,
	).Scan(&count)
	return count > 0, err
}

// QueryCaptures is a flexible query builder for captures
type CaptureQuery struct {
	Days    int
	Project string
	Branch  string
	Machine string
	Limit   int
	Since   *time.Time
	Until   *time.Time
	Keyword  string   // Text search in content (case-insensitive LIKE)
	Tags     []string // Match captures having ANY of these tags (LIKE-based)
	Projects []string // Match captures from ANY of these projects (IN clause)
}

func (db *DB) QueryCaptures(q CaptureQuery) ([]Capture, error) {
	query := `SELECT id, timestamp, content, COALESCE(project, ''), COALESCE(tags, ''), COALESCE(branch, ''), COALESCE(machine, '') FROM captures WHERE 1=1`
	var args []interface{}

	if q.Days > 0 {
		query += ` AND timestamp >= datetime('now', '-' || ? || ' days')`
		args = append(args, q.Days)
	}
	if q.Since != nil {
		query += ` AND timestamp >= ?`
		args = append(args, q.Since.Format("2006-01-02 15:04:05"))
	}
	if q.Project != "" {
		query += ` AND project = ?`
		args = append(args, q.Project)
	}
	if q.Branch != "" {
		query += ` AND branch = ?`
		args = append(args, q.Branch)
	}
	if q.Machine != "" {
		query += ` AND machine = ?`
		args = append(args, q.Machine)
	}
	if q.Until != nil {
		query += ` AND timestamp <= ?`
		args = append(args, q.Until.Format("2006-01-02 15:04:05"))
	}
	if q.Keyword != "" {
		words := strings.Fields(q.Keyword)
		for _, word := range words {
			query += ` AND content LIKE ?`
			args = append(args, "%"+word+"%")
		}
	}
	if len(q.Projects) > 0 {
		placeholders := make([]string, len(q.Projects))
		for i, p := range q.Projects {
			placeholders[i] = "?"
			args = append(args, p)
		}
		query += ` AND project IN (` + strings.Join(placeholders, ",") + `)`
	}
	if len(q.Tags) > 0 {
		tagClauses := make([]string, len(q.Tags))
		for i, tag := range q.Tags {
			tagClauses[i] = "tags LIKE ?"
			args = append(args, "%"+strings.TrimSpace(tag)+"%")
		}
		query += ` AND (` + strings.Join(tagClauses, " OR ") + `)`
	}

	query += ` ORDER BY timestamp DESC`

	if q.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, q.Limit)
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var captures []Capture
	for rows.Next() {
		var c Capture
		if err := rows.Scan(&c.ID, &c.Timestamp, &c.Content, &c.Project, &c.Tags, &c.Branch, &c.Machine); err != nil {
			return nil, err
		}
		captures = append(captures, c)
	}
	return captures, rows.Err()
}

// Convenience methods using QueryCaptures
func (db *DB) GetRecentCaptures(days int, project string) ([]Capture, error) {
	return db.QueryCaptures(CaptureQuery{Days: days, Project: project})
}

func (db *DB) GetAllCaptures() ([]Capture, error) {
	return db.QueryCaptures(CaptureQuery{})
}
