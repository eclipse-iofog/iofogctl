package wasm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const (
	metadataFilename      = "metadata.json"
	airgapBinariesDirname = "airgap-binaries"
	defaultConfigBasename = ".iofog/v3"
)

var cacheRootOverride string

// SetCacheRootForTest overrides the WASM cache root (tests only).
func SetCacheRootForTest(dir string) {
	cacheRootOverride = dir
}

// ResetCacheRootForTest clears the WASM cache root override.
func ResetCacheRootForTest() {
	cacheRootOverride = ""
}

func wasmCacheRoot() string {
	if cacheRootOverride != "" {
		return cacheRootOverride
	}
	if root := os.Getenv("IOFOG_CONFIG_DIR"); root != "" {
		return root
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join("/tmp", defaultConfigBasename)
	}
	return filepath.Join(home, defaultConfigBasename)
}

type cacheMetadata struct {
	Handler        string    `json:"handler"`
	OS             string    `json:"os"`
	Arch           string    `json:"arch"`
	SourceURL      string    `json:"sourceUrl,omitempty"`
	SourcePath     string    `json:"sourcePath,omitempty"`
	SourceChecksum string    `json:"sourceChecksum,omitempty"`
	CanonicalName  string    `json:"canonicalName"`
	BinaryChecksum string    `json:"binaryChecksum,omitempty"`
	BinarySize     int64     `json:"binarySize,omitempty"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func cacheDir(namespace, handler, osName, archName string) string {
	platform := osName + "-" + archName
	pathElems := []string{wasmCacheRoot(), airgapBinariesDirname}
	if namespace != "" {
		pathElems = append(pathElems, namespace)
	}
	pathElems = append(pathElems, "wasm", handler, platform)
	return filepath.Join(pathElems...)
}

func sourceCachePath(namespace, handler, osName, archName string) string {
	return filepath.Join(cacheDir(namespace, handler, osName, archName), "source")
}

func binaryCachePath(namespace, handler, osName, archName, canonicalName string) string {
	return filepath.Join(cacheDir(namespace, handler, osName, archName), "bin", canonicalName)
}

func metadataPath(namespace, handler, osName, archName string) string {
	return filepath.Join(cacheDir(namespace, handler, osName, archName), metadataFilename)
}

func loadCacheMetadata(path string) (*cacheMetadata, error) {
	data, err := util.ReadValidatedFile(path)
	if err != nil {
		return nil, err
	}
	var meta cacheMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func saveCacheMetadata(path string, meta cacheMetadata) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return util.WriteValidatedFile(path, data, util.FilePerm)
}

func canReuseExtractedBinary(binPath string, meta *cacheMetadata, handler, osName, archName, canonicalName, sourceChecksum string) (bool, string) {
	if meta == nil {
		return false, ""
	}
	if meta.Handler != handler || meta.OS != osName || meta.Arch != archName || meta.CanonicalName != canonicalName {
		return false, ""
	}
	if meta.SourceChecksum != sourceChecksum {
		return false, fmt.Sprintf("WASM source for %s changed; refreshing cache", handler)
	}
	if meta.BinaryChecksum == "" {
		return false, "WASM cache is missing binary checksum metadata; refreshing cache"
	}
	info, err := os.Stat(binPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Sprintf("Cached WASM binary for %s is missing on disk; refreshing cache", handler)
		}
		return false, fmt.Sprintf("Failed to stat cached WASM binary %s: %v", binPath, err)
	}
	checksum, size, err := fileSHA256(binPath)
	if err != nil {
		return false, fmt.Sprintf("Failed to verify cached WASM binary: %v", err)
	}
	if checksum != meta.BinaryChecksum {
		return false, "Cached WASM binary checksum mismatch; refreshing cache"
	}
	if size != meta.BinarySize || info.Size() != meta.BinarySize {
		return false, "Cached WASM binary size mismatch; refreshing cache"
	}
	return true, ""
}

func ensureCacheDir(namespace, handler, osName, archName string) error {
	return os.MkdirAll(filepath.Join(cacheDir(namespace, handler, osName, archName), "bin"), util.DirPerm)
}

// CacheDir returns the cache directory for a handler/platform (mirrors config.GetAirgapWasmCacheDir).
func CacheDir(namespace, handler, osName, archName string) string {
	return cacheDir(namespace, handler, osName, archName)
}
