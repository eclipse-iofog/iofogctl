package client

import (
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
)

func TestMergeRemoteAgentFromBackendPreservesSSHHost(t *testing.T) {
	t.Parallel()

	registrationHost := "192.168.139.184"
	sshHost := "0.0.0.0"
	cached := &rsc.RemoteAgent{
		Name:   "edge-1",
		Host:   sshHost,
		SSH:    rsc.SSH{User: "ubuntu2", Port: 32222, KeyFile: "/tmp/id_ed25519"},
		Airgap: true,
		Package: rsc.Package{
			Version: "v1.0.0-rc.6",
		},
	}

	merged := mergeRemoteAgentFromBackend(cached, &client.AgentInfo{
		Name: "edge-1",
		UUID: "055bcc8d-91d2-445e-b7a2-f48e5bd98046",
		Host: registrationHost,
	})

	if merged.Host != sshHost {
		t.Fatalf("SSH host = %q, want %q", merged.Host, sshHost)
	}
	if merged.Config == nil || merged.Config.Host == nil || *merged.Config.Host != registrationHost {
		t.Fatalf("registration host = %v, want %q", merged.Config, registrationHost)
	}
	if merged.SSH.User != "ubuntu2" || !merged.Airgap || merged.Package.Version != "v1.0.0-rc.6" {
		t.Fatalf("cached deploy metadata lost: %+v", merged)
	}
	if merged.UUID != "055bcc8d-91d2-445e-b7a2-f48e5bd98046" {
		t.Fatalf("UUID = %q", merged.UUID)
	}
}

func TestMergeRemoteAgentFromBackendWithoutCacheUsesBackendHost(t *testing.T) {
	t.Parallel()

	merged := mergeRemoteAgentFromBackend(nil, &client.AgentInfo{
		Name: "edge-1",
		UUID: "uuid-1",
		Host: "192.168.139.184",
	})

	if merged.Host != "192.168.139.184" {
		t.Fatalf("Host = %q", merged.Host)
	}
	if merged.Config == nil || merged.Config.Host == nil || *merged.Config.Host != "192.168.139.184" {
		t.Fatalf("registration host = %v", merged.Config)
	}
}
