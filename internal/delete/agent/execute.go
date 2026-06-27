package deleteagent

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type executor struct {
	name        string
	namespace   string
	useDetached bool
	force       bool
}

func NewExecutor(namespace, name string, useDetached, force bool) (execute.Executor, error) {
	return executor{name: name, namespace: namespace, useDetached: useDetached, force: force}, nil
}

func (exe executor) GetName() string {
	return exe.name
}

func (exe executor) Execute() (err error) {
	util.SpinStart("Deleting Agent")

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}

	var baseAgent rsc.Agent

	// Detached from config
	if exe.useDetached {
		baseAgent, err = config.GetDetachedAgent(exe.name)
		if err != nil {
			return err
		}

		// Update config
		if err := config.DeleteDetachedAgent(baseAgent.GetName()); err != nil {
			return err
		}
		return config.Flush()
	}

	// Update Agent cache
	if err := clientutil.SyncAgentInfo(exe.namespace); err != nil {
		return err
	}

	baseAgent, err = ns.GetAgent(exe.name)
	if err != nil {
		if util.IsNotFoundError(err) {
			clientutil.InvalidateAgentCache(exe.namespace)
			backendAgents, backendErr := clientutil.GetBackendAgents(exe.namespace)
			if backendErr == nil && !agentListedInBackend(exe.name, backendAgents) {
				return nil
			}
		}
		return err
	}

	// Check if it has microservices running on it
	if !exe.force {
		if err := exe.checkMicroservices(baseAgent.GetName(), baseAgent.GetUUID()); err != nil {
			return err
		}
	}

	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		install.Verbose("Deprovisioning edgelet on local agent " + agent.GetName())
		if err = exe.deprovisionLocalEdgelet(agent); err != nil {
			util.PrintInfo(fmt.Sprintf("Could not deprovision Agent on the local host %s. Error: %s\n", agent.GetHost(), err.Error()))
		}
	case *rsc.RemoteAgent:
		install.Verbose("Deprovisioning edgelet on remote agent " + agent.GetName())
		if err = exe.deprovisionRemoteEdgelet(agent); err != nil {
			util.PrintInfo(fmt.Sprintf("Could not deprovision Agent on the remote host %s. Error: %s\n", agent.GetHost(), err.Error()))
		}
	}

	// Delete from Controller while it is still reachable after deprovision.
	ctrl, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		util.PrintInfo(fmt.Sprintf("Could not delete Agent %s from the Controller. Error: %s\n", exe.name, err.Error()))
	} else if baseAgent.GetUUID() != "" {
		if err := ctrl.DeleteAgent(baseAgent.GetUUID()); err != nil && !isControllerAgentNotFound(err) {
			return err
		}
	}

	clientutil.InvalidateAgentCache(exe.namespace)

	// Remove edgelet from host
	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		install.Verbose("Uninstalling edgelet from local agent " + agent.GetName())
		if err = exe.uninstallLocalEdgelet(agent); err != nil {
			util.PrintInfo(fmt.Sprintf("Could not remove Agent from the local host %s. Error: %s\n", agent.GetHost(), err.Error()))
		}
	case *rsc.RemoteAgent:
		install.Verbose("Uninstalling edgelet from remote agent " + agent.GetName())
		if err = exe.uninstallRemoteEdgelet(agent); err != nil {
			util.PrintInfo(fmt.Sprintf("Could not remove Agent from the remote host %s. Error: %s\n", agent.GetHost(), err.Error()))
		}
	}

	if err := ns.DeleteAgent(baseAgent.GetName()); err != nil {
		return err
	}

	// Update and/or Delete Volumes pertaining to deleted Agent
	vols := ns.GetVolumes()
	var rmVols []rsc.Volume
	var updateVols []rsc.Volume
	for _, vol := range vols {
		for idx, volAgent := range vol.Agents {
			if volAgent == baseAgent.GetName() {
				if len(vol.Agents) == 1 {
					// Remove the Volume
					rmVols = append(rmVols, vol)
				} else {
					// Remove the Agent from Volume
					vol.Agents = append(vol.Agents[:idx], vol.Agents[idx+1:]...)
					updateVols = append(updateVols, vol)
				}
				break
			}
		}
	}
	for idx := range rmVols {
		if err := ns.DeleteVolume(rmVols[idx].Name); err != nil {
			util.PrintInfo(fmt.Sprintf("Could not delete Volume %s", rmVols[idx].Name))
		}
	}
	for idx := range updateVols {
		ns.UpdateVolume(&updateVols[idx])
	}

	return config.Flush()
}

func agentListedInBackend(name string, agents []client.AgentInfo) bool {
	for idx := range agents {
		if agents[idx].Name == name {
			return true
		}
	}
	return false
}

func isControllerAgentNotFound(err error) bool {
	if err == nil {
		return false
	}
	if util.IsNotFoundError(err) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "notfound") || strings.Contains(msg, "not found")
}

func (exe executor) checkMicroservices(agentName, agentUUID string) (err error) {
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
		if msvc.AgentUUID == agentUUID {
			msg := "Could not delete Agent %s because it still has microservices running. Remove the microservices first, or use the --force option."
			return util.NewInputError(fmt.Sprintf(msg, agentName))
		}
	}
	return
}
