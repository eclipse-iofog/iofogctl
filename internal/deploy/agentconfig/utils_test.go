package deployagentconfig

import (
	"encoding/json"
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"
	"github.com/stretchr/testify/require"
)

func TestGetAgentUpdateRequestPreservesArchID(t *testing.T) {
	arch := "arm64"
	cfg := &rsc.AgentConfiguration{
		Name: "remote-1",
		Arch: &arch,
		AgentConfiguration: client.AgentConfiguration{
			Host: iutil.MakeStrPtr("192.168.1.1"),
		},
	}
	PrepareForControllerAPI(cfg)

	req := getAgentUpdateRequestFromAgentConfig(cfg, nil)
	require.NotNil(t, req.ArchID)
	require.Equal(t, int64(2), *req.ArchID)

	body, err := json.Marshal(&client.CreateAgentRequest{AgentUpdateRequest: req})
	require.NoError(t, err)
	require.Contains(t, string(body), `"archId":2`)
}

func TestPrepareForControllerAPISystemAgentDefaults(t *testing.T) {
	cfg := &rsc.AgentConfiguration{
		Name: "cp-system",
		AgentConfiguration: client.AgentConfiguration{
			IsSystem: iutil.MakeBoolPtr(true),
		},
	}
	PrepareForControllerAPI(cfg)

	require.NotNil(t, cfg.RouterMode)
	require.Equal(t, "interior", *cfg.RouterMode)
	require.NotNil(t, cfg.NatsMode)
	require.Equal(t, "server", *cfg.NatsMode)
	require.NotNil(t, cfg.MessagingPort)
	require.Equal(t, DefaultMessagingPort, *cfg.MessagingPort)
	require.NotNil(t, cfg.NatsClusterPort)
	require.Equal(t, DefaultNatsClusterPort, *cfg.NatsClusterPort)
}

func TestPrepareForControllerAPIEdgeAgentDefaults(t *testing.T) {
	cfg := &rsc.AgentConfiguration{Name: "edge-1"}
	PrepareForControllerAPI(cfg)

	require.NotNil(t, cfg.RouterMode)
	require.Equal(t, "edge", *cfg.RouterMode)
	require.NotNil(t, cfg.NatsMode)
	require.Equal(t, "leaf", *cfg.NatsMode)
	require.NotNil(t, cfg.NatsLeafPort)
	require.Equal(t, DefaultNatsLeafPort, *cfg.NatsLeafPort)
}

func TestValidateSystemAgentRejectsEdgeRouter(t *testing.T) {
	edge := "edge"
	cfg := &rsc.AgentConfiguration{
		Name: "sys",
		AgentConfiguration: client.AgentConfiguration{
			IsSystem: iutil.MakeBoolPtr(true),
			RouterConfig: client.RouterConfig{
				RouterMode: &edge,
			},
		},
	}
	err := Validate(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "routerMode must be interior")
}

func TestValidateEdgeAgentAllowsOmittedModes(t *testing.T) {
	cfg := &rsc.AgentConfiguration{Name: "edge-1"}
	require.NoError(t, Validate(cfg))
}
