package deletecontroller

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type LocalExecutor struct {
	controlPlane          *rsc.LocalControlPlane
	namespace             string
	name                  string
	localControllerConfig *install.LocalContainerConfig
}

func NewLocalExecutor(controlPlane *rsc.LocalControlPlane, namespace, name string) *LocalExecutor {
	exe := &LocalExecutor{
		controlPlane:          controlPlane,
		namespace:             namespace,
		name:                  name,
		localControllerConfig: install.NewLocalControllerConfig("", install.Credentials{}, install.Auth{}, install.Database{}, install.Events{}, nil),
	}
	return exe
}

func (exe *LocalExecutor) GetName() string {
	return exe.name
}

func (exe *LocalExecutor) Execute() error {
	if err := exe.deleteLegacyControllerContainer(); err != nil {
		util.PrintNotify(fmt.Sprintf("Could not clean Controller container: %v", err))
	}

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	if err := ns.DeleteController(exe.name); err != nil {
		return err
	}
	ns.SetControlPlane(exe.controlPlane)
	return config.Flush()
}

func (exe *LocalExecutor) deleteLegacyControllerContainer() error {
	client, err := install.NewLocalContainerClient(install.DefaultLocalContainerEngine, nil)
	if err != nil {
		return err
	}
	return client.CleanContainer(exe.localControllerConfig.ContainerName)
}
