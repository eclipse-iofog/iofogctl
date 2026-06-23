package deleteagent

import (
	"fmt"

	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe executor) newRemoteEdgelet(agent *rsc.RemoteAgent) (*install.RemoteEdgelet, error) {
	cfg := deployairgap.EdgeletInstallConfig("linux", agent.Config, agent.Package)
	return install.NewRemoteEdgelet(
		agent.SSH.User,
		agent.Host,
		agent.SSH.Port,
		agent.SSH.KeyFile,
		agent.Name,
		agent.UUID,
		cfg,
	)
}

func (exe executor) deprovisionRemoteEdgelet(agent *rsc.RemoteAgent) error {
	if agent.ValidateSSH() != nil {
		util.PrintNotify("Could not deprovision daemon for Agent " + agent.Name + ". SSH details missing from local configuration. Use configure command to add SSH details.")
		return nil
	}
	edgelet, err := exe.newRemoteEdgelet(agent)
	if err != nil {
		return err
	}
	if err := edgelet.Deprovision(); err != nil {
		util.PrintNotify(fmt.Sprintf("Could not deprovision edgelet on Agent %s: %v", agent.Name, err))
	}
	return nil
}

func (exe executor) uninstallRemoteEdgelet(agent *rsc.RemoteAgent) error {
	if agent.ValidateSSH() != nil {
		util.PrintNotify("Could not stop daemon for Agent " + agent.Name + ". SSH details missing from local configuration. Use configure command to add SSH details.")
		return nil
	}
	edgelet, err := exe.newRemoteEdgelet(agent)
	if err != nil {
		return err
	}
	if err := edgelet.Uninstall(true); err != nil {
		util.PrintNotify(fmt.Sprintf("Failed to stop daemon on Agent %s. %s", agent.Name, err.Error()))
	}
	return nil
}

func (exe executor) deleteRemoteAgent(agent *rsc.RemoteAgent) error {
	if err := exe.deprovisionRemoteEdgelet(agent); err != nil {
		return err
	}
	return exe.uninstallRemoteEdgelet(agent)
}
