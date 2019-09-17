package install

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type ConnectorOptions struct {
	Name               string
	User               string
	Host               string
	Port               int
	PrivKeyFilename    string
	Version            string
	PackageCloudToken  string
	IofogUser          IofogUser
	ControllerEndpoint string
}

type Connector struct {
	*ConnectorOptions
	ssh *util.SecureShellClient
}

func NewConnector(options *ConnectorOptions) *Connector {
	ssh := util.NewSecureShellClient(options.User, options.Host, options.PrivKeyFilename)
	ssh.SetPort(options.Port)
	return &Connector{
		ConnectorOptions: options,
		ssh:              ssh,
	}
}

func (cnct *Connector) Install() (err error) {
	// Connect to server
	verbose("Connecting to server")
	if err = cnct.ssh.Connect(); err != nil {
		return
	}
	defer cnct.ssh.Disconnect()

	// Copy installation scripts to remote host
	verbose("Copying install files to server")
	installConnectorScript := util.GetStaticFile("install_connector.sh")
	reader := strings.NewReader(installConnectorScript)
	if err := cnct.ssh.CopyTo(reader, "/tmp/", "install_connector.sh", "0775", len(installConnectorScript)); err != nil {
		return err
	}

	// Define commands
	cmds := []string{
		fmt.Sprintf("/tmp/install_connector.sh %s %s", cnct.Version, cnct.PackageCloudToken),
	}

	// Execute commands
	for _, cmd := range cmds {
		verbose("Running command: " + cmd)
		_, err = cnct.ssh.Run(cmd)
		if err != nil {
			return
		}
	}

	// Specify errors to ignore while waiting
	ignoredErrors := []string{
		"Process exited with status 7", // curl: (7) Failed to connect to localhost port 8080: Connection refused
	}
	// Wait for Connector
	verbose("Waiting for Connector")
	if err = cnct.ssh.RunUntil(
		regexp.MustCompile("\"status\":\"running\""),
		fmt.Sprintf("curl --request POST --url http://localhost:%s/api/v2/status --header 'Content-Type: application/x-www-form-urlencoded' --data mappingid=all", iofog.ConnectorPortString),
		ignoredErrors,
	); err != nil {
		return
	}

	// Provision the Connector with Controller
	verbose("Provisioning Connector")
	ctrlClient := client.New(cnct.ControllerEndpoint)
	loginRequest := client.LoginRequest{
		Email:    cnct.IofogUser.Email,
		Password: cnct.IofogUser.Password,
	}
	if err = ctrlClient.Login(loginRequest); err != nil {
		return
	}
	if err = ctrlClient.AddConnector(client.ConnectorInfo{
		IP:      cnct.Host,
		Domain:  cnct.Host,
		Name:    cnct.Name,
		DevMode: true,
	}); err != nil {
		return
	}

	return
}

func (cnct *Connector) Stop() (err error) {
	// Connect to server
	if err = cnct.ssh.Connect(); err != nil {
		return
	}
	defer cnct.ssh.Disconnect()

	// TODO: Clear the database
	// Define commands
	cmds := []string{
		"sudo systemctl stop iofog-connector",
	}

	// Execute commands
	for _, cmd := range cmds {
		_, err = cnct.ssh.Run(cmd)
		if err != nil {
			return
		}
	}

	return
}
