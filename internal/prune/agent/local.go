package pruneagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe executor) localAgentPrune() error {
	containerClient, err := install.NewLocalContainerClient()
	if err != nil {
		return err
	}
	if _, err = containerClient.ExecuteCmd(install.GetLocalContainerName("agent", false), []string{
		"sudo",
		"iofog-agent",
		"prune",
	}); err != nil {
		return util.NewInternalError(fmt.Sprintf("Could not prune local agent. Error: %s\n", err.Error()))
	}

	return nil
}
