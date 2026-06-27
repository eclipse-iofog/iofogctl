package deletecontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
)

type LocalExecutor struct {
	controlPlane *rsc.LocalControlPlane
	namespace    string
	name         string
}

func NewLocalExecutor(controlPlane *rsc.LocalControlPlane, namespace, name string) *LocalExecutor {
	return &LocalExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
		name:         name,
	}
}

func (exe *LocalExecutor) GetName() string {
	return exe.name
}

func (exe *LocalExecutor) Execute() error {
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
