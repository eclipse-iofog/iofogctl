package deploy

import (
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"
	"github.com/stretchr/testify/require"
)

func TestBuildSyntheticAgentConfiguration_ForcesNonSystemAgent(t *testing.T) {
	cfg := buildSyntheticAgentConfiguration(nil, "192.168.139.27")
	require.NotNil(t, cfg.Host)
	require.Equal(t, "192.168.139.27", *cfg.Host)
	require.NotNil(t, cfg.IsSystem)
	require.False(t, *cfg.IsSystem)
}

func TestBuildSyntheticAgentConfiguration_PreservesInlineConfig(t *testing.T) {
	deployConfig := &rsc.AgentConfiguration{
		AgentConfiguration: client.AgentConfiguration{
			DeploymentType:  iutil.MakeStrPtr("native"),
			ContainerEngine: iutil.MakeStrPtr("edgelet"),
		},
	}

	cfg := buildSyntheticAgentConfiguration(deployConfig, "192.168.139.27")
	require.NotNil(t, cfg.IsSystem)
	require.False(t, *cfg.IsSystem)
	require.NotNil(t, cfg.DeploymentType)
	require.Equal(t, "native", *cfg.DeploymentType)
	require.NotNil(t, cfg.ContainerEngine)
	require.Equal(t, "edgelet", *cfg.ContainerEngine)
}
