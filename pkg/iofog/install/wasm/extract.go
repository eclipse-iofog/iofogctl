package wasm

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const maxWasmExtractSize = 250 * 1024 * 1024 // 250MiB transfer cap

func extractArtifact(sourcePath, handler string) (destPath, matchedName string, err error) {
	candidates := Candidates(handler)
	if len(candidates) == 0 {
		return "", "", fmt.Errorf("unknown WASM handler %q", handler)
	}

	if isGzipTar(sourcePath) {
		return extractFromTarGz(sourcePath, candidates)
	}
	if isRawELF(sourcePath) {
		canonical := candidates[0]
		return sourcePath, canonical, nil
	}
	return "", "", fmt.Errorf("unsupported WASM artifact format in %s", sourcePath)
}

func isGzipTar(path string) bool {
	file, err := util.OpenValidatedFile(path)
	if err != nil {
		return false
	}
	defer util.IgnoreClose(file)

	header := make([]byte, 2)
	if _, err := io.ReadFull(file, header); err != nil {
		return false
	}
	return header[0] == 0x1f && header[1] == 0x8b
}

func isRawELF(path string) bool {
	file, err := util.OpenValidatedFile(path)
	if err != nil {
		return false
	}
	defer util.IgnoreClose(file)

	header := make([]byte, 4)
	if _, err := io.ReadFull(file, header); err != nil {
		return false
	}
	return string(header) == "\x7fELF"
}

func extractFromTarGz(sourcePath string, candidates []string) (string, string, error) {
	file, err := util.OpenValidatedFile(sourcePath)
	if err != nil {
		return "", "", err
	}
	defer util.IgnoreClose(file)

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", "", fmt.Errorf("open gzip archive: %w", err)
	}
	defer util.IgnoreClose(gzReader)

	tarReader := tar.NewReader(gzReader)
	entryBasenames := make(map[string]struct{})

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", "", fmt.Errorf("read tar archive: %w", err)
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			continue
		}
		base := filepath.Base(header.Name)
		if base == "." || base == "/" || base == "" {
			continue
		}
		if _, seen := entryBasenames[base]; seen {
			continue
		}
		entryBasenames[base] = struct{}{}
	}

	for _, candidate := range candidates {
		if _, ok := entryBasenames[candidate]; ok {
			return materializeTarEntry(sourcePath, candidate)
		}
	}

	return "", "", fmt.Errorf("no catalog candidate found in archive (tried: %s)", strings.Join(candidates, ", "))
}

func materializeTarEntry(sourcePath, matchedName string) (string, string, error) {
	file, err := util.OpenValidatedFile(sourcePath)
	if err != nil {
		return "", "", err
	}
	defer util.IgnoreClose(file)

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", "", err
	}
	defer util.IgnoreClose(gzReader)

	tarReader := tar.NewReader(gzReader)
	destPath, err := os.CreateTemp("", "iofogctl-wasm-*")
	if err != nil {
		return "", "", err
	}
	destName := destPath.Name()
	cleanup := true
	defer func() {
		util.IgnoreClose(destPath)
		if cleanup {
			_ = os.Remove(destName)
		}
	}()

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", "", err
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			continue
		}
		if filepath.Base(header.Name) != matchedName {
			continue
		}
		if header.Size < 0 || header.Size > maxWasmExtractSize {
			return "", "", fmt.Errorf("archive entry %q exceeds size limit", matchedName)
		}
		limited := io.LimitReader(tarReader, maxWasmExtractSize+1)
		n, err := io.Copy(destPath, limited)
		if err != nil {
			return "", "", fmt.Errorf("extract %q from archive: %w", matchedName, err)
		}
		if n > maxWasmExtractSize {
			return "", "", fmt.Errorf("extract %q: decompressed size exceeds limit", matchedName)
		}
		if err := destPath.Chmod(util.ExecPerm); err != nil {
			return "", "", err
		}
		cleanup = false
		return destName, matchedName, nil
	}

	return "", "", fmt.Errorf("candidate %q not found while extracting archive", matchedName)
}

func copyExtractedBinary(srcPath, destPath string) error {
	src, err := util.OpenValidatedFile(srcPath)
	if err != nil {
		return err
	}
	defer util.IgnoreClose(src)

	if err := os.MkdirAll(filepath.Dir(destPath), util.DirPerm); err != nil {
		return err
	}

	dst, err := util.CreateUserFile(destPath, util.ExecPerm)
	if err != nil {
		return err
	}
	defer util.IgnoreClose(dst)

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}
