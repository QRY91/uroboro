package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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
			tags TEXT
		);
		CREATE INDEX idx_captures_timestamp ON captures(timestamp);
		CREATE INDEX idx_captures_project ON captures(project);
		`
		if _, err := db.db.Exec(schema); err != nil {
			return fmt.Errorf("create schema: %w", err)
		}
	}

	return nil
}

func (db *DB) InsertCapture(content, project, tags string) (*Capture, error) {
	ts := time.Now()
	result, err := db.db.Exec(
		`INSERT INTO captures (content, project, tags, timestamp) VALUES (?, ?, ?, ?)`,
		content, project, tags, ts,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &Capture{ID: id, Timestamp: ts, Content: content, Project: project, Tags: tags}, nil
}

// QueryCaptures is a flexible query builder for captures
type CaptureQuery struct {
	Days    int
	Project string
	Limit   int
	Since   *time.Time
}

func (db *DB) QueryCaptures(q CaptureQuery) ([]Capture, error) {
	query := `SELECT id, timestamp, content, COALESCE(project, ''), COALESCE(tags, '') FROM captures WHERE 1=1`
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
		if err := rows.Scan(&c.ID, &c.Timestamp, &c.Content, &c.Project, &c.Tags); err != nil {
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
