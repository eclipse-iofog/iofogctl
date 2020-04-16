package agents

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
)

func Connect(ctrlPlane rsc.ControlPlane, endpoint, namespace string) error {
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
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}
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
	return config.Flush()
}
