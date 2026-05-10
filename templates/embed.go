package templates

import "embed"

//go:embed agents/*.md
var AgentFS embed.FS
