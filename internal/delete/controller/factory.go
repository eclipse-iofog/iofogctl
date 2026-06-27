package deletecontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func NewExecutor(namespace, name string) (execute.Executor, error) {
	// Get controller from config
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}
	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		return nil, err
	}
	switch controlPlane := baseControlPlane.(type) {
	case *rsc.KubernetesControlPlane:
		return nil, util.NewInputError("Cannot delete Kubernetes Controller, delete the Control Plane instead.")
	case *rsc.RemoteControlPlane:
		return NewRemoteExecutor(controlPlane, namespace, name), nil
	case *rsc.LocalControlPlane:
		return NewLocalExecutor(controlPlane, namespace, name), nil
	}

	return nil, util.NewInternalError("Could not determine what kind of Control Plane is in Namespace " + namespace)
}
