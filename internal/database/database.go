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
			tags TEXT,
			branch TEXT
		);
		CREATE INDEX idx_captures_timestamp ON captures(timestamp);
		CREATE INDEX idx_captures_project ON captures(project);
		`
		if _, err := db.db.Exec(schema); err != nil {
			return fmt.Errorf("create schema: %w", err)
		}
	} else {
		// Migrate: add branch column if missing
		var hasBranch bool
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
			if name == "branch" {
				hasBranch = true
			}
		}
		if !hasBranch {
			if _, err := db.db.Exec(`ALTER TABLE captures ADD COLUMN branch TEXT`); err != nil {
				return fmt.Errorf("add branch column: %w", err)
			}
		}
	}

	return nil
}

func (db *DB) InsertCapture(content, project, tags, branch string, timestamp *time.Time) (*Capture, error) {
	ts := time.Now()
	if timestamp != nil {
		ts = *timestamp
	}
	result, err := db.db.Exec(
		`INSERT INTO captures (content, project, tags, branch, timestamp) VALUES (?, ?, ?, ?, ?)`,
		content, project, tags, branch, ts,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &Capture{ID: id, Timestamp: ts, Content: content, Project: project, Tags: tags, Branch: branch}, nil
}

// QueryCaptures is a flexible query builder for captures
type CaptureQuery struct {
	Days    int
	Project string
	Branch  string
	Limit   int
	Since   *time.Time
	Keyword string // Text search in content (case-insensitive LIKE)
}

func (db *DB) QueryCaptures(q CaptureQuery) ([]Capture, error) {
	query := `SELECT id, timestamp, content, COALESCE(project, ''), COALESCE(tags, ''), COALESCE(branch, '') FROM captures WHERE 1=1`
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
	if q.Keyword != "" {
		words := strings.Fields(q.Keyword)
		for _, word := range words {
			query += ` AND content LIKE ?`
			args = append(args, "%"+word+"%")
		}
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
		if err := rows.Scan(&c.ID, &c.Timestamp, &c.Content, &c.Project, &c.Tags, &c.Branch); err != nil {
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
