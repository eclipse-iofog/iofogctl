package detachexecagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Name      string
	Namespace string
}

type executor struct {
	name      string
	namespace string
}

func NewExecutor(opt Options) execute.Executor {
	return &executor{
		name:      opt.Name,
		namespace: opt.Namespace,
	}
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) Execute() error {
	util.SpinStart("Removing debug exec from Agent")

	// Init client
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	agent, err := clt.GetAgentByName(exe.name)
	if err != nil {
		return fmt.Errorf("failed to get Agent by name: %w", err)
	}

	req := client.DetachExecFromAgentRequest{
		UUID: agent.UUID,
	}
	err = clt.DetachExecFromAgent(&req)
	if err != nil {
		return fmt.Errorf("failed to detach Exec Session from Agent: %w", err)
	}

	return nil
}
