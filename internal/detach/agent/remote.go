package detachagent

import (
	"fmt"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe executor) remoteDeprovision(agent *rsc.RemoteAgent) error {
	if agent.ValidateSSH() != nil {
		util.PrintNotify("Could not deprovision daemon for Agent " + agent.Name + ". SSH details missing from local configuration. Use configure command to add SSH details.")
	} else {
		sshAgent, err := install.NewRemoteAgent(
			agent.SSH.User,
			agent.Host,
			agent.SSH.Port,
			agent.SSH.KeyFile,
			agent.Name,
			agent.UUID)
		if err != nil {
			return err
		}
		if err := sshAgent.Deprovision(); err != nil {
			util.PrintNotify(fmt.Sprintf("Failed to deprovision daemon on Agent %s. %s", agent.Name, err.Error()))
		}
	}
	return nil
}
