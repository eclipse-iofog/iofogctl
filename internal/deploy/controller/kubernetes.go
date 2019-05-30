package deploycontroller

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
	"github.com/eclipse-iofog/cli/pkg/iofog"
	"github.com/eclipse-iofog/cli/pkg/util"
)

type kubernetesExecutor struct {
	configManager *config.Manager
	opt           *Options
}

func newKubernetesExecutor(opt *Options) *kubernetesExecutor {
	k := &kubernetesExecutor{}
	k.configManager = config.NewManager()
	k.opt = opt
	return k
}

func (exe *kubernetesExecutor) Execute(namespace, name string) (err error) {
	// Check controller already exists
	_, err = exe.configManager.GetController(namespace, name)
	if err == nil {
		return util.NewConflictError(namespace + "/" + name)
	}

	// Update configuration
	configEntry := config.Controller{
		Name:       name,
		KubeConfig: exe.opt.KubeConfig,
	}
	err = exe.configManager.AddController(namespace, configEntry)
	if err != nil {
		return
	}

	// Get Kubernetes cluster
	k8s, err := iofog.NewKubernetes(exe.opt.KubeConfig)
	if err != nil {
		return
	}

	// Initialize the cluster
	err = k8s.Init()
	if err != nil {
		return
	}

	// Create controller on cluster
	err = k8s.CreateController()
	if err != nil {
		return
	}

	fmt.Printf("\nController %s/%s successfully deployed.\n", namespace, name)
	return nil
}
