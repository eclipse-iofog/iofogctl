package logs

import (
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteMicroserviceExecutor struct {
	namespace string
	name      string
	logConfig *LogTailConfig
}

func newRemoteMicroserviceExecutor(namespace, name string, logConfig *LogTailConfig) *remoteMicroserviceExecutor {
	m := &remoteMicroserviceExecutor{}
	m.namespace = namespace
	m.name = name
	m.logConfig = logConfig
	return m
}

func (ms *remoteMicroserviceExecutor) GetName() string {
	return ms.name
}

func (ms *remoteMicroserviceExecutor) Execute() error {
	// Get image name of the microservice and details of the Agent its deployed on
	baseAgent, msvc, isSystem, err := getAgentAndMicroservice(ms.namespace, ms.name)
	if err != nil {
		return err
	}

	if msvc.Status.Status != "RUNNING" {
		return util.NewError("The microservice is not currently running")
	}

	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		sdkCfg := localAgentSDKConfig(agent)
		lc, err := install.NewLocalContainerClient(install.LocalContainerEngineForHostOps(sdkCfg), sdkCfg)
		if err != nil {
			return err
		}
		containerName := "iofog_" + msvc.UUID
		stdout, stderr, err := lc.GetLogsByName(containerName)
		if err != nil {
			return err
		}

		printContainerLogs(stdout, stderr)

		return nil
	case *rsc.RemoteAgent:
		util.SpinStart("Connecting to Microservice Logs")

		clt, err := clientutil.NewControllerClient(ms.namespace)
		if err != nil {
			util.SpinHandlePromptComplete()
			return err
		}

		opts := ms.logConfig.ToSDKOptions()
		return runRemoteLogStream(clt, func(clt *client.Client) (*client.LogSession, error) {
			if isSystem {
				return clt.DialSystemMicroserviceLogs(msvc.UUID, opts)
			}
			return clt.DialMicroserviceLogs(msvc.UUID, opts)
		})
	}

	return nil
}

// func (ms *remoteMicroserviceExecutor) runDockerCommand(cmd string, ssh *util.SecureShellClient) (stdout bytes.Buffer, err error) {
// 	stdout, err = ssh.Run(cmd)
// 	if err != nil {
// 		if !strings.Contains(strings.ToLower(err.Error()), "permission denied") {
// 			return
// 		}
// 		// Retry with sudo
// 		cmd = strings.Replace(cmd, "docker", "sudo docker", -1)

// 		stdout, err = ssh.Run(cmd)
// 		if err != nil {
// 			return
// 		}
// 	}
// 	return
// }

func getAgentAndMicroservice(namespace, msvcFQName string) (agent rsc.Agent, msvc client.MicroserviceInfo, isSystem bool, err error) {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return
	}

	ctrlClient, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return
	}

	appName, msvcName, err := clientutil.ParseFQName(msvcFQName, "Microservice")
	if err != nil {
		return agent, msvc, false, err
	}

	// Get microservice details from Controller
	msvcPtr, err := ctrlClient.GetMicroserviceByName(appName, msvcName)
	isSystem = false
	if err != nil {
		// Check if error indicates application not found
		if strings.Contains(err.Error(), "Invalid application id") {
			// Try system application
			msvcPtr, err = ctrlClient.GetSystemMicroserviceByName(appName, msvcName)
			if err != nil {
				return
			}
			isSystem = true
		} else {
			// Return other types of errors
			return
		}
	}

	msvc = *msvcPtr

	// Get Agent running the microservice
	agentResponse, err := ctrlClient.GetAgentByID(msvc.AgentUUID)
	if err != nil {
		return
	}
	agent, err = ns.GetAgent(agentResponse.Name)
	if err != nil {
		return
	}
	return agent, msvc, isSystem, nil
}
