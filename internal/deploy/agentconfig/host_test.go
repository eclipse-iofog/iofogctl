package deployagentconfig

import (
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"
	"github.com/stretchr/testify/require"
)

func TestResolveAgentAPIHost_UsesSpecHostWhenConfigHostUnset(t *testing.T) {
	cfg := &rsc.AgentConfiguration{
		AgentConfiguration: client.AgentConfiguration{
			ContainerEngine: iutil.MakeStrPtr("edgelet"),
		},
	}

	require.Equal(t, "34.175.100.1", ResolveAgentAPIHost("34.175.100.1", cfg))
}

func TestResolveAgentAPIHost_PrefersConfigHostWhenSet(t *testing.T) {
	cfg := &rsc.AgentConfiguration{
		AgentConfiguration: client.AgentConfiguration{
			Host: iutil.MakeStrPtr("10.0.0.5"),
		},
	}

	require.Equal(t, "10.0.0.5", ResolveAgentAPIHost("34.175.100.1", cfg))
}

func TestProcess_PreservesExplicitHostOverIPAddressExternal(t *testing.T) {
	interior := "interior"
	host := "34.175.100.1"
	cfg := &rsc.AgentConfiguration{
		AgentConfiguration: client.AgentConfiguration{
			Host: &host,
			RouterConfig: client.RouterConfig{
				RouterMode: &interior,
			},
		},
	}

	require.NoError(t, Process(cfg, "edge-2", "0.0.0.0", nil))
	require.NotNil(t, cfg.Host)
	require.Equal(t, "34.175.100.1", *cfg.Host)
}
