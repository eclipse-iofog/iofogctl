package connect

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
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

	// Login user
	loginRequest := iofog.LoginRequest{
		Email:    exe.opt.Email,
		Password: exe.opt.Password,
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
			Email:    exe.opt.Email,
			Password: exe.opt.Password,
		},
		KubeConfig: exe.opt.KubeFile,
	}
	err = config.AddController(exe.opt.Namespace, ctrlConfig)
	if err != nil {
		return err
	}

	return config.Flush()
}
