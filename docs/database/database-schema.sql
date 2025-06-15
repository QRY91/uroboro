-- uroboro SQLite Schema Design
-- Supporting cross-tool communication and data analysis

-- Core captures table - replaces daily markdown files
CREATE TABLE captures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    content TEXT NOT NULL,
    project TEXT,
    tags TEXT,
    source_tool TEXT DEFAULT 'uroboro',
    metadata JSON,  -- Flexible field for tool-specific data
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Published content - tracks what we've generated
CREATE TABLE publications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    format TEXT NOT NULL, -- 'markdown', 'html', 'text'
    type TEXT NOT NULL,    -- 'blog', 'devlog', 'social', 'patch-notes'
    source_captures TEXT,  -- JSON array of capture IDs used
    project TEXT,
    target_path TEXT,      -- Where it was saved/published
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Projects - for better organization
CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    git_path TEXT,         -- Path to git repo if applicable
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity DATETIME
);

-- Cross-tool communication - for uroboro <-> doggowoof
CREATE TABLE tool_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_tool TEXT NOT NULL,
    to_tool TEXT NOT NULL,
    message_type TEXT NOT NULL,  -- 'issue_detected', 'fix_documented', 'pattern_found'
    data JSON NOT NULL,          -- Tool-specific payload
    processed BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at DATETIME
);

-- Analytics/usage tracking (local only)
CREATE TABLE usage_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command TEXT NOT NULL,
    project TEXT,
    duration_ms INTEGER,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
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

-- Example queries for QRY's analysis needs:

-- Find all React work from Q4
-- SELECT * FROM captures WHERE project LIKE '%react%' AND timestamp >= '2024-10-01';

-- Get patch notes generation history
-- SELECT * FROM publications WHERE type = 'patch-notes' ORDER BY created_at DESC;

-- Cross-tool communication: doggowoof reports issue, uroboro documents fix
-- INSERT INTO tool_messages (from_tool, to_tool, message_type, data) 
-- VALUES ('doggowoof', 'uroboro', 'issue_detected', '{"file": "auth.go", "issue": "potential memory leak", "severity": "medium"}');

-- Token cost analysis - find captures that need reorganization
-- SELECT project, COUNT(*) as capture_count, GROUP_CONCAT(tags) as all_tags 
-- FROM captures GROUP BY project HAVING capture_count > 10; 