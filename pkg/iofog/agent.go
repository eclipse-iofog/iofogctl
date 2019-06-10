package iofog

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	pb "github.com/schollz/progressbar"
	"os"
)

type Agent struct {
	ssh  *util.SecureShellClient
	name string
}

func NewAgent(user, host string, port int, privKeyFilename, agentName string) *Agent {
	ssh := util.NewSecureShellClient(user, host, privKeyFilename)
	ssh.SetPort(port)
	return &Agent{
		ssh:  ssh,
		name: agentName,
	}
}

func (agent *Agent) Bootstrap() error {
	// Connect to agent over SSH
	err := agent.ssh.Connect()
	if err != nil {
		return err
	}
	defer agent.ssh.Disconnect()

	// Instantiate install arguments
	installURL := "https://raw.githubusercontent.com/eclipse-iofog/platform/feature/dogfood-environment/infrastructure/ansible/scripts/agent.sh"
	installArgs := ""
	pkgCloudToken := os.Getenv("PACKAGE_CLOUD_TOKEN")
	agentVersion := os.Getenv("AGENT_VERSION")
	if pkgCloudToken != "" {
		installArgs += "dev " + pkgCloudToken + " " + agentVersion
	}

	// Execute commands
	cmds := []command{
		{"echo 'APT::Get::AllowUnauthenticated \"true\";' | sudo tee /etc/apt/apt.conf.d/99temp", 1},
		{"sudo apt --assume-yes install apt-transport-https ca-certificates curl software-properties-common jq", 5},
		{"curl " + installURL + " sudo tee /opt/linux.sh " + installArgs, 2},
		{"sudo chmod +x /opt/linux.sh", 1},
		{"sudo /opt/linux.sh", 70},
		{"sudo service iofog-agent start", 3},
		{"sudo iofog-agent config -cf 10 -sf 10", 1},
		{"echo '" + waitForAgentScript + "' | tee ~/wait-for-agent.sh", 1},
		{"sudo chmod +x ~/wait-for-agent.sh", 1},
		{"~/wait-for-agent.sh", 15},
	}

	// Prepare progress bar
	pb := pb.New(100)
	defer pb.Clear()

	// Execute commands
	for _, cmd := range cmds {
		_, err = agent.ssh.Run(cmd.cmd)
		pb.Add(cmd.pbSlice)
		if err != nil {
			return err
		}
	}

	return nil
}

func (agent *Agent) Configure(controllerEndpoint string, user User) (uuid string, err error) {
	pb := pb.New(100)
	defer pb.Clear()

	// Connect to controller
	ctrl := NewController(controllerEndpoint)

	// Log in
	loginRequest := LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	}
	loginResponse, err := ctrl.Login(loginRequest)
	if err != nil {
		return
	}
	token := loginResponse.AccessToken
	pb.Add(20)

	// Create agent
	createRequest := CreateAgentRequest{
		Name:    agent.name,
		FogType: 0,
	}
	createResponse, err := ctrl.CreateAgent(createRequest, token)
	if err != nil {
		return
	}
	uuid = createResponse.UUID
	pb.Add(20)

	// Get provisioning key
	provisionResponse, err := ctrl.GetAgentProvisionKey(uuid, token)
	if err != nil {
		return
	}
	pb.Add(20)
	key := provisionResponse.Key

	// Establish SSH to agent
	err = agent.ssh.Connect()
	if err != nil {
		return
	}

	// Prepare progress bar
	defer agent.ssh.Disconnect()
	pb.Add(20)

	// Instantiate commands
	controllerBaseURL := fmt.Sprintf("http://%s/api/v3", controllerEndpoint)
	cmds := []command{
		{"sudo iofog-agent config -a " + controllerBaseURL, 10},
		{"sudo iofog-agent provision " + key, 10},
	}

	// Execute commands
	for _, cmd := range cmds {
		_, err = agent.ssh.Run(cmd.cmd)
		pb.Add(cmd.pbSlice)
		if err != nil {
			return
		}
	}

	return
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
