package client

import (
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
)

func TestMergeLocalAgentFromBackendOnlySeedsSystemAgentMetadataForSystemAgents(t *testing.T) {
	t.Parallel()

	cp := &rsc.LocalControlPlane{
		SystemAgent: &rsc.SystemAgentConfig{
			Package: rsc.Package{Version: "system-pkg"},
		},
	}

	nonSystem := mergeLocalAgentFromBackend(nil, &client.AgentInfo{
		Name:     "edge-3",
		UUID:     "uuid-edge-3",
		Host:     "192.168.139.27",
		IsSystem: false,
	}, cp)
	if nonSystem.Package.Version == "system-pkg" {
		t.Fatalf("non-system agent inherited CP system package: %+v", nonSystem.Package)
	}

	system := mergeLocalAgentFromBackend(nil, &client.AgentInfo{
		Name:     "iofog",
		UUID:     "uuid-iofog",
		Host:     "192.168.1.6",
		IsSystem: true,
	}, cp)
	if system.Package.Version != "system-pkg" {
		t.Fatalf("system agent package = %+v, want system-pkg", system.Package)
	}
}

func TestMergeRemoteAgentFromBackendPreservesSSHHostForLocalCPAgents(t *testing.T) {
	t.Parallel()

	sshHost := "0.0.0.0"
	registrationHost := "192.168.139.27"
	cached := &rsc.RemoteAgent{
		Name: "edge-3",
		Host: sshHost,
		SSH:  rsc.SSH{User: "ubuntu", Port: 22, KeyFile: "/tmp/id_ed25519"},
		Package: rsc.Package{
			Version: "v1.0.0-rc.8",
		},
	}

	merged := mergeRemoteAgentFromBackend(cached, &client.AgentInfo{
		Name:     "edge-3",
		UUID:     "uuid-edge-3",
		Host:     registrationHost,
		IsSystem: false,
	})

	if merged.Host != sshHost {
		t.Fatalf("SSH host = %q, want %q", merged.Host, sshHost)
	}
	if merged.SSH.User != "ubuntu" {
		t.Fatalf("SSH user = %q", merged.SSH.User)
	}
	if merged.Config == nil || merged.Config.Host == nil || *merged.Config.Host != registrationHost {
		t.Fatalf("registration host = %v, want %q", merged.Config, registrationHost)
	}
}
