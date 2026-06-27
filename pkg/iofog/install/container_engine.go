package install

import (
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	// TODO(v3.8.0): remove Moby client after local and remote control plane no longer use Go container deploy.
	dockengine "github.com/eclipse-iofog/iofogctl/pkg/containerengine/docker"
)

// DefaultLocalContainerEngine is the host runtime used for local edgelet container operations.
const DefaultLocalContainerEngine = "docker"

// DefaultContainerEngineURL returns the default unix socket URL for a container engine type.
func DefaultContainerEngineURL(engine string) string {
	return defaultContainerEngineURL(engine)
}

// ResolveContainerEngine returns the configured container engine name.
func ResolveContainerEngine(cfg *client.AgentConfiguration) string {
	return resolveContainerEngine(cfg)
}

// ResolveContainerEngineURL resolves the daemon socket URL from agent config and engine type.
func ResolveContainerEngineURL(engine string, cfg *client.AgentConfiguration) string {
	return resolveContainerEngineURL(engine, cfg)
}

// NewContainerEngineClient dials the container engine using ResolveContainerEngineURL.
func NewContainerEngineClient(engine string, agentCfg *client.AgentConfiguration) (*dockengine.Client, error) {
	return dockengine.NewWithHost(ResolveContainerEngineURL(engine, agentCfg))
}

// LocalContainerEngineForHostOps maps agent config to the host runtime used for docker/podman API calls.
// Container-mode local agents use docker/podman; native edgelet-managed engines are not API-dial targets here.
func LocalContainerEngineForHostOps(agentCfg *client.AgentConfiguration) string {
	engine := ResolveContainerEngine(agentCfg)
	if strings.EqualFold(engine, "edgelet") {
		return DefaultLocalContainerEngine
	}
	return engine
}
