package detachedgeresource

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type executor struct {
	name      string
	version   string
	namespace string
	agent     string
}

func NewExecutor(namespace, name, version, agent string) execute.Executor {
	return executor{name: name,
		version:   version,
		namespace: namespace,
		agent:     agent}
}

func (exe executor) GetName() string {
	return fmt.Sprintf("%s/%s", exe.name, exe.version)
}

func (exe executor) Execute() error {
	util.SpinStart("Detaching Edge Resource")

	// Init client
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Get Agent UUID
	// agentInfo, err := clt.GetAgentByName(exe.agent, false)
	agentInfo, err := clt.GetAgentByName(exe.agent)
	if err != nil {
		return err
	}
	// Detach from agent
	req := client.LinkEdgeResourceRequest{
		AgentUUID:           agentInfo.UUID,
		EdgeResourceName:    exe.name,
		EdgeResourceVersion: exe.version,
	}
	if err := clt.UnlinkEdgeResource(req); err != nil {
		return err
	}

	return nil
}
