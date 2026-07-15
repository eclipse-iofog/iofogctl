package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const defaultInstallDir = "/usr/local/bin"

var localInstallDir = defaultInstallDir

var httpClient = http.DefaultClient

// StagedBinary is a resolved WASM shim ready for local install or remote SCP.
type StagedBinary struct {
	Handler       string
	CanonicalName string
	LocalPath     string
	RemotePath    string
	SHA256        string
	Changed       bool
}

// ResolveWasmArtifacts resolves configured handlers to staged shim binaries on the operator machine.
func ResolveWasmArtifacts(ctx context.Context, namespace, platform string, wasm map[string]Pack, airgap bool) ([]StagedBinary, error) {
	if len(wasm) == 0 {
		return nil, nil
	}

	osName, archName, err := platformToOSArch(platform)
	if err != nil {
		return nil, err
	}

	handlers := make([]string, 0, len(wasm))
	for handler := range wasm {
		handlers = append(handlers, handler)
	}
	sort.Strings(handlers)

	staged := make([]StagedBinary, 0, len(handlers))
	for _, handler := range handlers {
		entry, err := resolveHandler(ctx, namespace, handler, wasm[handler], osName, archName, airgap)
		if err != nil {
			return nil, err
		}
		staged = append(staged, entry)
	}
	return staged, nil
}

func resolveHandler(ctx context.Context, namespace, handler string, pack Pack, osName, archName string, airgap bool) (StagedBinary, error) {
	canonicalName := CanonicalName(handler)
	if canonicalName == "" {
		return StagedBinary{}, fmt.Errorf("unknown WASM handler %q", handler)
	}

	if err := ensureCacheDir(namespace, handler, osName, archName); err != nil {
		return StagedBinary{}, err
	}

	sourcePath, sourceMeta, err := ensureSource(ctx, namespace, handler, pack, osName, archName, airgap)
	if err != nil {
		return StagedBinary{}, err
	}

	metaPath := metadataPath(namespace, handler, osName, archName)
	if cached, err := loadCacheMetadata(metaPath); err == nil && cached.CanonicalName != "" {
		binPath := binaryCachePath(namespace, handler, osName, archName, cached.CanonicalName)
		if ok, reason := canReuseExtractedBinary(binPath, cached, handler, osName, archName, cached.CanonicalName, sourceMeta.checksum); ok {
			if err := verifyOptionalSHA256(pack.SHA256, cached.BinaryChecksum, handler); err != nil {
				return StagedBinary{}, err
			}
			changed, err := binaryChanged(cached.CanonicalName, cached.BinaryChecksum)
			if err != nil {
				return StagedBinary{}, err
			}
			return StagedBinary{
				Handler:       handler,
				CanonicalName: cached.CanonicalName,
				LocalPath:     binPath,
				SHA256:        cached.BinaryChecksum,
				Changed:       changed,
			}, nil
		} else if reason != "" {
			util.PrintNotify(reason)
		}
	}

	extractedPath, matchedName, err := extractArtifact(sourcePath, handler)
	if err != nil {
		return StagedBinary{}, fmt.Errorf("extract WASM artifact for %s: %w", handler, err)
	}
	installName := matchedName
	if installName == "" {
		installName = canonicalName
	}

	binPath := binaryCachePath(namespace, handler, osName, archName, installName)

	defer func() {
		if extractedPath != sourcePath {
			_ = os.Remove(extractedPath)
		}
	}()

	if err := copyExtractedBinary(extractedPath, binPath); err != nil {
		return StagedBinary{}, fmt.Errorf("cache WASM binary for %s: %w", handler, err)
	}

	checksum, size, err := fileSHA256(binPath)
	if err != nil {
		return StagedBinary{}, err
	}
	if err := verifyOptionalSHA256(pack.SHA256, checksum, handler); err != nil {
		return StagedBinary{}, err
	}

	if err := saveCacheMetadata(metaPath, cacheMetadata{
		Handler:        handler,
		OS:             osName,
		Arch:           archName,
		SourceURL:      sourceMeta.url,
		SourcePath:     sourceMeta.path,
		SourceChecksum: sourceMeta.checksum,
		CanonicalName:  installName,
		BinaryChecksum: checksum,
		BinarySize:     size,
		UpdatedAt:      time.Now().UTC(),
	}); err != nil {
		return StagedBinary{}, err
	}

	changed, err := binaryChanged(installName, checksum)
	if err != nil {
		return StagedBinary{}, err
	}

	return StagedBinary{
		Handler:       handler,
		CanonicalName: installName,
		LocalPath:     binPath,
		SHA256:        checksum,
		Changed:       changed,
	}, nil
}

type sourceMeta struct {
	url      string
	path     string
	checksum string
}

func ensureSource(ctx context.Context, namespace, handler string, pack Pack, osName, archName string, airgap bool) (string, sourceMeta, error) {
	switch {
	case pack.Path != "":
		return ensureSourceFromPath(namespace, handler, pack.Path, osName, archName)
	case pack.URL != "":
		return ensureSourceFromURL(ctx, namespace, handler, pack.URL, osName, archName, airgap)
	default:
		return "", sourceMeta{}, fmt.Errorf("WASM handler %s requires url or path", handler)
	}
}

