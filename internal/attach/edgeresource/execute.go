package attachedgeresource

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Name      string
	Version   string
	Agent     string
	Namespace string
}

type executor struct {
	Options
}

func NewExecutor(opt Options) execute.Executor {
	return executor{opt}
}

func (exe executor) GetName() string {
	return fmt.Sprintf("%s/%s", exe.Name, exe.Version)
}

func (exe executor) Execute() error {
	util.SpinStart("Attaching Edge Resource")

	// Init client
	clt, err := clientutil.NewControllerClient(exe.Namespace)
	if err != nil {
		return err
	}

	// Get Agent UUID
	// agentInfo, err := clt.GetAgentByName(exe.Agent, false)
	agentInfo, err := clt.GetAgentByName(exe.Agent)
	if err != nil {
		return err
	}
	// Attach to agent
	req := client.LinkEdgeResourceRequest{
		AgentUUID:           agentInfo.UUID,
		EdgeResourceName:    exe.Name,
		EdgeResourceVersion: exe.Version,
	}
	if err := clt.LinkEdgeResource(req); err != nil {
		return err
	}

	return nil
}
