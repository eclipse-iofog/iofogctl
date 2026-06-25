package deployairgap

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"go.podman.io/image/v5/signature"
	"go.podman.io/image/v5/transports/alltransports"
	"go.podman.io/image/v5/types"
)

const (
	remoteAirgapDir = "/tmp/iofogctl-airgap"
	archiveFilename = "image.tar.gz"
)

// imageArtifact represents a pulled and compressed image
type imageArtifact struct {
	platform string
	imageRef string
	digest   string
	path     string
	cleanup  func() error
}

// transferPlan represents a plan to transfer images to a host
type transferPlan struct {
	host     string
	ssh      *rsc.SSH
	platform string
	opts     AirgapTransferOptions
	images   []string // List of image references to transfer
}

// TransferAirgapImages transfers required images to a remote host for airgap deployment
func TransferAirgapImages(ctx context.Context, namespace string, host string, ssh *rsc.SSH, platform string, opts AirgapTransferOptions, images []string) error {
	// Validate inputs
	if host == "" {
		return util.NewInputError("host is required for airgap image transfer")
	}
	if ssh == nil || ssh.User == "" || ssh.KeyFile == "" {
		return util.NewInputError("SSH configuration is required for airgap image transfer")
	}
	if platform == "" {
		return util.NewInputError("platform is required for airgap image transfer")
	}
	if len(images) == 0 {
		return util.NewInputError("at least one image is required for airgap image transfer")
	}

	plan := transferPlan{
		host:     host,
		ssh:      ssh,
		platform: platform,
		opts:     opts,
		images:   images,
	}

	// Prepare artifacts (pull and compress images)
	artifacts, err := prepareArtifacts(ctx, namespace, plan)
	if err != nil {
		return fmt.Errorf("failed to prepare artifacts: %w", err)
	}
	defer cleanupArtifacts(artifacts)

	// Transfer and load each image
	for _, artifact := range artifacts {
		if err := transferAndLoadImage(plan, artifact); err != nil {
			return fmt.Errorf("failed to transfer image %s: %w", artifact.imageRef, err)
		}
	}

	return nil
}

// prepareArtifacts pulls and compresses all required images
func prepareArtifacts(ctx context.Context, namespace string, plan transferPlan) ([]*imageArtifact, error) {
	artifacts := make([]*imageArtifact, 0, len(plan.images))
	for _, imageRef := range plan.images {
		if imageRef == "" {
			continue // Skip empty image references
		}
		artifact, err := ensureArtifact(ctx, plan.platform, imageRef, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare artifact for %s: %w", imageRef, err)
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, nil
}

// ensureArtifact pulls and compresses an image, using persistent cache (same style as offline-image).
func ensureArtifact(ctx context.Context, platform, imageRef string, namespace string) (*imageArtifact, error) {
	sysCtx, err := buildSystemContext(platform, nil)
	if err != nil {
		return nil, err
	}

	cacheDir := config.GetAirgapImageCacheDir(namespace, imageRef, platform)
	if err := os.MkdirAll(cacheDir, util.DirPerm); err != nil {
		return nil, err
	}
	archivePath := filepath.Join(cacheDir, archiveFilename)
	metaPath := filepath.Join(cacheDir, cacheMetadataFilename)

	remoteDigest, err := fetchRemoteDigest(ctx, imageRef, sysCtx)
	if err != nil {
		return nil, err
	}
	remoteDigestStr := remoteDigest.String()

	if cached, err := loadCacheMetadata(metaPath); err == nil {
		if ok, reason := canReuseCachedArtifact(archivePath, cached, imageRef, platform, remoteDigestStr); ok {
			util.PrintInfo(fmt.Sprintf("Reusing cached airgap image for %s (%s)", imageRef, platform))
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

// buildSystemContext builds a system context for image operations
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

// pullCompressedImage pulls an image and compresses it to a tar.gz file
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

	util.PrintInfo(label)
	manifestBytes, err := copy.Image(ctx, policyCtx, destRef, srcRef, &copy.Options{
		SourceCtx:          sysCtx,
		ImageListSelection: copy.CopySystemImage,
	})
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

// parseDockerReference parses a docker image reference
func parseDockerReference(imageRef string) (types.ImageReference, error) {
	if strings.Contains(imageRef, "://") {
		return alltransports.ParseImageName(imageRef)
	}
	return alltransports.ParseImageName("docker://" + imageRef)
}

// insecurePolicyContext creates a policy context that accepts any image
func insecurePolicyContext() (*signature.PolicyContext, error) {
	policyJSON := []byte(`{"default":[{"type":"insecureAcceptAnything"}]}`)
	policy, err := signature.NewPolicyFromBytes(policyJSON)
	if err != nil {
		return nil, err
	}
	return signature.NewPolicyContext(policy)
}

// compressToGzip compresses a file to gzip format
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

// calculateFileChecksum calculates SHA256 checksum and size of a file
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

// transferAndLoadImage transfers an image artifact to remote host and loads it
func transferAndLoadImage(plan transferPlan, artifact *imageArtifact) error {
	ssh, err := util.NewSecureShellClient(plan.ssh.User, plan.host, plan.ssh.KeyFile)
	if err != nil {
		return err
	}
	ssh.SetPort(plan.ssh.Port)
	if err := ssh.Connect(); err != nil {
		return err
	}
	defer util.Log(ssh.Disconnect)

	// Create remote directory
	hostDir := util.JoinAgentPath(remoteAirgapDir, SanitizeSegment(plan.host))
	if err := ssh.CreateFolder(hostDir); err != nil {
		return err
	}

	// Open and transfer file
	file, err := util.OpenValidatedFile(artifact.path)
	if err != nil {
		return err
	}
	defer util.IgnoreClose(file)
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create filename from image reference
	imageName := strings.ReplaceAll(strings.ReplaceAll(artifact.imageRef, "/", "_"), ":", "_")
	filename := fmt.Sprintf("%s-%s.tar.gz", SanitizeSegment(imageName), strings.ReplaceAll(plan.platform, "/", "_"))

	// Copy file to remote host
	copyErr := ssh.CopyTo(file, util.AddTrailingSlash(hostDir), filename, "0600", info.Size())
	if copyErr != nil {
		return copyErr
	}
	remotePath := util.JoinAgentPath(hostDir, filename)

	// Load image using deployment/engine matrix
	loadCmd := ImageLoadCommand(plan.opts, remotePath)
	if _, err := ssh.Run(loadCmd); err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	// Clean up remote archive (decompressed tar removed by edgelet load command)
	if _, err := ssh.Run("sudo rm -f " + remotePath); err != nil {
		util.PrintNotify(fmt.Sprintf("Warning: Failed to remove remote file %s: %v", remotePath, err))
	}

	util.PrintInfo(fmt.Sprintf("%s transfer to %s complete", artifact.imageRef, plan.host))
	return nil
}

// cleanupArtifacts cleans up temporary artifacts
func cleanupArtifacts(artifacts []*imageArtifact) {
	for _, artifact := range artifacts {
		if artifact == nil || artifact.cleanup == nil {
			continue
		}
		if err := artifact.cleanup(); err != nil {
			util.PrintNotify(err.Error())
		}
	}
}
