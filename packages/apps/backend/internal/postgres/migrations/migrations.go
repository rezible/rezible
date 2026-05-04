package migrations

import "embed"

//go:embed *
var FS embed.FS
var EmbedFSDir = "."
