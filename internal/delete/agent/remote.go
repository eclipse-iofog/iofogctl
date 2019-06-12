package deleteagent

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"strings"
)

type remoteExecutor struct {
	namespace string
	name      string
}

func newRemoteExecutor(namespace, name string) *remoteExecutor {
	exe := &remoteExecutor{}
	exe.namespace = namespace
	exe.name = name
	return exe
}

func (exe *remoteExecutor) Execute() error {
	// Check the agent exists
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err != nil {
		return err
	}
	// Get Controller for the namespace
	ctrlConfigs, err := config.GetControllers(exe.namespace)
	if err != nil {
		return err
	}
	if len(ctrlConfigs) != 1 {
		return util.NewInputError("Cannot delete an Agent without a Controller in namespace " + exe.namespace)
	}

	// Get Controller endpoint and connect to Controller
	endpoint := ctrlConfigs[0].Endpoint
	ctrl := iofog.NewController(endpoint)

	// Log into Controller
	userConfig := ctrlConfigs[0].IofogUser
	user := iofog.LoginRequest{
		Email:    userConfig.Email,
		Password: userConfig.Password,
	}
	loginResponse, err := ctrl.Login(user)
	if err != nil {
		return err
	}
	token := loginResponse.AccessToken

	// Perform deletion of Agent through Controller
	if err = ctrl.DeleteAgent(agent.UUID, token); err != nil {
		if !strings.Contains(err.Error(), "NotFoundError") {
			return err
		}
	}

	// Update configuration
	if err = config.DeleteAgent(exe.namespace, exe.name); err != nil {
		return err
	}

	fmt.Printf("\nAgent %s/%s successfully deleted.\n", exe.namespace, exe.name)

	return config.Flush()
}
