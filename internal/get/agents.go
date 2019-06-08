package get

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"time"
)

type agentExecutor struct {
	namespace string
}

func newAgentExecutor(namespace string) *agentExecutor {
	a := &agentExecutor{}
	a.namespace = namespace
	return a
}

func (exe *agentExecutor) Execute() error {
	// Get Config
	agents, err := config.GetAgents(exe.namespace)
	if err != nil {
		return err
	}
	ctrls, err := config.GetControllers(exe.namespace)
	if err != nil {
		return err
	}
	if len(ctrls) != 1 {
		return util.NewInternalError("Expected one controller in namespace " + exe.namespace)
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

	// Generate table and headers
	table := make([][]string, len(agents)+1)
	headers := []string{"AGENT", "STATUS", "AGE", "UPTIME"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, agent := range agents {
		getAgentResponse, err := ctrl.GetAgent(agent.UUID, token)
		if err != nil {
			return err
		}
		age, err := util.Elapsed(util.FromInt(getAgentResponse.Created), util.Now())
		if err != nil {
			return err
		}
		uptime := time.Duration(getAgentResponse.DaemonUptime)
		row := []string{
			agent.Name,
			getAgentResponse.DaemonStatus,
			age,
			util.FormatDuration(uptime),
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	// Print table
	err = print(table)
	return err
}
