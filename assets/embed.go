package assets

import "embed"

// FS holds install scripts and service unit templates bundled with the CLI.
//
//go:embed controller container-controller airgap-controller edgelet
var FS embed.FS
