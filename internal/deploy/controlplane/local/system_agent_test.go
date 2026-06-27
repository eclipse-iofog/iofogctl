package deploylocalcontrolplane

import (
	"testing"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/stretchr/testify/require"
)

func TestBuildLocalSystemAgentConfig_ForcesSystemDefaults(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-datasance.yaml")
	cfg, err := buildLocalSystemAgentConfig(&cp, "iofog")
	require.NoError(t, err)

	require.NotNil(t, cfg.IsSystem)
	require.True(t, *cfg.IsSystem)
	require.NotNil(t, cfg.RouterMode)
	require.Equal(t, iofog.RouterModeInterior, *cfg.RouterMode)
	require.NotNil(t, cfg.DeploymentType)
	require.Equal(t, deploymentTypeNative, *cfg.DeploymentType)
	require.NotNil(t, cfg.UpstreamRouters)
	require.Empty(t, *cfg.UpstreamRouters)
	require.NotNil(t, cfg.NatsMode)
	require.Equal(t, iofog.NatsModeServer, *cfg.NatsMode)
	require.Equal(t, "iofog", cfg.Name)
	require.NotNil(t, cfg.Host)
	require.NotEmpty(t, *cfg.Host)
}

func TestBuildLocalSystemAgentConfig_OverridesFalseIsSystem(t *testing.T) {
	cp := loadLocalControlPlaneFixture(t, "controlplane-datasance.yaml")
	cp.SystemAgent.AgentConfiguration.IsSystem = iutil.MakeBoolPtr(false)

	cfg, err := buildLocalSystemAgentConfig(&cp, "iofog")
	require.NoError(t, err)
	require.True(t, *cfg.IsSystem)
}

func TestResolveLocalControlPlaneEndpoint(t *testing.T) {
	cp := rsc.LocalControlPlane{
		Endpoint: "https://cp.example.com",
	}
	require.Equal(t, "https://cp.example.com", resolveLocalControlPlaneEndpoint(&cp))

	cp = rsc.LocalControlPlane{
		Controller: rsc.LocalControllerSpec{
			ControllerConfig: rsc.ControllerConfig{PublicUrl: "https://public.example.com"},
		},
	}
	require.Equal(t, "https://public.example.com", resolveLocalControlPlaneEndpoint(&cp))

	cp = rsc.LocalControlPlane{}
	require.Equal(t, "http://localhost:51121", resolveLocalControlPlaneEndpoint(&cp))
}
