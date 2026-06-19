package logs

import (
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type localControllerExecutor struct {
	controlPlane *rsc.LocalControlPlane
	namespace    string
	name         string
}

func newLocalControllerExecutor(controlPlane *rsc.LocalControlPlane, namespace, name string) *localControllerExecutor {
	return &localControllerExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
		name:         name,
	}
}

func (exe *localControllerExecutor) GetName() string {
	return exe.name
}

func (exe *localControllerExecutor) Execute() error {
	lc, err := install.NewLocalContainerClient()
	if err != nil {
		return err
	}
	containerName := install.GetLocalContainerName("controller", false)
	stdout, stderr, err := lc.GetLogsByName(containerName)
	if err != nil {
		return err
	}

	printContainerLogs(stdout, stderr)

	return nil
}
