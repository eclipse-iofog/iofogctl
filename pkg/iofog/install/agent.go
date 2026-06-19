package install

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Agent interface {
	Bootstrap() error
	getProvisionKey(string, IofogUser) (string, string, string, error)
}

// defaultAgent implements commong behavior
type defaultAgent struct {
	name string
	uuid string
}

func (agent *defaultAgent) getProvisionKey(controllerEndpoint string, user IofogUser) (key string, caCert string, err error) {
	// Connect to controller
	baseURL, err := util.GetBaseURL(controllerEndpoint)
	if err != nil {
		return
	}
	// Log in
	util.SpinHandlePrompt()
	ctrl, err := client.SessionLogin(client.Options{BaseURL: baseURL}, user.RefreshToken, user.Email, user.Password)
	if err != nil {
		return
	}
	util.SpinHandlePromptComplete()
	Verbose("Accessing Controller to generate Provisioning Key")
	// loginRequest := client.LoginRequest{
	// 	Email:    user.Email,
	// 	Password: user.Password,
	// }
	// if err = ctrl.Login(loginRequest); err != nil {
	// 	return
	// }

	// System agents have uuid passed through, normal agents dont
	if agent.uuid == "" {
		var agentInfo *client.AgentInfo
		// agentInfo, err = ctrl.GetAgentByName(agent.name, false)
		agentInfo, err = ctrl.GetAgentByName(agent.name)
		if err != nil {
			return
		}
		agent.uuid = agentInfo.UUID
	}

	// Get provisioning key
	provisionResponse, err := ctrl.GetAgentProvisionKey(agent.uuid)
	if err != nil {
		return
	}
	key = provisionResponse.Key
	caCert = provisionResponse.CaCert
	return key, caCert, err
}
