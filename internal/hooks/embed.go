package hooks

import _ "embed"

//go:embed session-start.sh
var SessionStartScript []byte

//go:embed session-audit.sh
var SessionAuditScript []byte

//go:embed pre-compact.sh
var PreCompactScript []byte

//go:embed post-tool-nudge.sh
var PostToolNudgeScript []byte
