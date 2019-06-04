package iofog

import (
	"fmt"
	"github.com/eclipse-iofog/cli/pkg/util"
	pb "github.com/schollz/progressbar"
)

type Agent struct {
	ssh  *util.SecureShellClient
	name string
}

func NewAgent(user, host, privKeyFilename, agentName string) *Agent {
	return &Agent{
		ssh:  util.NewSecureShellClient(user, host, privKeyFilename),
		name: agentName,
	}
}

func (agent *Agent) Bootstrap() error {
	err := agent.ssh.Connect()
	if err != nil {
		return err
	}
	defer agent.ssh.Disconnect()

	cmds := []command{
		{"echo 'APT::Get::AllowUnauthenticated \"true\";' | sudo tee /etc/apt/apt.conf.d/99temp", 1},
		{"sudo apt --assume-yes install apt-transport-https ca-certificates curl software-properties-common jq", 5},
		{"curl https://raw.githubusercontent.com/eclipse-iofog/iofog.org/saeid/jdk/static/linux.sh | sudo tee /opt/linux.sh", 2},
		{"sudo chmod +x /opt/linux.sh", 1},
		{"sudo /opt/linux.sh", 70},
		{"sudo service iofog-agent start", 3},
		{"sudo iofog-agent config -cf 10 -sf 10", 1},
		{"echo '" + waitForAgentScript + "' | tee ~/wait-for-agent.sh", 1},
		{"chmod +x ~/wait-for-agent.sh", 1},
		{"~/wait-for-agent.sh", 15},
	}

	pb := pb.New(100)
	for _, cmd := range cmds {
		err = agent.ssh.Run(cmd.cmd)
		pb.Add(cmd.pbSlice)
		if err != nil {
			return err
		}
	}

	return nil
}

func (agent *Agent) Configure(controllerEndpoint string, user User) error {
	pb := pb.New(100)

	// Connect to controller
	ctrl := NewController(controllerEndpoint)

	// Log in
	token, err := ctrl.GetAuthToken(user)
	if err != nil {
		return err
	}
	pb.Add(20)

	// Create agent
	uuid, err := ctrl.CreateAgent(token, agent.name)
	if err != nil {
		return err
	}
	pb.Add(20)

	// Get provisioning key
	key, err := ctrl.GetAgentProvisionKey(token, uuid)
	if err != nil {
		return err
	}
	pb.Add(20)

	// Establish SSH to agent
	err = agent.ssh.Connect()
	if err != nil {
		return err
	}
	defer agent.ssh.Disconnect()
	pb.Add(20)

	controllerBaseURL := fmt.Sprintf("http://%s/api/v3", controllerEndpoint)
	cmds := []command{
		{"sudo iofog-agent config -a " + controllerBaseURL, 10},
		{"sudo iofog-agent provision " + key, 10},
	}

	for _, cmd := range cmds {
		err = agent.ssh.Run(cmd.cmd)
		pb.Add(cmd.pbSlice)
		if err != nil {
			return err
		}
	}

	return nil
}

type command struct {
	cmd     string
	pbSlice int
}

var waitForAgentScript = `STATUS=""
ITER=0
while [ "$STATUS" != "RUNNING" ] ; do
    ITER=$((ITER+1))
    if [ "$ITER" -gt 30 ]; then
        echo 'Timed out waiting for Agent to be RUNNING'
        exit 1
    fi
    sleep 1
    STATUS=$(sudo iofog-agent status | cut -f2 -d: | head -n 1 | tr -d '[:space:]')
done
exit 0`
