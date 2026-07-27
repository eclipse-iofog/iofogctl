package deployremotecontroller

import (
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/stretchr/testify/require"
)

func strPtr(v string) *string { return &v }

func TestUnmarshalRemoteControllerAirgap(t *testing.T) {
	ctrl, err := rsc.UnmarshallRemoteController([]byte(`name: remote-2
host: 10.0.0.2
airgap: true
ssh:
  user: ubuntu
  keyFile: ~/.ssh/id_rsa
  port: 22
systemAgent:
  config:
    arch: amd64
`))
	require.NoError(t, err)
	require.True(t, ctrl.Airgap)
}

func TestControllerAddOnRequiresSystemAgentWhenControllerAirgap(t *testing.T) {
	cp := &rsc.RemoteControlPlane{
		Auth: rsc.Auth{Mode: "embedded"},
		Controller: rsc.LocalControllerSpec{
			ControllerConfig: rsc.ControllerConfig{PublicUrl: "http://10.0.0.1:51121"},
		},
		Controllers: []rsc.RemoteController{{
			Name: "remote-1",
			Host: "10.0.0.1",
		}},
	}
	ctrl := &rsc.RemoteController{
		Name:   "remote-2",
		Host:   "10.0.0.2",
		Airgap: true,
		SSH: rsc.SSH{
			User:    "ubuntu",
			KeyFile: "/tmp/key",
			Port:    22,
		},
	}

	require.True(t, deployairgap.ControllerAirgapEnabled(cp, ctrl))
	require.Error(t, deployairgap.ValidateControllerAirgapRequirements(ctrl))
}

func TestControllerAddOnAcceptsSystemAgentWhenControllerAirgap(t *testing.T) {
	arch, found := rsc.ArchStringToID("amd64")
	if !found {
		arch = 0
	}
	ctrl := &rsc.RemoteController{
		Name:   "remote-2",
		Airgap: true,
		SystemAgent: &rsc.SystemAgentConfig{
			AgentConfiguration: &rsc.AgentConfiguration{
				AgentConfiguration: client.AgentConfiguration{
					ArchID:          &arch,
					ContainerEngine: strPtr("edgelet"),
				},
			},
		},
	}

	require.NoError(t, deployairgap.ValidateControllerAirgapRequirements(ctrl))
}
