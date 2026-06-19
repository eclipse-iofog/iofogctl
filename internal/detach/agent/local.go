package detachagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe executor) localDeprovision() error {
	containerClient, err := install.NewLocalContainerClient()
	if err != nil {
		util.PrintNotify(fmt.Sprintf("Could not deprovision local iofog-agent container. Error: %s\n", err.Error()))
	} else if _, err = containerClient.ExecuteCmd(install.GetLocalContainerName("agent", false), []string{
		"sudo",
		"iofog-agent",
		"deprovision",
	}); err != nil {
		util.PrintNotify(fmt.Sprintf("Could not deprovision local iofog-agent container. Error: %s\n", err.Error()))
	}
	return nil
}
