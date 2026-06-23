package install

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// DeleteControlPlane removes the edgelet-managed control plane deployment on the local host.
func (agent *LocalEdgelet) DeleteControlPlane() error {
	msg := "Deleting edgelet control plane on " + agent.name
	cmd := agent.edgeletCommand("edgelet controlplane delete", msg)
	if err := agent.runShell(cmd); err != nil && !isEdgeletControlPlaneRemovedError(err) {
		return err
	}
	return nil
}

// DeployFromFile applies an edgelet manifest (Registry or ControlPlane) on the local host.
func (agent *LocalEdgelet) DeployFromFile(manifestPath string) error {
	cmd := fmt.Sprintf("edgelet deploy -f %s", shellQuoteArg(manifestPath))
	return agent.runShell(agent.edgeletCommand(cmd, "Deploying edgelet manifest from "+manifestPath))
}

// RegistryList runs edgelet registry ls and returns stdout.
func (agent *LocalEdgelet) RegistryList() (string, error) {
	stdout, err := util.Exec("", "sh", "-c", agent.edgeletCommand("edgelet registry ls", "Listing edgelet registries"))
	if err != nil {
		return "", err
	}
	return stdout.String(), nil
}

// WriteTempManifest writes YAML to a temp file for edgelet deploy -f.
// For desktop container edgelet, the file is placed under EdgeletContainerManifestDir
// so docker exec edgelet can read it via the container bind mount.
func WriteTempManifest(data []byte, prefix string, cfg EdgeletInstallConfig) (path string, cleanup func(), err error) {
	dir := ""
	if IsDesktopContainerDeploy(cfg) {
		dir = EdgeletContainerManifestDir
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", nil, err
		}
	}

	f, err := os.CreateTemp(dir, prefix+"-*.yaml")
	if err != nil {
		return "", nil, err
	}
	path = f.Name()
	if _, err = f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return "", nil, err
	}
	if err = f.Close(); err != nil {
		os.Remove(path)
		return "", nil, err
	}
	return path, func() { _ = os.Remove(path) }, nil
}

// WriteDeployManifest writes a manifest using this edgelet's install config.
func (agent *LocalEdgelet) WriteDeployManifest(data []byte, prefix string) (path string, cleanup func(), err error) {
	return WriteTempManifest(data, prefix, agent.cfg)
}

// ParseEdgeletRegistryID finds a registry row matching URL and optional username in edgelet registry ls output.
func ParseEdgeletRegistryID(output, registryURL, username string) (int, error) {
	registryURL = strings.TrimSpace(registryURL)
	username = strings.TrimSpace(username)
	if registryURL == "" {
		return 0, fmt.Errorf("registry URL is required")
	}

	for i, line := range strings.Split(output, "\n") {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		id, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		url := fields[1]
		user := fields[3]
		if url != registryURL {
			continue
		}
		if username != "" && user != username {
			continue
		}
		return id, nil
	}
	return 0, fmt.Errorf("edgelet registry %q not found in registry ls output", registryURL)
}
