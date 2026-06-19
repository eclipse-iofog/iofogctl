package assets

import "embed"

// FS holds install scripts and service unit templates bundled with the CLI.
//
//go:embed agent airgap-agent airgap-controller container-agent container-controller controller
var FS embed.FS
