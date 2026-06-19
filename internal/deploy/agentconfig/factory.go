package deployagentconfig

import (
	// "errors"
	"fmt"
	"net/url"
	"strings"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/internal/util"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"gopkg.in/yaml.v2"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
	Tags      *[]string
}

type AgentConfigExecutor interface {
	execute.Executor
	GetAgentUUID() string
	SetHost(string)
	SetTags(*[]string)
	GetConfiguration() *rsc.AgentConfiguration
	GetNamespace() string
}

type RemoteExecutor struct {
	name        string
	uuid        string
	agentConfig *rsc.AgentConfiguration
	namespace   string
	tags        *[]string
}

func NewRemoteExecutor(name string, conf *rsc.AgentConfiguration, namespace string, tags *[]string) *RemoteExecutor {
	return &RemoteExecutor{
		name:        name,
		agentConfig: conf,
		namespace:   namespace,
		tags:        tags,
	}
}

func (exe *RemoteExecutor) GetNamespace() string {
	return exe.namespace
}

func (exe *RemoteExecutor) GetConfiguration() *rsc.AgentConfiguration {
	return exe.agentConfig
}

func (exe *RemoteExecutor) SetHost(host string) {
	exe.agentConfig.Host = &host
}

func (exe *RemoteExecutor) SetTags(tags *[]string) {
	// Merge tags
	if tags != nil {
		if exe.tags == nil {
			exe.tags = tags
		} else {
			newTagsSlice := append(*exe.tags, *tags...)
			exe.tags = &newTagsSlice
		}
	}
}

func (exe *RemoteExecutor) GetAgentUUID() string {
	return exe.uuid
}

func (exe *RemoteExecutor) GetName() string {
	return exe.name
}

func isOverridingSystemAgent(controllerHost, agentHost, agentName string, isSystem bool) (err error) {
	// Generate controller endpoint
	controllerURL, err := url.Parse(controllerHost)
	if err != nil || controllerURL.Host == "" {
		controllerURL, err = url.Parse("//" + controllerHost) // Try to see if controllerEndpoint is an IP, in which case it needs to be pefixed by //
		if err != nil {
			return err
		}
	}
	agentURL, err := url.Parse(agentHost)
	if err != nil || agentURL.Host == "" {
		agentURL, err = url.Parse("//" + agentHost) // Try to see if controllerEndpoint is an IP, in which case it needs to be pefixed by //
		if err != nil {
			return err
		}
	}
	if agentURL.Hostname() == controllerURL.Hostname() && isSystem == false && agentName != iofog.VanillaLocalAgentName {
		return util.NewConflictError("Cannot deploy an agent on the same host than the Controller\n")
	}
	return nil
}

func (exe *RemoteExecutor) Execute() error {
	isSystem := iutil.IsSystemAgent(exe.agentConfig)

	// Check we are not about to override Vanilla system agent
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil || len(controlPlane.GetControllers()) == 0 {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return err
	}

	endpoint, err := controlPlane.GetEndpoint()
	if err != nil {
		fmt.Println("Error occurred while fetching endpoint from controlplane", err)
		return err
	}

	if !isSystem || install.IsVerbose() {
		util.SpinStart(fmt.Sprintf("Deploying agent %s configuration", exe.GetName()))
	}

	host := ""
	if exe.agentConfig.Host != nil {
		host = *exe.agentConfig.Host
	}

	if err := isOverridingSystemAgent(endpoint, host, exe.name, isSystem); err != nil {
		return err
	}

	// Get the Agent in question with auth retry
	var agent *client.AgentInfo
	err = clientutil.ExecuteWithAuthRetry(exe.namespace, func(ctrlClient *client.Client) error {
		var err error
		agent, err = ctrlClient.GetAgentByName(exe.name)
		// TODO: replace this check with built-in IsNewNotFound() func from go-sdk
		if err != nil && !strings.Contains(err.Error(), "not find agent") {
			return err
		}
		return nil // Return nil for "not found" errors as they are expected
	})
	if err != nil {
		return err
	}
	ip := ""
	if agent != nil {
		ip = agent.IPAddressExternal
	}
	// Get all other non-system Agents with auth retry
	var agentList client.ListAgentsResponse
	err = clientutil.ExecuteWithAuthRetry(exe.namespace, func(ctrlClient *client.Client) error {
		var err error
		agentList, err = ctrlClient.ListAgents(client.ListAgentsRequest{})
		return err
	})
	if err != nil {
		return err
	}
	// Process needs to be done at execute time because agent might have been created during deploy
	if err := Process(exe.agentConfig, exe.name, ip, agentList.Agents); err != nil {
		return err
	}

	// Create if Agent does not exist
	if agent == nil {
		var uuid string
		err = clientutil.ExecuteWithAuthRetry(exe.namespace, func(ctrlClient *client.Client) error {
			var err error
			uuid, err = createAgentFromConfiguration(exe.agentConfig, exe.tags, exe.name, ctrlClient)
			return err
		})
		if err != nil {
			return err
		}
		exe.uuid = uuid
		return nil
	}
	// Update existing Agent
	exe.uuid = agent.UUID
	return clientutil.ExecuteWithAuthRetry(exe.namespace, func(ctrlClient *client.Client) error {
		return updateAgentConfiguration(exe.agentConfig, exe.tags, agent.UUID, ctrlClient)
	})
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Unmarshal file
	agentConfig := rsc.AgentConfiguration{}
	if err = yaml.UnmarshalStrict(opt.Yaml, &agentConfig); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	if agentConfig.Name == "" {
		agentConfig.Name = opt.Name
	}

	if err = Validate(&agentConfig); err != nil {
		return
	}

	return &RemoteExecutor{
		name:        opt.Name,
		agentConfig: &agentConfig,
		namespace:   opt.Namespace,
		tags:        opt.Tags,
	}, nil
}
