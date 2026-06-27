package deployremotecontrolplane

import (
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/stretchr/testify/require"
)

func TestBuildFirstSystemAgentConfig_DefaultsNative(t *testing.T) {
	ctrl := &rsc.RemoteController{Name: "remote-1", Host: "10.0.0.1"}
	cfg := buildFirstSystemAgentConfig(nil, ctrl, &rsc.SystemAgentConfig{})

	require.NotNil(t, cfg.IsSystem)
	require.True(t, *cfg.IsSystem)
	require.NotNil(t, cfg.DeploymentType)
	require.Equal(t, deploymentTypeNative, *cfg.DeploymentType)
	require.NotNil(t, cfg.UpstreamRouters)
	require.Empty(t, *cfg.UpstreamRouters)
	require.NotNil(t, cfg.UpstreamNatsServers)
	require.Empty(t, *cfg.UpstreamNatsServers)
	require.Equal(t, iofog.RouterModeInterior, *cfg.RouterMode)
	require.Equal(t, iofog.NatsModeServer, *cfg.NatsMode)
}

func TestBuildNextSystemAgentConfig_DefaultUpstream(t *testing.T) {
	ctrl := &rsc.RemoteController{Name: "remote-2", Host: "10.0.0.2"}
	cfg := buildNextSystemAgentConfig(nil, ctrl, &rsc.SystemAgentConfig{})

	require.NotNil(t, cfg.UpstreamRouters)
	require.Equal(t, []string{"default-router"}, *cfg.UpstreamRouters)
	require.NotNil(t, cfg.UpstreamNatsServers)
	require.Equal(t, []string{"default-nats-hub"}, *cfg.UpstreamNatsServers)
	require.Equal(t, deploymentTypeNative, *cfg.DeploymentType)
}

func TestBuildNextSystemAgentConfig_AppendsDefaultUpstream(t *testing.T) {
	ctrl := &rsc.RemoteController{Name: "remote-2", Host: "10.0.0.2"}
	upstreamRouters := []string{"custom-router"}
	upstreamNats := []string{"custom-nats"}
	sys := &rsc.SystemAgentConfig{
		AgentConfiguration: &rsc.AgentConfiguration{
			AgentConfiguration: client.AgentConfiguration{
				UpstreamRouters:     &upstreamRouters,
				UpstreamNatsServers: &upstreamNats,
			},
		},
	}

	cfg := buildNextSystemAgentConfig(nil, ctrl, sys)
	require.Equal(t, []string{"custom-router", "default-router"}, *cfg.UpstreamRouters)
	require.Equal(t, []string{"custom-nats", "default-nats-hub"}, *cfg.UpstreamNatsServers)
}

func TestBuildFirstSystemAgentConfig_ContainerWhenImageSet(t *testing.T) {
	ctrl := &rsc.RemoteController{Name: "remote-1", Host: "10.0.0.1"}
	sys := &rsc.SystemAgentConfig{
		Package: rsc.Package{
			Container: rsc.RemoteContainer{Image: "ghcr.io/example/edgelet:1.0"},
		},
	}
	cfg := buildFirstSystemAgentConfig(nil, ctrl, sys)
	require.Equal(t, deploymentTypeContainer, *cfg.DeploymentType)
}

func TestBuildFirstSystemAgentConfig_ForcesSystem(t *testing.T) {
	ctrl := &rsc.RemoteController{Name: "remote-1", Host: "10.0.0.1"}
	sys := &rsc.SystemAgentConfig{
		AgentConfiguration: &rsc.AgentConfiguration{
			AgentConfiguration: client.AgentConfiguration{
				IsSystem: iutil.MakeBoolPtr(false),
			},
		},
	}
	cfg := buildFirstSystemAgentConfig(nil, ctrl, sys)
	require.True(t, *cfg.IsSystem)
}
