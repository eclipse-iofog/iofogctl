package install

import (
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

func TestDefaultContainerEngineURL(t *testing.T) {
	if got := DefaultContainerEngineURL("docker"); got != "unix:///var/run/docker.sock" {
		t.Fatalf("docker default = %q", got)
	}
}

func TestResolveContainerEngineURLPrefersAgentConfig(t *testing.T) {
	custom := "unix:///custom/docker.sock"
	cfg := &client.AgentConfiguration{ContainerEngineURL: strPtr(custom)}
	if got := ResolveContainerEngineURL("docker", cfg); got != custom {
		t.Fatalf("got %q want %q", got, custom)
	}
}

func TestLocalContainerEngineForHostOps(t *testing.T) {
	if got := LocalContainerEngineForHostOps(nil); got != DefaultLocalContainerEngine {
		t.Fatalf("nil cfg engine = %q", got)
	}
	cfg := &client.AgentConfiguration{ContainerEngine: strPtr("edgelet")}
	if got := LocalContainerEngineForHostOps(cfg); got != DefaultLocalContainerEngine {
		t.Fatalf("edgelet engine mapped to %q", got)
	}
	cfg.ContainerEngine = strPtr("podman")
	if got := LocalContainerEngineForHostOps(cfg); got != "podman" {
		t.Fatalf("podman engine = %q", got)
	}
}
