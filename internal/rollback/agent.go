package rollback

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/nodeversion"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type agentExecutor struct {
	namespace string
	name      string
	semver    string
}

func newAgentExecutor(opt Options) *agentExecutor {
	return &agentExecutor{
		namespace: opt.Namespace,
		name:      opt.Name,
		semver:    opt.Semver,
	}
}

func (exe *agentExecutor) GetName() string {
	return exe.name
}

func (exe *agentExecutor) Execute() error {
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	// Update local cache based on Controller
	if err := clientutil.SyncAgentInfo(exe.namespace); err != nil {
		return err
	}

	// Get the Agent to verify it exists
	agent, err := ns.GetAgent(exe.name)
	if err != nil {
		return err
	}

	// Talk to Controller
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	if err := clt.RollbackNode(agent.GetUUID(), nodeversion.SemverPtr(exe.semver)); err != nil {
		return nodeversion.MapError("rollback", err)
	}

	return nil
}
