package iofog

import (
	"github.com/eclipse-iofog/cli/pkg/util"
)

type Agent struct {
	ssh *util.SecureShellClient
}

func NewAgent(user, host, privKeyFilename string) *Agent {
	return &Agent{
		ssh: util.NewSecureShellClient(user, host, privKeyFilename),
	}
}

func (agent *Agent) Bootstrap() error {
	err := agent.ssh.Connect()
	if err != nil {
		return err
	}
	defer agent.ssh.Disconnect()

	cmds := []string{
		"echo 'APT::Get::AllowUnauthenticated \"true\";' | sudo tee /etc/apt/apt.conf.d/99temp",
		"sudo apt --assume-yes install apt-transport-https ca-certificates curl software-properties-common jq",
		"curl https://raw.githubusercontent.com/eclipse-iofog/iofog.org/saeid/jdk/static/linux.sh | sudo tee /opt/linux.sh",
		"sudo chmod +x /opt/linux.sh",
		"sudo /opt/linux.sh",
		"sudo service iofog-agent start",
		"sudo iofog-agent config -cf 10 -sf 10",
	}
	for _, cmd := range cmds {
		err = agent.ssh.Run(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (agent *Agent) Configure() error {
	err := agent.ssh.Connect()
	if err != nil {
		return err
	}
	defer agent.ssh.Disconnect()

	err = agent.ssh.Run("ls / && echo configure")
	if err != nil {
		return err
	}
	return nil
}
