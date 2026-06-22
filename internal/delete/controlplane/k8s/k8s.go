package deletek8scontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/internal/trust"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor struct {
	namespace       string
	deleteNamespace bool
}

func NewExecutor(namespace string, deleteNamespace bool) (execute.Executor, error) {
	exe := &Executor{
		namespace:       namespace,
		deleteNamespace: deleteNamespace,
	}
	return exe, nil
}

func (exe *Executor) GetName() string {
	return "Delete Control Plane"
}

func (exe *Executor) Execute() (err error) {
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}

	controlPlane, ok := baseControlPlane.(*rsc.KubernetesControlPlane)
	if !ok {
		return util.NewError("Could not convert Control Plane to Kubernetes Control Plane")
	}

	k8s, err := install.NewKubernetes(controlPlane.KubeConfig, exe.namespace)
	if err != nil {
		return err
	}

	if err = k8s.DeleteControlPlane(exe.deleteNamespace); err != nil {
		return err
	}

	if err = trust.RemoveCA(exe.namespace); err != nil {
		return err
	}

	ns.DeleteControlPlane()
	return config.Flush()
}
