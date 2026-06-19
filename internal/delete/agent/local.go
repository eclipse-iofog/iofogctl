package deleteagent

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/util"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

func (exe executor) deleteLocalContainer() error {
	client, err := install.NewLocalContainerClient()
	if err != nil {
		return err
	}

	// Clean agent containers (normal and system)
	if errClean := client.CleanContainer(install.GetLocalContainerName("agent", false)); errClean != nil {
		util.PrintNotify(fmt.Sprintf("Could not clean Agent container: %v", errClean))
	}

	// Clean microservices
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
