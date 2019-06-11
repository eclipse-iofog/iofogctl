package connect

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
)

func connect(opt *Options, endpoint string) error {
	// Connect to Controller
	ctrl := iofog.NewController(endpoint)

	// Login user
	loginRequest := iofog.LoginRequest{
		Email:    opt.Email,
		Password: opt.Password,
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
		err = config.AddAgent(opt.Namespace, agentConfig)
		if err != nil {
			return err
		}
	}

	// Update Controller config
	ctrlConfig := config.Controller{
		Name:     opt.Name,
		Endpoint: endpoint,
		IofogUser: config.IofogUser{
			Email:    opt.Email,
			Password: opt.Password,
		},
		KubeConfig: opt.KubeFile,
	}
	err = config.AddController(opt.Namespace, ctrlConfig)
	if err != nil {
		return err
	}

	return nil
}