func ensureSourceFromPath(namespace, handler, userPath, osName, archName string) (string, sourceMeta, error) {
	clean, err := util.ValidateReadableFile(userPath)
	if err != nil {
		return "", sourceMeta{}, fmt.Errorf("read WASM path for %s: %w", handler, err)
	}

	checksum, _, err := fileSHA256(clean)
	if err != nil {
		return "", sourceMeta{}, err
	}

	dest := sourceCachePath(namespace, handler, osName, archName)
	if err := copyExtractedBinary(clean, dest); err != nil {
		return "", sourceMeta{}, fmt.Errorf("cache WASM source for %s: %w", handler, err)
	}

	return dest, sourceMeta{path: clean, checksum: checksum}, nil
}

func ensureSourceFromURL(ctx context.Context, namespace, handler, sourceURL, osName, archName string, airgap bool) (string, sourceMeta, error) {
	dest := sourceCachePath(namespace, handler, osName, archName)
	metaPath := metadataPath(namespace, handler, osName, archName)

	if cached, err := loadCacheMetadata(metaPath); err == nil && cached.SourceURL == sourceURL {
		if checksum, _, err := fileSHA256(dest); err == nil && checksum == cached.SourceChecksum {
			return dest, sourceMeta{url: sourceURL, checksum: checksum}, nil
		}
	}

	util.PrintInfo(fmt.Sprintf("Downloading WASM artifact for %s", handler))
	if err := downloadURL(ctx, sourceURL, dest); err != nil {
		return "", sourceMeta{}, fmt.Errorf("download WASM artifact for %s: %w", handler, err)
	}

	checksum, _, err := fileSHA256(dest)
	if err != nil {
		return "", sourceMeta{}, err
	}
	return dest, sourceMeta{url: sourceURL, checksum: checksum}, nil
}

func downloadURL(ctx context.Context, sourceURL, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer util.IgnoreClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, sourceURL)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), util.DirPerm); err != nil {
		return err
	}

	out, err := util.CreateUserFile(destPath, util.FilePerm)
	if err != nil {
		return err
	}
	defer util.IgnoreClose(out)

	reader := io.Reader(resp.Body)
	if resp.ContentLength > 0 {
		reader = newDownloadProgressReader(resp.Body, resp.ContentLength, fmt.Sprintf("Downloading %s", sourceURL))
	}

	if _, err := io.Copy(out, reader); err != nil {
		return err
	}
	return nil
}

type downloadProgressReader struct {
	reader  io.Reader
	total   int64
	read    int64
	label   string
	started bool
	lastPct int
}

func newDownloadProgressReader(r io.Reader, total int64, label string) *downloadProgressReader {
	return &downloadProgressReader{reader: r, total: total, label: label, lastPct: -1}
}

func (p *downloadProgressReader) Read(b []byte) (int, error) {
	if !p.started {
		p.started = true
		util.PrintProgress(p.label, 0, false)
	}
	n, err := p.reader.Read(b)
	if n > 0 && p.total > 0 {
		p.read += int64(n)
		percent := int(float64(p.read) / float64(p.total) * 100)
		if percent > 100 {
			percent = 100
		}
		if percent-p.lastPct >= 5 || (err == io.EOF && percent == 100) {
			p.lastPct = percent
			util.PrintProgress(p.label, percent, err == io.EOF && percent == 100)
		}
	}
	if err == io.EOF && p.total > 0 && p.lastPct < 100 {
		util.PrintProgress(p.label, 100, true)
	}
	return n, err
}

func verifyOptionalSHA256(expected, actual, handler string) error {
	expected = strings.TrimSpace(strings.ToLower(expected))
	if expected == "" {
		return nil
	}
	if actual != expected {
		return fmt.Errorf("WASM handler %s sha256 mismatch: got %s, want %s", handler, actual, expected)
	}
	return nil
}

func binaryChanged(canonicalName, checksum string) (bool, error) {
	installedPath := filepath.Join(localInstallDir, canonicalName)
	info, err := os.Stat(installedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	if info.IsDir() {
		return true, nil
	}
	installedChecksum, _, err := fileSHA256(installedPath)
	if err != nil {
		return false, err
	}
	return installedChecksum != checksum, nil
}

func fileSHA256(path string) (string, int64, error) {
	file, err := util.OpenValidatedFile(path)
	if err != nil {
		return "", 0, err
	}
	defer util.IgnoreClose(file)

	hasher := sha256.New()
	size, err := io.Copy(hasher, file)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(hasher.Sum(nil)), size, nil
}

func platformToOSArch(platform string) (osName, archName string, err error) {
	parts := strings.Split(platform, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", util.NewInternalError("invalid platform specification " + platform)
	}
	return parts[0], parts[1], nil
}
