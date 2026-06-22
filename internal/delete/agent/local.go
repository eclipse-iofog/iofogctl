package deleteagent

import (
	"fmt"
	"strings"

	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe executor) deleteLocalEdgelet(agent *rsc.LocalAgent) error {
	cfg := deployairgap.EdgeletInstallConfig(deployairgap.LocalEdgeletHostOS(), agent.Config, agent.Package)
	edgelet, err := install.NewLocalEdgelet(agent.Name, agent.UUID, cfg)
	if err != nil {
		return err
	}
	if err := edgelet.Uninstall(true); err != nil {
		util.PrintNotify(fmt.Sprintf("Could not remove edgelet from local host: %v", err))
	}

	client, err := install.NewLocalContainerClientFromEdgeletCfg(cfg)
	if err != nil {
		return err
	}
	containers, err := client.ListContainers()
	if err != nil {
		return err
	}
	for idx := range containers {
		container := &containers[idx]
		for _, containerName := range container.Names {
			if strings.HasPrefix(containerName, "/iofog_") {
				if errClean := client.CleanContainerByID(container.ID); errClean != nil {
					util.PrintNotify(fmt.Sprintf("Could not clean Microservice container: %v", errClean))
				}
			}
		}
	}
	return nil
}
