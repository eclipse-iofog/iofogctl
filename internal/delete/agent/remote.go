package deleteagent

import (
	"fmt"

	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe executor) deleteRemoteAgent(agent *rsc.RemoteAgent) error {
	if agent.ValidateSSH() != nil {
		util.PrintNotify("Could not stop daemon for Agent " + agent.Name + ". SSH details missing from local cofiguration. Use configure command to add SSH details.")
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
	if err := edgelet.Uninstall(true); err != nil {
		util.PrintNotify(fmt.Sprintf("Failed to stop daemon on Agent %s. %s", agent.Name, err.Error()))
	}
	return nil
}
