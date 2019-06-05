package deleteagent

import (
	"fmt"
	"github.com/eclipse-iofog/cli/internal/config"
	"github.com/eclipse-iofog/cli/pkg/iofog"
	"github.com/eclipse-iofog/cli/pkg/util"
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
		return util.NewInternalError("Expected one Controller in namespace " + exe.namespace)
	}

	// Get Controller endpoint and connect to Controller
	endpoint := ctrlConfigs[0].Endpoint
	ctrl := iofog.NewController(endpoint)

	// Log into Controller
	userConfig := ctrlConfigs[0].IofogUser
	user := iofog.User{
		Name:     userConfig.Name,
		Surname:  userConfig.Surname,
		Email:    userConfig.Email,
		Password: userConfig.Password,
	}
	token, err := ctrl.GetAuthToken(user)
	if err != nil {
		return err
	}

	// Perform deletion of Agent through Controller
	err = ctrl.DeleteAgent(token, agent.UUID)
	if err != nil {
		return err
	}

	// Update configuration
	err = config.DeleteAgent(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	// TODO (Serge) Handle config file error, retry..?

	fmt.Printf("\nAgent %s/%s successfully deleted.\n", exe.namespace, exe.name)

	return nil
}
