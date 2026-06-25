package assets

import "embed"

// FS holds install scripts bundled with the CLI.
//
//go:embed edgelet
var FS embed.FS
