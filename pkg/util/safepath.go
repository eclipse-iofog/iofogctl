package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ValidateReadableFile cleans a user-supplied path and ensures it is a readable regular file.
func ValidateReadableFile(path string) (string, error) {
	clean := filepath.Clean(strings.TrimSpace(path))
	if clean == "" || clean == "." {
		return "", fmt.Errorf("invalid file path")
	}
	info, err := os.Stat(clean)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory", clean)
	}
	return clean, nil
}

// ReadUserFile reads a validated user-supplied file path (CLI -f, --ca, etc.).
func ReadUserFile(path string) ([]byte, error) {
	clean, err := ValidateReadableFile(path)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(clean) // #nosec G304 -- path validated by ValidateReadableFile
}

// CreateUserFile creates or truncates a user-supplied output path.
func CreateUserFile(path string, perm os.FileMode) (*os.File, error) {
	clean, err := validateLocalPath(path)
	if err != nil {
		return nil, err
	}
	return os.OpenFile(clean, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm) // #nosec G304 -- validated path
}

func validateLocalPath(path string) (string, error) {
	clean := filepath.Clean(strings.TrimSpace(path))
	if clean == "" || clean == "." {
		return "", fmt.Errorf("invalid file path")
	}
	for _, seg := range strings.Split(filepath.ToSlash(clean), "/") {
		if seg == ".." {
			return "", fmt.Errorf("path %q contains traversal", clean)
		}
	}
	return clean, nil
}

// ReadValidatedFile reads a path after rejecting traversal segments (cache/internal paths).
func ReadValidatedFile(path string) ([]byte, error) {
	clean, err := validateLocalPath(path)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(clean) // #nosec G304 -- validated path
}

// OpenValidatedFile opens a path after rejecting traversal segments.
func OpenValidatedFile(path string) (*os.File, error) {
	clean, err := validateLocalPath(path)
	if err != nil {
		return nil, err
	}
	return os.Open(clean) // #nosec G304 -- validated path
}

// WriteValidatedFile writes data to a validated path.
func WriteValidatedFile(path string, data []byte, perm os.FileMode) error {
	clean, err := validateLocalPath(path)
	if err != nil {
		return err
	}
	return os.WriteFile(clean, data, perm)
}

// PathUnderRoot joins elems under root and rejects traversal outside root.
func PathUnderRoot(root string, elems ...string) (string, error) {
	root = filepath.Clean(root)
	all := append([]string{root}, elems...)
	joined := filepath.Clean(filepath.Join(all...))
	if joined != root && !strings.HasPrefix(joined, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %q escapes root %q", joined, root)
	}
	return joined, nil
}

// ReadFileUnderRoot reads rel from an opened root directory.
func ReadFileUnderRoot(root, rel string) ([]byte, error) {
	f, err := OpenUnderRoot(root, rel)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// OpenUnderRoot opens rel inside root using os.OpenRoot when available.
func OpenUnderRoot(root, rel string) (*os.File, error) {
	root = filepath.Clean(root)
	rel = normalizeRelPath(rel)
	if err := rejectTraversal(rel); err != nil {
		return nil, err
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return r.Open(rel)
}

func normalizeRelPath(rel string) string {
	rel = filepath.ToSlash(filepath.Clean(rel))
	rel = strings.TrimPrefix(rel, "/")
	return rel
}

func rejectTraversal(rel string) error {
	if rel == ".." || strings.HasPrefix(rel, "../") || strings.Contains(rel, "/../") {
		return fmt.Errorf("path %q escapes root", rel)
	}
	return nil
}
