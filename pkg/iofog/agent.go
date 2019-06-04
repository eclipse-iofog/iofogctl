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

func (agent *Agent) Configure(controllerBaseURL string) error {
	err := agent.ssh.Connect()
	if err != nil {
		return err
	}
	defer agent.ssh.Disconnect()

	cmds := []string{
		"echo '" + waitForAgentScript + "' | tee ~/wait-for-agent.sh",
		"chmod +x ~/wait-for-agent.sh",
		"~/wait-for-agent.sh",
		"sudo iofog-agent config -a " + controllerBaseURL,
		"echo '" + initScript + "' | tee ~/init.sh",
		"chmod +x ~/init.sh",
		"~/init.sh " + controllerBaseURL,
	}
	for _, cmd := range cmds {
		err = agent.ssh.Run(cmd)
		if err != nil {
			return err
		}
	}

	return nil
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

var initScript = `#!/usr/bin/env bash
set -e
set -x

CONTROLLER_HOST=$1

token=""
uuid=""

function login() {
    echo "Logging in"
    login=$(curl --request POST \
        --url $CONTROLLER_HOST/user/login \
        --header "Content-Type: application/json" \
        --data "{\"email\":\"user@domain.com\",\"password\":\"#Bugs4Fun\"}")
    echo "$login"
    token=$(echo $login | jq -r .accessToken)
}


function create-node() {
    echo "Creating node"
    node=$(curl --request POST \
        --url $CONTROLLER_HOST/iofog \
        --header "Authorization: $token" \
        --header "Content-Type: application/json" \
        --data "{\"name\":\"agent-smith\",\"fogType\":0}")
    echo "$node"
    uuid=$(echo $node | jq -r .uuid)
}

function provision() {
    echo "Provisioning key"
    provisioning=$(curl --request GET \
        --url $CONTROLLER_HOST/iofog/$uuid/provisioning-key \
        --header "Authorization: $token" \
        --header "Content-Type: application/json")
    echo "$provisioning"
    key=$(echo $provisioning | jq -r .key)

    sudo iofog-agent provision $key
}

login
create-node
provision`
