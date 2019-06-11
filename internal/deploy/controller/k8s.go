package deploycontroller

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
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

	// Configure images
	k8s.SetImages(exe.opt.Images)

	var configUser config.IofogUser
	// Check existing controller
	ctrl, err := config.GetController(exe.opt.Namespace, exe.opt.Name)
	if err == nil {
		// Use existing user
		configUser = ctrl.IofogUser
	} else {
		// Generate new user
		password := util.RandomString(10, util.AlphaNum)
		email := util.RandomString(5, util.AlphaLower) + "@domain.com"
		configUser = config.IofogUser{
			Name:     "N" + util.RandomString(10, util.AlphaLower),
			Surname:  "S" + util.RandomString(10, util.AlphaLower),
			Email:    email,
			Password: password,
		}
	}
	// Assign user
	user := iofog.User{
		Name:     configUser.Name,
		Surname:  configUser.Surname,
		Email:    configUser.Email,
		Password: configUser.Password,
	}
	// Create controller on cluster
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
			Password: user.Password,
		},
		Created: util.Now(),
	}
	err = config.UpdateController(exe.opt.Namespace, configEntry)
	if err != nil {
		return
	}

	fmt.Printf("\nController %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)

	return config.Flush()
}
