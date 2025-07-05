# FEAST IMPLEMENTATION PLAN

## Overview

Add auto-feast functionality with digest to uroboro to prevent cognitive overload while preserving data. The snake eats its tail gracefully, with a final "spark test" before archiving.

## Core Problem Solved

- **"That's a lot of captures... nevermind" effect** - Too many old captures create decision paralysis
- **Cognitive load** - Historical data interferes with present-moment workflow
- **Fear of data loss** - Users hesitant to delete potentially important work records

## Solution: Feast with Digest

Auto-archive old captures with optional digest review - second exposure for "spark test" before permanent archiving.

## Database Schema Changes

### New Archive Table
```sql
CREATE TABLE archived_sessions (
    id INTEGER PRIMARY KEY,
    original_id INTEGER,
    description TEXT,
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    directory TEXT,
    tags TEXT,
    archived_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    archive_reason TEXT DEFAULT 'auto_feast'
);
```

### Migration Strategy
- Add archived_sessions table
- Keep existing sessions table structure
- Add archived_at column to sessions for tracking

## Command Implementation

### New `feast` Command
```bash
# Manual feast with digest
uro feast

# Auto-feast (runs on startup if conditions met)
uro feast --auto

# Feast with specific age threshold
uro feast --days 30

# Skip digest, archive immediately
uro feast --silent
```

### Integration with Existing Commands
```bash
# Modified status to show clean active view
uro status                    # Shows last 30 days only
uro status --all             # Shows active + archived count
uro status --archived        # Browse archived sessions
```

## User Experience Flow

### Auto-Feast Startup Flow
1. `uro status` or any command triggers auto-feast check
2. If items older than 30 days exist, show digest
3. User can rescue items or let them archive
4. Clean active workspace remains

### Digest Display
```
🐍 Auto-feast digest (23 items ready for archive):
   ✨ 3 days ago: "Fixed OAuth redirect bug in user service"
   ✨ 5 days ago: "Researched Redis caching patterns for API"  
   ✨ 1 week ago: "Meeting notes: Q4 architecture review"
   
Press 'r' to rescue items, 's' to skip digest, Enter to archive all
```

### Rescue Flow
```
Enter numbers to rescue (comma-separated): 1,3
Rescued: "Fixed OAuth redirect bug" and "Meeting notes: Q4 architecture review"
Archived: 21 items

🐍 Feasted on 21 items, 2 rescued and kept active
```

## Implementation Phases

### Phase 1: Database Foundation
- [ ] Add archived_sessions table
- [ ] Create migration script
- [ ] Add archive/restore functions
- [ ] Test data integrity

### Phase 2: Basic Feast Command
- [ ] Implement manual `uro feast` command
- [ ] Add archiving logic with age threshold
- [ ] Basic digest display
- [ ] Simple rescue functionality

### Phase 3: Auto-Feast Integration
- [ ] Add auto-feast trigger to startup
- [ ] Implement digest UI with user interaction
- [ ] Add configuration for auto-feast threshold
- [ ] Test user experience flow

### Phase 4: Enhanced Features
- [ ] Add feast statistics and reporting
- [ ] Implement archived session browsing
- [ ] Add search within archived sessions
- [ ] Configuration options for digest behavior

## Configuration Options

### Auto-Feast Settings
```yaml
# ~/.uroboro/config.yaml
feast:
  auto_enabled: true
  age_threshold_days: 30
  show_digest: true
  max_digest_items: 10
  rescue_enabled: true
```

## Technical Implementation Details

### Core Functions Needed
```go
type FeastEngine struct {
    db *sql.DB
    config *Config
}

func (f *FeastEngine) GetItemsForArchive(days int) []Session
func (f *FeastEngine) ShowDigest(items []Session) DigestResult
func (f *FeastEngine) HandleRescue(items []Session) []Session
func (f *FeastEngine) ArchiveItems(items []Session) error
func (f *FeastEngine) AutoFeastCheck() error
```

### Key Considerations
- **Performance**: Pagination for large archives
- **Safety**: Confirmation prompts for manual feast
- **Flexibility**: Configurable thresholds
- **Recovery**: Ability to restore from archive if needed

## Success Metrics

### User Experience
- [ ] "That's a lot of captures" feeling eliminated
- [ ] Status command shows manageable number of items
- [ ] Users feel safe letting items archive
- [ ] Second exposure actually helps surface important work

### Technical
- [ ] Auto-feast runs smoothly on startup
- [ ] Database performance remains good with archived data
- [ ] No data loss during archiving process
- [ ] Easy recovery of archived items when needed

## Branch Strategy

### Recommended Approach
1. Create feature branch: `feature/auto-feast-digest`
2. Implement in phases with incremental commits
3. Test thoroughly with real uroboro data
4. Get user feedback before merging to main

### Testing Strategy
- Unit tests for archiving logic
- Integration tests for auto-feast flow
- Manual testing with realistic data volumes
- Performance testing with large archives

## Future Enhancements

### Potential Features
- **Smart archiving**: ML-based relevance scoring
- **Seasonal unarchiving**: Bring back archived items by project/context
- **Archive analytics**: Insights on work patterns from archived data
- **Export options**: Extract archived data for external tools

### Core vs Extended Positioning
- **Core**: Basic auto-feast with digest
- **Extended**: Advanced archive analytics, ML features, complex recovery tools

## Implementation Priority

**HIGH PRIORITY**: Auto-feast with digest - solves immediate cognitive load problem
**MEDIUM PRIORITY**: Enhanced archive browsing and search
**LOW PRIORITY**: Advanced analytics and ML features

---

*This plan addresses the core insight: "Archive is deletion for your attention while preserving data for edge cases."*

*The snake eats its tail gracefully, with a final chance for important work to resurface.*