package deploycontroller

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
	"github.com/eclipse-iofog/cli/pkg/iofog"
	"github.com/eclipse-iofog/cli/pkg/util"
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

	// Generate a password
	password := util.RandomString(10)

	// Create controller on cluster
	user := iofog.User{
		Name:     "Automated",
		Surname:  "Deployment",
		Email:    "user@domain.com",
		Password: password,
	}
	endpoint, err := k8s.CreateController(user)
	if err != nil {
		return
	}

	// Update configuration
	configEntry := config.Controller{
		Name:       exe.opt.Name,
		KubeConfig: exe.opt.KubeConfig,
		Endpoint:   endpoint,
		IofogUser: config.IofogUser{
			Name:     user.Name,
			Surname:  user.Surname,
			Email:    user.Email,
			Password: password,
		},
	}
	err = config.AddController(exe.opt.Namespace, configEntry)
	if err != nil {
		return
	}

	fmt.Printf("\nController %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)
	return nil
}
