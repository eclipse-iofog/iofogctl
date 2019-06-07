package logs

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type agentExecutor struct {
	namespace string
	name      string
}

func newAgentExecutor(namespace, name string) *agentExecutor {
	exe := &agentExecutor{}
	exe.namespace = namespace
	exe.name = name
	return exe
}

func (exe *agentExecutor) Execute() error {
	// Get agent config
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	// Establish SSH connection
	ssh := util.NewSecureShellClient(agent.User, agent.Host, agent.KeyFile)
	err = ssh.Connect()
	if err != nil {
		return err
	}

	// Get logs
	out, err := ssh.Run("sudo cat /var/log/iofog-agent/*")
	if err != nil {
		return err
	}
	println(out.String())

	return nil
}
