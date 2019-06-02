package deletecontroller

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
	"github.com/eclipse-iofog/cli/pkg/iofog"
)

type kubernetesExecutor struct {
	namespace string
	name      string
}

func newKubernetesExecutor(namespace, name string) *kubernetesExecutor {
	exe := &kubernetesExecutor{}
	exe.namespace = namespace
	exe.name = name
	return exe
}

func (exe *kubernetesExecutor) Execute() error {
	// Find the requested controller
	ctrl, err := config.GetController(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	// Instantiate Kubernetes object
	k8s, err := iofog.NewKubernetes(ctrl.KubeConfig)

	// Delete Controller on cluster
	err = k8s.DeleteController()
	if err != nil {
		return err
	}

	// Update configuration
	err = config.DeleteController(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	fmt.Printf("\nController %s/%s successfully deleted.\n", exe.namespace, exe.name)
	return nil
}
