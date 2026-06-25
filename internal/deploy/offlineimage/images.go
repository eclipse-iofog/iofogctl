package deployofflineimage

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/opencontainers/go-digest"
	"go.podman.io/image/v5/copy"
	"go.podman.io/image/v5/manifest"
	"go.podman.io/image/v5/signature"
	"go.podman.io/image/v5/transports/alltransports"
	"go.podman.io/image/v5/types"
)

type imageArtifact struct {
	platform string
	imageRef string
	digest   string
	path     string
	cleanup  func() error
}

type cacheMetadata struct {
	Image       string    `json:"image"`
	Digest      string    `json:"digest"`
	Platform    string    `json:"platform"`
	TarChecksum string    `json:"tarChecksum,omitempty"`
	TarSize     int64     `json:"tarSize,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

const archiveFilename = "image.tar.gz"

func (exe *executor) ensureArtifact(ctx context.Context, platform, imageRef string) (*imageArtifact, error) {
	sysCtx, err := buildSystemContext(platform, exe.spec.Auth)
	if err != nil {
		return nil, err
	}

	if exe.noCache {
		return exe.pullToTemp(ctx, platform, imageRef, sysCtx)
	}

	cacheDir := config.GetOfflineImageCacheDir(exe.namespace, exe.spec.Name, sanitizeSegment(platform))
	if err := os.MkdirAll(cacheDir, util.DirPerm); err != nil {
		return nil, err
	}
	archivePath := filepath.Join(cacheDir, archiveFilename)
	metaPath := filepath.Join(cacheDir, "metadata.json")

	remoteDigest, err := fetchRemoteDigest(ctx, imageRef, sysCtx)
	if err != nil {
		return nil, err
	}

	remoteDigestStr := remoteDigest.String()

	if cached, err := loadCacheMetadata(metaPath); err == nil {
		if ok, reason := canReuseCachedArtifact(archivePath, cached, imageRef, platform, remoteDigestStr); ok {
			util.PrintInfo(fmt.Sprintf("Reusing cached offline image for %s (%s)", imageRef, platform))
			return &imageArtifact{
				platform: platform,
				imageRef: imageRef,
				digest:   cached.Digest,
				path:     archivePath,
			}, nil
		} else if reason != "" {
			util.PrintNotify(reason)
			_ = os.Remove(archivePath)
		}
	}

	label := fmt.Sprintf("Pulling %s (%s)", imageRef, platform)
	_, checksum, size, err := pullCompressedImage(ctx, imageRef, archivePath, sysCtx, label)
	if err != nil {
		return nil, err
	}
	if err := saveCacheMetadata(metaPath, cacheMetadata{
		Image:       imageRef,
		Digest:      remoteDigestStr,
		Platform:    platform,
		TarChecksum: checksum,
		TarSize:     size,
		UpdatedAt:   time.Now().UTC(),
	}); err != nil {
		return nil, err
	}

	return &imageArtifact{
		platform: platform,
		imageRef: imageRef,
		digest:   remoteDigestStr,
		path:     archivePath,
	}, nil
}

func (exe *executor) pullToTemp(ctx context.Context, platform, imageRef string, sysCtx *types.SystemContext) (*imageArtifact, error) {
	dir, err := os.MkdirTemp("", "iofogctl-offline-*")
	if err != nil {
		return nil, err
	}
	tarPath := filepath.Join(dir, archiveFilename)
	label := fmt.Sprintf("Pulling %s (%s)", imageRef, platform)
	digestValue, _, _, err := pullCompressedImage(ctx, imageRef, tarPath, sysCtx, label)
	if err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
	}
	return &imageArtifact{
		platform: platform,
		imageRef: imageRef,
		digest:   digestValue,
		path:     tarPath,
		cleanup: func() error {
			return os.RemoveAll(dir)
		},
	}, nil
}

func buildSystemContext(platform string, auth *rsc.OfflineImageAuth) (*types.SystemContext, error) {
	parts := strings.Split(platform, "/")
	if len(parts) < 2 {
		return nil, util.NewInternalError("invalid platform specification " + platform)
	}
	ctx := &types.SystemContext{
		OSChoice:           parts[0],
		ArchitectureChoice: parts[1],
	}
	if ctx.ArchitectureChoice == "arm64" {
		ctx.VariantChoice = "v8"
	}
	if auth != nil && auth.Username != "" {
		ctx.DockerAuthConfig = &types.DockerAuthConfig{
			Username: auth.Username,
			Password: auth.Password,
		}
	}
	return ctx, nil
}

func pullCompressedImage(ctx context.Context, imageRef, archivePath string, sysCtx *types.SystemContext, label string) (digestValue string, checksum string, size int64, err error) {
	destDir := filepath.Dir(archivePath)
	if err := os.MkdirAll(destDir, util.DirPerm); err != nil {
		return "", "", 0, err
	}
	rawPath := archivePath + ".raw"
	pathsToClean := []string{archivePath, rawPath}
	for _, p := range pathsToClean {
		if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
			return "", "", 0, err
		}
	}

	srcRef, err := parseDockerReference(imageRef)
	if err != nil {
		return "", "", 0, err
	}
	rawAbs, err := filepath.Abs(rawPath)
	if err != nil {
		return "", "", 0, err
	}
	destString := fmt.Sprintf("docker-archive:%s:%s", rawAbs, imageRef)
	destRef, err := alltransports.ParseImageName(destString)
	if err != nil {
		return "", "", 0, err
	}

	policyCtx, err := insecurePolicyContext()
	if err != nil {
		return "", "", 0, err
	}
	defer func() { _ = policyCtx.Destroy() }()

	progressCh := make(chan types.ProgressProperties, 1)
	progressDone := startProgressTracker(label, progressCh)

	manifestBytes, err := copy.Image(ctx, policyCtx, destRef, srcRef, &copy.Options{
		SourceCtx:          sysCtx,
		ImageListSelection: copy.CopySystemImage,
		Progress:           progressCh,
		ProgressInterval:   500 * time.Millisecond,
	})
	close(progressCh)
	progressDone()
	if err != nil {
		_ = os.Remove(rawPath)
		return "", "", 0, err
	}

	if err := compressToGzip(rawPath, archivePath); err != nil {
		_ = os.Remove(rawPath)
		return "", "", 0, err
	}
	_ = os.Remove(rawPath)

	checksum, size, err = calculateFileChecksum(archivePath)
	if err != nil {
		return "", "", 0, err
	}
	util.PrintInfo(fmt.Sprintf("%s complete", label))
	return digest.FromBytes(manifestBytes).String(), checksum, size, nil
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

func canReuseCachedArtifact(archivePath string, meta *cacheMetadata, imageRef, platform, digestValue string) (bool, string) {
	if meta == nil {
		return false, ""
	}
	if meta.Image != imageRef || meta.Platform != platform {
		return false, ""
	}
	if meta.Digest != digestValue {
		return false, "Cached offline image digest differs from registry; refreshing cache"
	}
	info, err := os.Stat(archivePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Sprintf("Cached offline image for %s (%s) is missing on disk; refreshing cache", imageRef, platform)
		}
		return false, fmt.Sprintf("Failed to stat cached offline image %s: %v", archivePath, err)
	}
	if meta.TarChecksum == "" {
		return false, "Cached offline image is missing checksum metadata; refreshing cache"
	}
	checksum, _, err := calculateFileChecksum(archivePath)
	if err != nil {
		return false, fmt.Sprintf("Failed to verify cached offline image: %v", err)
	}
	if checksum != meta.TarChecksum {
		return false, "Cached offline image checksum mismatch; refreshing cache"
	}
	if meta.TarSize > 0 && info.Size() != meta.TarSize {
		return false, "Cached offline image size mismatch; refreshing cache"
	}
	return true, ""
}

func parseDockerReference(imageRef string) (types.ImageReference, error) {
	if strings.Contains(imageRef, "://") {
		return alltransports.ParseImageName(imageRef)
	}
	return alltransports.ParseImageName("docker://" + imageRef)
}

func insecurePolicyContext() (*signature.PolicyContext, error) {
	policyJSON := []byte(`{"default":[{"type":"insecureAcceptAnything"}]}`)
	policy, err := signature.NewPolicyFromBytes(policyJSON)
	if err != nil {
		return nil, err
	}
	return signature.NewPolicyContext(policy)
}

func compressToGzip(src, dst string) error {
	source, err := util.OpenValidatedFile(src)
	if err != nil {
		return err
	}
	defer util.IgnoreClose(source)

	if err := os.RemoveAll(dst); err != nil && !os.IsNotExist(err) {
		return err
	}

	destFile, err := util.CreateUserFile(dst, util.FilePerm)
	if err != nil {
		return err
	}
	defer util.IgnoreClose(destFile)

	gzipWriter := gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	if _, err := io.Copy(gzipWriter, source); err != nil {
		return err
	}

	return nil
}

func calculateFileChecksum(path string) (string, int64, error) {
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
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return util.WriteValidatedFile(path, data, util.FilePerm)
}
