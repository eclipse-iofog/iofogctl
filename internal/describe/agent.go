package describe

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type agentExecutor struct {
	namespace string
	name      string
}

func newAgentExecutor(namespace, name string) *agentExecutor {
	a := &agentExecutor{}
	a.namespace = namespace
	a.name = name
	return a
}

func (exe *agentExecutor) Execute() error {
	// Get config
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err != nil {
		return err
	}
	ctrls, err := config.GetControllers(exe.namespace)
	if err != nil {
		return err
	}
	if len(ctrls) != 1 {
		return util.NewInputError("Cannot get Agent data without a Controller in namespace " + exe.namespace)
	}

	// Connect to controller
	ctrl := iofog.NewController(ctrls[0].Endpoint)
	loginRequest := iofog.LoginRequest{
		Email:    ctrls[0].IofogUser.Email,
		Password: ctrls[0].IofogUser.Password,
	}

	// Send requests to controller
	loginResponse, err := ctrl.Login(loginRequest)
	if err != nil {
		return err
	}
	token := loginResponse.AccessToken
	getAgentResponse, err := ctrl.GetAgent(agent.UUID, token)
	if err != nil {
		return err
	}

	// Print result
	if err = print(getAgentResponse); err != nil {
		return err
	}
	return nil
}
