package deploycontroller

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
	"github.com/eclipse-iofog/cli/pkg/iofog"
	"github.com/eclipse-iofog/cli/pkg/util"
)

type kubernetesExecutor struct {
	configManager *config.Manager
	opt           *options
}

func newKubernetesExecutor(opt *options) *kubernetesExecutor {
	k := &kubernetesExecutor{}
	k.configManager = config.NewManager()
	k.opt = opt
	return k
}

func (exe *kubernetesExecutor) execute(namespace, name string) (err error) {
	// Check controller already exists
	_, err = exe.configManager.GetController(namespace, name)
	if err != nil && err.Error() != util.NewNotFoundError(namespace + "/" + name).Error(){
		return
	}

	k8s := iofog.NewKubernetes(*exe.opt.kubeConfig)
	err = k8s.Init()
	if err != nil {
		return
	}

	// Update configuration
	configEntry := config.Controller{
		Name:       name,
		KubeConfig: *exe.opt.kubeConfig,
	}
	err = exe.configManager.AddController(namespace, configEntry)

	// TODO (Serge) Handle config file error, retry..?

	if err == nil {
		fmt.Printf("\nController %s/%s successfully deployed.\n", namespace, name)
	}
	return err
}
