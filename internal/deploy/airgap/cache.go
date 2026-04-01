package deployairgap

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"
)

const cacheMetadataFilename = "metadata.json"

type cacheMetadata struct {
	Image       string    `json:"image"`
	Digest      string    `json:"digest"`
	Platform    string    `json:"platform"`
	TarChecksum string    `json:"tarChecksum,omitempty"`
	TarSize     int64     `json:"tarSize,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func fetchRemoteDigest(ctx context.Context, imageRef string, sysCtx *types.SystemContext) (digest.Digest, error) {
	ref, err := parseDockerReference(imageRef)
	if err != nil {
		return "", err
	}
	src, err := ref.NewImageSource(ctx, sysCtx)
	if err != nil {
		return "", err
	}
	defer src.Close()

	manifestBytes, manifestType, err := src.GetManifest(ctx, nil)
	if err != nil {
		return "", err
	}
	if manifest.MIMETypeIsMultiImage(manifestType) {
		list, err := manifest.ListFromBlob(manifestBytes, manifestType)
		if err != nil {
			return "", err
		}
		instanceDigest, err := list.ChooseInstance(sysCtx)
		if err != nil {
			return "", err
		}
		manifestBytes, _, err = src.GetManifest(ctx, &instanceDigest)
		if err != nil {
			return "", err
		}
	}
	return digest.FromBytes(manifestBytes), nil
}

func loadCacheMetadata(path string) (*cacheMetadata, error) {
	data, err := os.ReadFile(path)
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
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func canReuseCachedArtifact(archivePath string, meta *cacheMetadata, imageRef, platform, digestValue string) (bool, string) {
	if meta == nil {
		return false, ""
	}
	if meta.Image != imageRef || meta.Platform != platform {
		return false, ""
	}
	if meta.Digest != digestValue {
		return false, "Cached airgap image digest differs from registry; refreshing cache"
	}
	info, err := os.Stat(archivePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Sprintf("Cached airgap image for %s (%s) is missing on disk; refreshing cache", imageRef, platform)
		}
		return false, fmt.Sprintf("Failed to stat cached airgap image %s: %v", archivePath, err)
	}
	if meta.TarChecksum == "" {
		return false, "Cached airgap image is missing checksum metadata; refreshing cache"
	}
	checksum, _, err := calculateFileChecksum(archivePath)
	if err != nil {
		return false, fmt.Sprintf("Failed to verify cached airgap image: %v", err)
	}
	if checksum != meta.TarChecksum {
		return false, "Cached airgap image checksum mismatch; refreshing cache"
	}
	if meta.TarSize > 0 && info.Size() != meta.TarSize {
		return false, "Cached airgap image size mismatch; refreshing cache"
	}
	return true, ""
}
