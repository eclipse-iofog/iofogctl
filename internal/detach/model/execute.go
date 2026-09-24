package detachmodel

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
	util.SpinStart("Detaching Model")

	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	agentUUIDs, err := clientutil.AgentUUIDsFromNames(clt, exe.agents)
	if err != nil {
		return err
	}

	req := client.FogLinkRequest{FogUUIDs: agentUUIDs}
	return clt.UnlinkModel(exe.name, &req)
}
