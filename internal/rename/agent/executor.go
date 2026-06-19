package agent

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace, name, newName string, useDetached bool) error {
	if err := util.IsLowerAlphanumeric("Agent", newName); err != nil {
		return err
	}
	util.SpinStart(fmt.Sprintf("Renaming Agent %s", name))

	if useDetached {
		if err := config.RenameDetachedAgent(name, newName); err != nil {
			return err
		}
		return config.Flush()
	}

	// Get config
	// Update local cache based on Controller
	if err := clientutil.SyncAgentInfo(namespace); err != nil {
		return err
	}
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}
	agent, err := ns.GetAgent(name)
	if err != nil {
		return err
	}

	// Init remote resources
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	if _, err = clt.UpdateAgent(&client.AgentUpdateRequest{
		UUID: agent.GetUUID(),
		Name: newName,
	}); err != nil {
		return err
	}
	if err := ns.DeleteAgent(name); err != nil {
		return err
	}
	agent.SetName(newName)
	if err := ns.AddAgent(agent); err != nil {
		return err
	}

	return config.Flush()
}
