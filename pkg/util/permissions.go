package util

const (
	// DirPerm is the default mode for CLI config and cache directories.
	DirPerm = 0o700
	// FilePerm is the default mode for CLI config and cache files.
	FilePerm = 0o600
	// ExecPerm is for downloaded binaries and materialized shell scripts that must run.
	ExecPerm = 0o755
)
