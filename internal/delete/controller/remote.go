package deletecontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deleteremotecontrolplane "github.com/eclipse-iofog/iofogctl/internal/delete/controlplane/remote"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type RemoteExecutor struct {
	controlPlane *rsc.RemoteControlPlane
	namespace    string
	name         string
}

func NewRemoteExecutor(controlPlane *rsc.RemoteControlPlane, namespace, name string) *RemoteExecutor {
	return &RemoteExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
		name:         name,
	}
}

func (exe *RemoteExecutor) GetName() string {
	return exe.name
}

func (exe *RemoteExecutor) Execute() error {
	baseCtrl, err := exe.controlPlane.GetController(exe.name)
	if err != nil {
		return err
	}

	ctrl, ok := baseCtrl.(*rsc.RemoteController)
	if !ok {
		return util.NewInternalError("Could not assert Controller type to Remote Controller")
	}

	if err := deleteremotecontrolplane.TeardownRemoteControllerHost(exe.namespace, exe.controlPlane, ctrl); err != nil {
		return err
	}

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	_ = ns.DeleteAgent(exe.name)
	clientutil.InvalidateAgentCache(exe.namespace)
	if err := ns.DeleteController(exe.name); err != nil {
		return err
	}
	return config.Flush()
}
