package agentpack

import "embed"

// Files is the canonical instruction pack shipped with the executable.
//
//go:embed all:.agents
var Files embed.FS
