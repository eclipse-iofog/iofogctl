package deployairgap

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const binaryMetadataFilename = "metadata.json"

type binaryCacheMetadata struct {
	OS        string    `json:"os"`
	Arch      string    `json:"arch"`
	Version   string    `json:"version"`
	Checksum  string    `json:"checksum,omitempty"`
	Size      int64     `json:"size,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// EnsureEdgeletBinary downloads or reuses a cached edgelet release binary.
func EnsureEdgeletBinary(_ context.Context, namespace, osName, archName string) (string, error) {
	localPath := config.GetAirgapBinaryCachePath(namespace, osName, archName)
	cacheDir := filepath.Dir(localPath)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}

	version := util.GetEdgeletBinaryVersion()
	metaPath := filepath.Join(cacheDir, binaryMetadataFilename)
	if cached, err := loadBinaryCacheMetadata(metaPath); err == nil {
		if cached.Version == version && cached.OS == osName && cached.Arch == archName {
			if ok, reason := canReuseCachedBinary(localPath, cached); ok {
				util.PrintInfo(fmt.Sprintf("Reusing cached edgelet binary for %s/%s", osName, archName))
				return localPath, nil
			} else if reason != "" {
				util.PrintNotify(reason)
			}
		}
	}

	util.PrintInfo(fmt.Sprintf("Downloading edgelet binary for %s/%s", osName, archName))
	if err := util.DownloadEdgeletBinary(osName, archName, localPath); err != nil {
		return "", fmt.Errorf("failed to download edgelet binary: %w", err)
	}

	checksum, size, err := calculateFileChecksum(localPath)
	if err != nil {
		return "", err
	}
	if err := saveBinaryCacheMetadata(metaPath, binaryCacheMetadata{
		OS:        osName,
		Arch:      archName,
		Version:   version,
		Checksum:  checksum,
		Size:      size,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		return "", err
	}
	return localPath, nil
}

func loadBinaryCacheMetadata(path string) (*binaryCacheMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta binaryCacheMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func saveBinaryCacheMetadata(path string, meta binaryCacheMetadata) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func canReuseCachedBinary(path string, cached *binaryCacheMetadata) (bool, string) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Sprintf("Cached edgelet binary for %s/%s is missing on disk; refreshing cache", cached.OS, cached.Arch)
		}
		return false, fmt.Sprintf("Failed to stat cached edgelet binary %s: %v", path, err)
	}
	if cached.Checksum == "" {
		return false, "Cached edgelet binary is missing checksum metadata; refreshing cache"
	}
	checksum, size, err := calculateFileChecksum(path)
	if err != nil {
		return false, fmt.Sprintf("Failed to verify cached edgelet binary: %v", err)
	}
	if checksum != cached.Checksum {
		return false, "Cached edgelet binary checksum mismatch; refreshing cache"
	}
	if size != cached.Size || info.Size() != cached.Size {
		return false, "Cached edgelet binary size mismatch; refreshing cache"
	}
	return true, ""
}

// TransferAirgapBinary SCPs a raw edgelet binary to a remote host and returns the remote path.
func TransferAirgapBinary(host string, ssh *rsc.SSH, osName, archName, localPath string) (string, error) {
	if host == "" {
		return "", util.NewInputError("host is required for airgap binary transfer")
	}
	if ssh == nil || ssh.User == "" || ssh.KeyFile == "" {
		return "", util.NewInputError("SSH configuration is required for airgap binary transfer")
	}
	if localPath == "" {
		return "", util.NewInputError("local binary path is required for airgap binary transfer")
	}

	filename, err := util.EdgeletBinaryArtifact(osName, archName)
	if err != nil {
		return "", err
	}

	client, err := util.NewSecureShellClient(ssh.User, host, ssh.KeyFile)
	if err != nil {
		return "", err
	}
	client.SetPort(ssh.Port)
	if err := client.Connect(); err != nil {
		return "", err
	}
	defer util.Log(client.Disconnect)

	hostDir := util.JoinAgentPath(remoteAirgapDir, SanitizeSegment(host))
	if err := client.CreateFolder(hostDir); err != nil {
		return "", err
	}

	file, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	if err := client.CopyTo(file, util.AddTrailingSlash(hostDir), filename, "0700", info.Size()); err != nil {
		return "", err
	}

	remotePath := util.JoinAgentPath(hostDir, filename)
	util.PrintInfo(fmt.Sprintf("Edgelet binary transfer to %s complete", host))
	return remotePath, nil
}

// EnsureAndTransferEdgeletBinary caches, then SCPs, an edgelet binary for the given platform.
func EnsureAndTransferEdgeletBinary(ctx context.Context, namespace, host, platform string, ssh *rsc.SSH) (string, error) {
	osName, archName, err := PlatformToOSArch(platform)
	if err != nil {
		return "", err
	}
	localPath, err := EnsureEdgeletBinary(ctx, namespace, osName, archName)
	if err != nil {
		return "", err
	}
	return TransferAirgapBinary(host, ssh, osName, archName, localPath)
}
