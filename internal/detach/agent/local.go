package detachagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe executor) localDeprovision(agentName, agentUUID string, cfg install.EdgeletInstallConfig) error {
	edgelet, err := install.NewLocalEdgelet(agentName, agentUUID, cfg)
	if err != nil {
		util.PrintNotify(fmt.Sprintf("Could not deprovision local edgelet. Error: %s\n", err.Error()))
		return nil
	}
	if err := edgelet.Deprovision(); err != nil {
		util.PrintNotify(fmt.Sprintf("Could not deprovision local edgelet. Error: %s\n", err.Error()))
	}
	if err := edgelet.Uninstall(false); err != nil {
		util.PrintNotify(fmt.Sprintf("Could not uninstall local edgelet. Error: %s\n", err.Error()))
	}
	return nil
}
