package pruneagent

import (
	"fmt"
	"strings"

	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe executor) remoteAgentPrune(agent rsc.Agent) error {
	ctrl, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}
	if err = ctrl.PruneAgent(agent.GetUUID()); err != nil {
		if !strings.Contains(err.Error(), "NotFoundError") {
			return err
		}
	}
	return nil
}

func (exe executor) remoteDetachedAgentPrune(agent *rsc.RemoteAgent) error {
	if err := agent.ValidateSSH(); err != nil {
		return err
	}
	cfg := deployairgap.EdgeletInstallConfig("linux", agent.Config, agent.Package)
	edgelet, err := install.NewRemoteEdgelet(agent.SSH.User, agent.Host, agent.SSH.Port, agent.SSH.KeyFile, agent.Name, agent.UUID, cfg)
	if err != nil {
		return err
	}
	if err := edgelet.Prune(); err != nil {
		return util.NewInternalError(fmt.Sprintf("Failed to Prune edgelet resource %s. %s", agent.Name, err.Error()))
	}
	return nil
}
