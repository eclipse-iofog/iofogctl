package resource

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func edgeletFixturePath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "edgelet", name)
}

func loadEdgeletFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(edgeletFixturePath(name))
	require.NoError(t, err)
	return data
}

func strPtr(v string) *string { return &v }

func assertGoldenAgentConfig(t *testing.T, cfg *AgentConfiguration) {
	t.Helper()
	require.NotNil(t, cfg)
	require.Equal(t, "edgelet running on device", cfg.Description)
	require.Equal(t, 46.204391, cfg.Latitude)
	require.Equal(t, 6.143158, cfg.Longitude)
	require.NotNil(t, cfg.Arch)
	require.Equal(t, "amd64", *cfg.Arch)
	require.NotNil(t, cfg.DeploymentType)
	require.Equal(t, "native", *cfg.DeploymentType)
	require.NotNil(t, cfg.ContainerEngine)
	require.Equal(t, "edgelet", *cfg.ContainerEngine)
	require.NotNil(t, cfg.ContainerEngineURL)
	require.Equal(t, "unix:///run/edgelet/containerd.sock", *cfg.ContainerEngineURL)
	require.NotNil(t, cfg.DiskLimit)
	require.Equal(t, int64(50), *cfg.DiskLimit)
	require.NotNil(t, cfg.RouterMode)
	require.Equal(t, "edge", *cfg.RouterMode)
	require.NotNil(t, cfg.MessagingPort)
	require.Equal(t, 5671, *cfg.MessagingPort)
	require.NotNil(t, cfg.NatsMode)
	require.Equal(t, "leaf", *cfg.NatsMode)
	require.NotNil(t, cfg.UpstreamNatsServers)
	require.Equal(t, []string{"default-nats-hub"}, *cfg.UpstreamNatsServers)
}

func TestGoldenUnmarshalRemoteAgent(t *testing.T) {
	raw := loadEdgeletFixture(t, "remote-agent-spec.yaml")
	agent, err := UnmarshallRemoteAgent(raw)
	require.NoError(t, err)
	require.Equal(t, "30.40.50.6", agent.Host)
	require.Equal(t, "foo", agent.SSH.User)
	require.Contains(t, agent.SSH.KeyFile, "id_rsa")
	require.Equal(t, 22, agent.SSH.Port)
	require.NotNil(t, agent.Config)
	assertGoldenAgentConfig(t, agent.Config)
}

func TestGoldenUnmarshalRemoteAgentPackageRegistry(t *testing.T) {
	raw := loadEdgeletFixture(t, "remote-agent-package-registry.yaml")
	agent, err := UnmarshallRemoteAgent(raw)
	require.NoError(t, err)
	require.True(t, agent.Airgap)
	require.Equal(t, "1.0.0-rc.8", agent.Package.Version)
	require.Equal(t, "ghcr.io/datasance/edgelet:1.0.0-rc.8", agent.Package.Container.Image)
	require.Equal(t, "ghcr.io", agent.Package.Container.Registry)
	require.Equal(t, "foo", agent.Package.Container.Username)
	require.Equal(t, "bar", agent.Package.Container.Password)
	require.NotNil(t, agent.Scripts)
	require.Equal(t, "/tmp/my-scripts", agent.Scripts.Directory)
	require.Equal(t, "install_deps.sh", agent.Scripts.Deps.Name)
	require.Equal(t, "install.sh", agent.Scripts.Install.Name)
	require.Equal(t, []string{"1.0.0-rc.8"}, agent.Scripts.Install.Args)
	require.Equal(t, "uninstall.sh", agent.Scripts.Uninstall.Name)
}

func TestGoldenUnmarshalLocalAgent(t *testing.T) {
	raw := loadEdgeletFixture(t, "local-agent-spec.yaml")
	agent, err := UnmarshallLocalAgent(raw)
	require.NoError(t, err)
	require.Equal(t, "local", agent.Name)
	require.Equal(t, "1.0.0-rc.8", agent.Package.Version)
	require.Equal(t, "ghcr.io/datasance/edgelet:1.0.0-rc.8", agent.Package.Container.Image)
	require.NotNil(t, agent.Config)
	assertGoldenAgentConfig(t, agent.Config)
	require.Equal(t, "30.40.50.6", agent.GetHost())
}

func TestGoldenUnmarshalAgentConfiguration(t *testing.T) {
	raw := loadEdgeletFixture(t, "agent-config-spec.yaml")
	cfg, err := UnmarshallAgentConfiguration(raw)
	require.NoError(t, err)
	require.Equal(t, "agent running on VM", cfg.Description)
	require.Equal(t, 46.204391, cfg.Latitude)
	require.NotNil(t, cfg.Arch)
	require.Equal(t, "riscv64", *cfg.Arch)
	require.NotNil(t, cfg.ContainerEngine)
	require.Equal(t, "edgelet", *cfg.ContainerEngine)
	require.NotNil(t, cfg.RouterMode)
	require.Equal(t, "edge", *cfg.RouterMode)
	require.NotNil(t, cfg.NatsMode)
	require.Equal(t, "leaf", *cfg.NatsMode)
	require.NotNil(t, cfg.LogLevel)
	require.Equal(t, "INFO", *cfg.LogLevel)
}

func TestGoldenRoundTripRemoteAgent(t *testing.T) {
	raw := loadEdgeletFixture(t, "remote-agent-spec.yaml")
	agent, err := UnmarshallRemoteAgent(raw)
	require.NoError(t, err)

	out, err := yaml.Marshal(&agent)
	require.NoError(t, err)

	var round RemoteAgent
	require.NoError(t, yaml.UnmarshalStrict(out, &round))
	require.Equal(t, agent.Host, round.Host)
	require.Equal(t, agent.SSH, round.SSH)
	require.Equal(t, *agent.Config.Arch, *round.Config.Arch)
}

func TestArchStringToID(t *testing.T) {
	id, ok := ArchStringToID("amd64")
	require.True(t, ok)
	require.Equal(t, int64(1), id)

	id, ok = ArchStringToID("riscv64")
	require.True(t, ok)
	require.Equal(t, int64(3), id)

	_, ok = ArchStringToID("x86")
	require.False(t, ok)
}

func TestArchIDToString(t *testing.T) {
	name, ok := ArchIDToString(2)
	require.True(t, ok)
	require.Equal(t, "arm64", name)

	name, ok = ArchIDToString(0)
	require.True(t, ok)
	require.Equal(t, "auto", name)
}

func TestLocalAgentGetHostFallback(t *testing.T) {
	agent := &LocalAgent{}
	require.Equal(t, "localhost", agent.GetHost())

	agent.Host = "192.168.1.10"
	require.Equal(t, "192.168.1.10", agent.GetHost())

	host := "10.0.0.5"
	agent.Config = &AgentConfiguration{}
	agent.Config.Host = &host
	require.Equal(t, "10.0.0.5", agent.GetHost())
}
