package deployremotecontrolplane

import (
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"
	"github.com/stretchr/testify/require"
)

func TestResolveControllerHostEndpointPrefersPublicURL(t *testing.T) {
	cp := &rsc.RemoteControlPlane{
		Controller: rsc.LocalControllerSpec{
			ControllerConfig: rsc.ControllerConfig{
				PublicUrl: "http://192.168.139.85:51121",
			},
		},
	}
	ctrl := &rsc.RemoteController{Name: "remote-1", Host: "0.0.0.0"}

	endpoint, err := ResolveControllerHostEndpoint(cp, ctrl)
	require.NoError(t, err)
	require.Equal(t, "http://192.168.139.85:51121", endpoint)
}

func TestResolveControllerHostEndpointUsesSystemAgentHostWhenSSHHostIsBindAddress(t *testing.T) {
	cp := &rsc.RemoteControlPlane{}
	ctrl := &rsc.RemoteController{
		Name: "remote-1",
		Host: "0.0.0.0",
		SystemAgent: &rsc.SystemAgentConfig{
			AgentConfiguration: &rsc.AgentConfiguration{
				AgentConfiguration: client.AgentConfiguration{
					Host: iutil.MakeStrPtr("192.168.139.85"),
				},
			},
		},
	}

	endpoint, err := ResolveControllerHostEndpoint(cp, ctrl)
	require.NoError(t, err)
	require.Equal(t, "http://192.168.139.85:51121", endpoint)
}

func TestResolveControllerHostEndpointUsesRoutableControllerHost(t *testing.T) {
	cp := &rsc.RemoteControlPlane{}
	ctrl := &rsc.RemoteController{Name: "remote-1", Host: "10.0.0.5"}

	endpoint, err := ResolveControllerHostEndpoint(cp, ctrl)
	require.NoError(t, err)
	require.Equal(t, "http://10.0.0.5:51121", endpoint)
}

func TestIsNonRoutableControllerHost(t *testing.T) {
	require.True(t, isNonRoutableControllerHost("0.0.0.0"))
	require.True(t, isNonRoutableControllerHost("127.0.0.1"))
	require.True(t, isNonRoutableControllerHost("localhost"))
	require.False(t, isNonRoutableControllerHost("192.168.139.85"))
}
