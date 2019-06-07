package get

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
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
	agents, err := config.GetAgents(exe.namespace)
	if err != nil {
		return err
	}

	// Generate table and headers
	table := make([][]string, len(agents)+1)
	headers := []string{"NAME", "STATUS", "AGE"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, agentConfig := range agents {
		// Get age
		age, err := util.Elapsed(agentConfig.Created, util.Now())
		if err != nil {
			return err
		}
		row := []string{
			agentConfig.Name,
			"-",
			age,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	// Print table
	err = print(table)
	return err
}
