package agents

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
)

func Connect(ctrlPlane rsc.ControlPlane, endpoint string, ns *rsc.Namespace) error {
	// Connect to Controller
	ctrl, err := client.NewAndLogin(client.Options{Endpoint: endpoint}, ctrlPlane.GetUser().Email, ctrlPlane.GetUser().Password)
	if err != nil {
		return err
	}

	// Get Agents
	listAgentsResponse, err := ctrl.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return err
	}

	// Update Agents config
	for _, agent := range listAgentsResponse.Agents {
		agentConfig := rsc.RemoteAgent{
			Name: agent.Name,
			UUID: agent.UUID,
			Host: agent.IPAddressExternal,
		}
		if err = ns.AddAgent(&agentConfig); err != nil {
			return err
		}
	}
	return nil
}
