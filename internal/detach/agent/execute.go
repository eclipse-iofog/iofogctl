package detachagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type executor struct {
	name      string
	force     bool
	namespace string
}

func NewExecutor(namespace, name string, force bool) execute.Executor {
	return executor{name: name, namespace: namespace, force: force}
}

func (exe executor) GetName() string {
	return exe.name
}

func (exe executor) Execute() error {
	util.SpinStart("Detaching Agent")

	// Check doesn't already exist with same name
	if _, err := config.GetDetachedAgent(exe.name); err == nil {
		return util.NewConflictError(fmt.Sprintf(
			"An Agent with the name '%s' is already detached. Detach or delete the existing detached Agent before trying again.",
			exe.name,
		))
	}

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}

	// Update local cache based on Controller
	if err := clientutil.SyncAgentInfo(exe.namespace); err != nil {
		return err
	}
	baseAgent, err := ns.GetAgent(exe.name)
	if err != nil {
		return err
	}

	// Check if it has microservices running on it
	if !exe.force {
		// Try to get a Controller client to talk to the REST API
		ctrl, err := clientutil.NewControllerClient(exe.namespace)
		if err != nil {
			return err
		}
		msvcList, err := ctrl.GetAllMicroservices()
		if err != nil {
			return err
		}
		for idx := range msvcList.Microservices {
			msvc := &msvcList.Microservices[idx]
			if msvc.AgentUUID == baseAgent.GetUUID() {
				msg := "Could not detach Agent %s because it still has microservices running. Remove the microservices first, or use the --force option."
				return util.NewInputError(fmt.Sprintf(msg, baseAgent.GetName()))
			}
		}
	}

	// Deprovision agent
	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		cfg := deployairgap.EdgeletInstallConfig(deployairgap.LocalEdgeletHostOS(), agent.Config, agent.Package)
		if err := exe.localDeprovision(agent.Name, agent.UUID, cfg); err != nil {
			return err
		}
	case *rsc.RemoteAgent:
		if err := exe.remoteDeprovision(agent); err != nil {
			return err
		}
	}

	// Try to get a Controller client to talk to the REST API
	ctrl, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Get Config before deletion
	agentConfig, _, _, err := clientutil.GetAgentConfig(exe.name, exe.namespace)
	if err != nil {
		return err
	}

	// Get UUID for deletion
	// agentInfo, err := ctrl.GetAgentByName(exe.name, false)
	agentInfo, err := ctrl.GetAgentByName(exe.name)
	if err != nil {
		return err
	}

	// Perform deletion of Agent through Controller
	if err := ctrl.DeleteAgent(agentInfo.UUID); err != nil {
		return err
	}

	// Try to detach from config
	if err := config.DetachAgent(exe.namespace, exe.name); err != nil {
		return err
	}

	// Update detached Agent with Agent Config
	agent, err := config.GetDetachedAgent(exe.name)
	if err != nil {
		return err
	}
	agent.SetConfig(&agentConfig)
	if err := config.UpdateDetachedAgent(agent); err != nil {
		return err
	}

	return config.Flush()
}
