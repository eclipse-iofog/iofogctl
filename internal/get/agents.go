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
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	if len(ns.Controllers) > 1 {
		return util.NewInternalError("Expected 0 or 1 controller in namespace " + exe.namespace)
	}

	// Generate table and headers
	table := make([][]string, len(ns.Agents)+1)
	headers := []string{"AGENT", "STATUS", "AGE", "UPTIME"}
	table[0] = append(table[0], headers...)

	// Print empty table if no controller
	if len(ns.Controllers) == 0 {
		if len(ns.Agents) != 0 {
			return util.NewInternalError("Found Agents without a Controller in Namespace " + ns.Name)
		}
		print(table)
		return nil
	}

	// Connect to controller
	ctrl := iofog.NewController(ns.Controllers[0].Endpoint)
	loginRequest := iofog.LoginRequest{
		Email:    ns.Controllers[0].IofogUser.Email,
		Password: ns.Controllers[0].IofogUser.Password,
	}

	// Send requests to controller
	loginResponse, err := ctrl.Login(loginRequest)
	if err != nil {
		return err
	}
	token := loginResponse.AccessToken

	// Populate rows
	for idx, agent := range ns.Agents {
		getAgentResponse, err := ctrl.GetAgent(agent.UUID, token)
		if err != nil {
			return err
		}
		age, err := util.Elapsed(util.FromInt(getAgentResponse.CreatedTimeMsUTC), util.Now())
		if err != nil {
			return err
		}
		uptime := time.Duration(getAgentResponse.DaemonUptimeDurationMsUTC)
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
