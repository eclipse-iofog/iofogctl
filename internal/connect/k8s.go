package connect

import (
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
	// Instantiate Kubernetes cluster object
	k8s, err := iofog.NewKubernetes(exe.opt.KubeFile)
	if err != nil {
		return err
	}

	// Get Controller endpoint
	endpoint, err := k8s.GetControllerEndpoint()
	if err != nil {
		return err
	}

	// Connect to Controller
	ctrl := iofog.NewController(endpoint)

	// Generate a user
	password := util.RandomString(10, util.AlphaNum)
	email := util.RandomString(5, util.AlphaLower) + "@domain.com"
	user := iofog.User{
		Name:     "N" + util.RandomString(10, util.AlphaLower),
		Surname:  "S" + util.RandomString(10, util.AlphaLower),
		Email:    email,
		Password: password,
	}

	// Sign user up
	err = ctrl.CreateUser(user)
	if err != nil {
		return err
	}
	// Login user
	loginRequest := iofog.LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	}
	loginResponse, err := ctrl.Login(loginRequest)
	if err != nil {
		return err
	}
	token := loginResponse.AccessToken

	// Get Agents
	listAgentsResponse, err := ctrl.ListAgents(token)
	if err != nil {
		return err
	}

	// Update Agents config
	for _, agent := range listAgentsResponse.Agents {
		agentConfig := config.Agent{
			Name: agent.Name,
			UUID: agent.UUID,
		}
		err = config.AddAgent(exe.opt.Namespace, agentConfig)
		if err != nil {
			return err
		}
	}

	// Update Controller config
	ctrlConfig := config.Controller{
		Name:     exe.opt.Name,
		Endpoint: endpoint,
		IofogUser: config.IofogUser{
			Name:     user.Name,
			Surname:  user.Surname,
			Email:    user.Email,
			Password: user.Password,
		},
		KubeConfig: exe.opt.KubeFile,
	}
	err = config.AddController(exe.opt.Namespace, ctrlConfig)
	if err != nil {
		return err
	}

	return config.Flush()
}
