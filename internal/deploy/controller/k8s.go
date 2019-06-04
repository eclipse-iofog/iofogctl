package deploycontroller

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
	"github.com/eclipse-iofog/cli/pkg/iofog"
)

type kubernetesExecutor struct {
	opt *Options
}

func newKubernetesExecutor(opt *Options) *kubernetesExecutor {
	k := &kubernetesExecutor{}
	k.opt = opt
	return k
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Get Kubernetes cluster
	k8s, err := iofog.NewKubernetes(exe.opt.KubeConfig)
	if err != nil {
		return
	}

	// Create controller on cluster
	endpoint, err := k8s.CreateController()
	if err != nil {
		return
	}

	// Update configuration
	configEntry := config.Controller{
		Name:       exe.opt.Name,
		KubeConfig: exe.opt.KubeConfig,
		Endpoint:   endpoint,
	}
	err = config.AddController(exe.opt.Namespace, configEntry)
	if err != nil {
		return
	}

	fmt.Printf("\nController %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)
	return nil
}
