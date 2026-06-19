package detachvolumemount

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Name      string
	Agents    []string
	Namespace string
}

type executor struct {
	name      string
	agents    []string
	namespace string
}

func NewExecutor(opt Options) execute.Executor {
	return &executor{
		name:      opt.Name,
		agents:    opt.Agents,
		namespace: opt.Namespace,
	}
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) Execute() error {
	util.SpinStart("Detaching Volume Mount")

	// Init client
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Get Agents UUID
	var agentUUIDs []string
	for _, agent := range exe.agents {
		agentInfo, err := clt.GetAgentByName(agent)
		if err != nil {
			return err
		}
		agentUUIDs = append(agentUUIDs, agentInfo.UUID)
	}

	// Detach from agent
	req := client.VolumeMountUnlinkRequest{
		Name:     exe.name,
		FogUUIDs: agentUUIDs,
	}
	if err := clt.UnlinkVolumeMount(&req); err != nil {
		return err
	}

	return nil
}
