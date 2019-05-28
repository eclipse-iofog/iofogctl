package deploycontroller

import (
	"fmt"
	"github.com/eclipse-iofog/cli/pkg/config"
)

type kubernetesExecutor struct {
	configManager *config.Manager
	opt *options
}

func newKubernetesExecutor(opt *options) *kubernetesExecutor {
	k := &kubernetesExecutor{}
	k.configManager = config.NewManager()
	k.opt = opt
	return k
}

func (exe *kubernetesExecutor) execute(namespace, name string) error {
	// TODO (Serge) Execute back-end logic

	// Update configuration
	configEntry := config.Controller{ 
		Name: name, 
		KubeConfig: *exe.opt.kubeConfig,
	}
	err := exe.configManager.AddController(namespace, configEntry)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nController %s/%s successfully deployed.\n", namespace, name)
	}
	return err
}