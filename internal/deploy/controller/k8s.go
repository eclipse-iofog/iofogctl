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
	images := make(map[string]string)
	err = util.UnmarshalYAML(exe.opt.ImagesFile, images)
	if err != nil {
		return err
	}
	k8s.SetImages(images)

	// Generate a user
	password := util.RandomString(10, util.AlphaNum)
	email := util.RandomString(5, util.AlphaLower) + "@domain.com"
	user := iofog.User{
		Name:     "N" + util.RandomString(10, util.AlphaLower),
		Surname:  "S" + util.RandomString(10, util.AlphaLower),
		Email:    email,
		Password: password,
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
			Password: password,
		},
		Created: util.Now(),
	}
	err = config.AddController(exe.opt.Namespace, configEntry)
	if err != nil {
		return
	}

	fmt.Printf("\nController %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)
	return nil
}
