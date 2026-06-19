package logs

import (
	"fmt"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type kubernetesControllerExecutor struct {
	controlPlane *rsc.KubernetesControlPlane
	namespace    string
	name         string
}

func newKubernetesControllerExecutor(controlPlane *rsc.KubernetesControlPlane, namespace, name string) *kubernetesControllerExecutor {
	return &kubernetesControllerExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
		name:         name,
	}
}

func (exe *kubernetesControllerExecutor) GetName() string {
	return exe.name
}

func (exe *kubernetesControllerExecutor) Execute() error {
	if err := exe.controlPlane.ValidateKubeConfig(); err != nil {
		return err
	}
	out, err := util.Exec("KUBECONFIG="+exe.controlPlane.KubeConfig, "kubectl", "logs", "-l", "name=controller", "-n", exe.namespace)
	if err != nil {
		return err
	}
	fmt.Print(out.String())

	return nil
}
