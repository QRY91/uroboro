package hooks

import _ "embed"

//go:embed session-start.sh
var SessionStartScript []byte

//go:embed session-audit.sh
var SessionAuditScript []byte
