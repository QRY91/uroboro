# Database Management Guide

## 🗃️ Database Strategy

uroboro uses SQLite for local data storage with a clear separation between **code** and **data**.

### What Gets Committed to Git

✅ **Schema files** (`docs/database-schema.sql`)  
✅ **Migration code** (`internal/database/database.go`)  
✅ **Database tests** (`internal/database/database_test.go`)  

### What NEVER Gets Committed

❌ **Database files** (`*.sqlite`, `*.db`)  
❌ **User captures** (personal development data)  
❌ **Test databases** (temporary files)  

## 📁 File Locations

### Production Database
```bash
# Default location (cross-platform)
Linux/macOS: ~/.local/share/uroboro/uroboro.sqlite
Windows:     %APPDATA%\uroboro\uroboro.sqlite

# Custom location
uro capture --db /path/to/custom.sqlite "content"
```

### Test Databases
```bash
# Tests use temporary files
/tmp/test_*.sqlite  # Automatically cleaned up
```

## 🚀 Deployment & CI/CD

### GitHub Actions / CI
```yaml
# Tests handle database creation automatically
- name: Run tests
  run: go test ./internal/...
  # ✅ Tests create temporary databases
  # ✅ Tests clean up after themselves
  # ✅ No persistent database needed
```

### Local Development
```bash
# Fresh start - database created automatically
uro capture --db ~/.local/share/uroboro/uroboro.sqlite "first capture"
# ✅ Database and schema created on first use
# ✅ Migrations run automatically
```

### Production Deployment
- Each user has their own local database
- No central database server needed
- No database setup required during installation

## 🔄 Schema Migrations

### Current System
- **Version 1**: Initial schema with all tables
- **Auto-migration**: Runs on database open
- **Migration tracking**: `schema_migrations` table

### Adding New Migrations
1. **Update schema** in `docs/database-schema.sql`
2. **Add migration code** in `database.go`
3. **Increment version number**
4. **Add tests** for new features

Example future migration:
```go
// In migrate() function
if latestVersion < 2 {
    err := db.migrateToVersion2()
    // Add new columns, tables, etc.
}
```

## 💾 Backup & Data Management

### User Data Export (Future Feature)
```bash
# Export captures to JSON
uro export --db ~/.local/share/uroboro/uroboro.sqlite --format json

# Import from another database
uro import --db ~/.local/share/uroboro/uroboro.sqlite --from backup.json
```

### Manual Backup
```bash
# Simple file copy
cp ~/.local/share/uroboro/uroboro.sqlite ~/backup-$(date +%Y%m%d).sqlite

# Query specific data
sqlite3 ~/.local/share/uroboro/uroboro.sqlite "SELECT * FROM captures WHERE project='uroboro';"
```

## 🔧 Database Debugging

### Check Database Status
```bash
# Verify database exists and has data
sqlite3 ~/.local/share/uroboro/uroboro.sqlite ".tables"
sqlite3 ~/.local/share/uroboro/uroboro.sqlite "SELECT COUNT(*) FROM captures;"
```

### Common Issues
- **Database locked**: Another uroboro process running
- **Permission denied**: Check directory permissions
- **Schema mismatch**: Delete database, let it recreate

## 🐕 Cross-Tool Communication

### For doggowoof Integration
```sql
-- doggowoof writes issues
INSERT INTO tool_messages (from_tool, to_tool, message_type, data)
VALUES ('doggowoof', 'uroboro', 'issue_detected', '{"file": "main.go", "issue": "potential bug"}');

-- uroboro reads and processes
SELECT * FROM tool_messages WHERE to_tool = 'uroboro' AND processed = FALSE;
```

### Database Sharing
```bash
# Both tools use same database
uro capture --db ~/shared.sqlite "fixed auth issue"
doggowoof scan --db ~/shared.sqlite ./src/
```

## 🛡️ Security Considerations

- **Local only**: Database never leaves user's machine
- **No credentials**: SQLite requires no authentication
- **File permissions**: Database inherits directory permissions
- **Encryption**: Consider file-level encryption for sensitive projects

## 📊 Performance Notes

- **SQLite limits**: 281TB max database size (plenty for captures!)
- **Concurrent access**: SQLite handles multiple readers, single writer
- **Indexes**: Already optimized for common queries
- **Vacuum**: SQLite auto-manages space (no maintenance needed)

---

**Key Principle**: Database files are **user data**, not **application code**. Keep them separate in version control and deployments. 