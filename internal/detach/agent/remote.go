package detachagent

import (
	"fmt"

	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe executor) remoteDeprovision(agent *rsc.RemoteAgent) error {
	if agent.ValidateSSH() != nil {
		util.PrintNotify("Could not deprovision daemon for Agent " + agent.Name + ". SSH details missing from local configuration. Use configure command to add SSH details.")
		return nil
	}

	cfg := deployairgap.EdgeletInstallConfig("linux", agent.Config, agent.Package)
	edgelet, err := install.NewRemoteEdgelet(
		agent.SSH.User,
		agent.Host,
		agent.SSH.Port,
		agent.SSH.KeyFile,
		agent.Name,
		agent.UUID,
		cfg,
	)
	if err != nil {
		return err
	}
	if err := edgelet.Deprovision(); err != nil {
		util.PrintNotify(fmt.Sprintf("Failed to deprovision daemon on Agent %s. %s", agent.Name, err.Error()))
	}
	if err := edgelet.Uninstall(false); err != nil {
		util.PrintNotify(fmt.Sprintf("Failed to uninstall edgelet on Agent %s. %s", agent.Name, err.Error()))
	}
	return nil
}
