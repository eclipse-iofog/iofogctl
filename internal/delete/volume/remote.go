package deletevolume

import (
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func deleteRemote(agent *rsc.RemoteAgent, volume *rsc.Volume) error {
	// Check SSH details
	if err := agent.ValidateSSH(); err != nil {
		return err
	}
	// Check agent is not local
	if util.IsLocalHost(agent.Host) {
		return util.NewError("Volume deletion is not supported for local Agents")
	}
	// Connect
	ssh, err := util.NewSecureShellClient(agent.SSH.User, agent.Host, agent.SSH.KeyFile)
	if err != nil {
		return err
	}
	if err := ssh.Connect(); err != nil {
		return err
	}
	// Delete
	if _, err := ssh.Run("sudo -S rm -rf " + util.AddTrailingSlash(volume.Destination) + "*"); err != nil {
		return err
	}
	return nil
}
