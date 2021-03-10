package deletecontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	deletek8scontrolplane "github.com/eclipse-iofog/iofogctl/v3/internal/delete/controlplane/k8s"
	deletelocalcontrolplane "github.com/eclipse-iofog/iofogctl/v3/internal/delete/controlplane/local"
	deleteremotecontrolplane "github.com/eclipse-iofog/iofogctl/v3/internal/delete/controlplane/remote"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

func NewExecutor(namespace string) (execute.Executor, error) {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}
	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		return nil, err
	}
	switch baseControlPlane.(type) {
	case *rsc.KubernetesControlPlane:
		return deletek8scontrolplane.NewExecutor(namespace)
	case *rsc.RemoteControlPlane:
		return deleteremotecontrolplane.NewExecutor(namespace)
	case *rsc.LocalControlPlane:
		return deletelocalcontrolplane.NewExecutor(namespace)
	}
	return nil, util.NewError("Could not convert Control Plane to dynamic type")
}
