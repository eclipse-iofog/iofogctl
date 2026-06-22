package deployagentconfig

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type RouterMode string

type NatsMode string

const (
	EdgeRouter     RouterMode = "edge"
	InteriorRouter RouterMode = "interior"
	NoneRouter     RouterMode = "none"
	NatsNone       NatsMode   = "none"
	NatsLeaf       NatsMode   = "leaf"
	NatsServer     NatsMode   = "server"
)

func getRouterMode(config *rsc.AgentConfiguration) RouterMode {
	if config.RouterMode != nil {
		return RouterMode(*config.RouterMode)
	}
	return EdgeRouter
}

func getNatsMode(config *rsc.AgentConfiguration) NatsMode {
	if config.NatsMode != nil {
		return NatsMode(*config.NatsMode)
	}
	return NatsLeaf
}

func Validate(config *rsc.AgentConfiguration) error {
	routerMode := getRouterMode(config)
	natsMode := getNatsMode(config)

	if routerMode != EdgeRouter && routerMode != InteriorRouter && routerMode != NoneRouter {
		msg := "agent config %s validation failed. RouterMode has to be one of edge, interior, none. Default is: edge"
		return util.NewInputError(fmt.Sprintf(msg, config.Name))
	}
	if routerMode != NoneRouter && config.NetworkRouter != nil {
		msg := "agent config %s validation failed. Cannot have a network if routerMode is different from none. Current router mode is: %s"
		return util.NewInputError(fmt.Sprintf(msg, config.Name, routerMode))
	}
	if routerMode == NoneRouter && config.UpstreamRouters != nil && len(*config.UpstreamRouters) > 0 {
		msg := "agent config %s validation failed. Cannot have a upstreamRouters if routerMode is none"
		return util.NewInputError(fmt.Sprintf(msg, config.Name))
	}
	if routerMode != InteriorRouter && (config.EdgeRouterPort != nil || config.InterRouterPort != nil) {
		msg := "agent config %s validation failed. Cannot have an edgeRouterPort or interRouterPort if routerMode is different from interior. Current router mode is: %s"
		return util.NewInputError(fmt.Sprintf(msg, config.Name, routerMode))
	}
	if natsMode != NatsServer && (config.NatsClusterPort != nil) {
		msg := "agent config %s validation failed. Cannot have a natsClusterPort if natsMode is different from server"
		return util.NewInputError(fmt.Sprintf(msg, config.Name))
	}

	return nil
}

func findAgentUUIDInList(list []client.AgentInfo, name string) (uuid string, err error) {
	// if name == iofog.VanillaRemoteAgentName {
	// 	return name, nil
	// }
	if name == iofog.VanillaRouterAgentName {
		return name, nil
	}
	for idx := range list {
		agent := &list[idx]
		if agent.Name == name {
			return agent.UUID, nil
		}
	}
	return "", util.NewNotFoundError(fmt.Sprintf("Could not find router: %s\n", name))
}

// Process update the config to translate agent names into uuids, and sets the host value if needed
func Process(agentConfig *rsc.AgentConfiguration, name, agentIP string, otherAgents []client.AgentInfo) error {
	routerMode := getRouterMode(agentConfig)

	if agentConfig.UpstreamRouters != nil {
		upstreamRoutersUUID := []string{}
		for _, agentName := range *agentConfig.UpstreamRouters {
			uuid, err := findAgentUUIDInList(otherAgents, agentName)
			if err != nil {
				return err
			}
			upstreamRoutersUUID = append(upstreamRoutersUUID, uuid)
		}
		agentConfig.UpstreamRouters = &upstreamRoutersUUID
	}

	if agentConfig.NetworkRouter != nil {
		uuid, err := findAgentUUIDInList(otherAgents, *agentConfig.NetworkRouter)
		if err != nil {
			return err
		}
		agentConfig.NetworkRouter = &uuid
	}
	if agentConfig.UpstreamNatsServers != nil {
		upstreamNatsServersUUID := []string{}
		for _, agentName := range *agentConfig.UpstreamNatsServers {
			uuid, err := findAgentUUIDInList(otherAgents, agentName)
			if err != nil {
				// Keep raw value for controller-reserved NATS hub aliases.
				upstreamNatsServersUUID = append(upstreamNatsServersUUID, agentName)
				continue
			}
			upstreamNatsServersUUID = append(upstreamNatsServersUUID, uuid)
		}
		agentConfig.UpstreamNatsServers = &upstreamNatsServersUUID
	}

	if routerMode != NoneRouter && agentConfig.Host == nil {
		agentConfig.Host = &agentIP
	}

	return nil
}

func getAgentUpdateRequestFromAgentConfig(agentConfig *rsc.AgentConfiguration, tags *[]string) (request client.AgentUpdateRequest) {
	var archPtr *int64
	if agentConfig.Arch != nil {
		arch, found := rsc.ArchStringToID(*agentConfig.Arch)
		if !found {
			arch = 0
		}
		archPtr = &arch
	}
	request.Location = agentConfig.Location
	request.Latitude = agentConfig.Latitude
	request.Longitude = agentConfig.Longitude
	request.Description = agentConfig.Description
	request.Name = agentConfig.Name
	request.ArchID = archPtr
	request.AgentConfiguration = agentConfig.AgentConfiguration
	request.Tags = tags
	return
}

func createAgentFromConfiguration(agentConfig *rsc.AgentConfiguration, tags *[]string, name string, clt *client.Client) (uuid string, err error) {
	updateAgentConfigRequest := getAgentUpdateRequestFromAgentConfig(agentConfig, tags)
	createAgentRequest := &client.CreateAgentRequest{
		AgentUpdateRequest: updateAgentConfigRequest,
	}
	if createAgentRequest.Name == "" {
		createAgentRequest.Name = name
	}
	if createAgentRequest.ArchID == nil {
		arch := int64(0)
		createAgentRequest.ArchID = &arch
	}
	agent, err := clt.CreateAgent(createAgentRequest)
	if err != nil {
		return "", err
	}
	return agent.UUID, nil
}

func updateAgentConfiguration(agentConfig *rsc.AgentConfiguration, tags *[]string, uuid string, clt *client.Client) (err error) {
	if agentConfig != nil {
		updateAgentConfigRequest := getAgentUpdateRequestFromAgentConfig(agentConfig, tags)
		updateAgentConfigRequest.UUID = uuid

		// Get current agent info to preserve host value only if not explicitly set in config
		agentInfo, getErr := clt.GetAgentByID(uuid)
		if getErr != nil {
			return getErr
		}

		// Only preserve the original host value if the config doesn't explicitly set a host
		if agentConfig.Host == nil {
			host := agentInfo.Host
			updateAgentConfigRequest.Host = &host
		}

		if _, err = clt.UpdateAgent(&updateAgentConfigRequest); err != nil {
			return
		}
	}
	return nil
}
