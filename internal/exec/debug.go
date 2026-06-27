package exec

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const (
	debugPollInterval    = 2 * time.Second
	debugPollMaxAttempts = 60
)

func isDebugMicroserviceName(name, agentName, agentUUID string) bool {
	switch name {
	case "debug", "debug-" + agentName, "debug-" + agentUUID:
		return true
	default:
		return false
	}
}

func isAgentRunning(agent *client.AgentInfo) bool {
	return strings.EqualFold(agent.DaemonStatus, "RUNNING")
}

func isMicroserviceRunning(msvc *client.MicroserviceInfo) bool {
	return strings.EqualFold(msvc.Status.Status, "RUNNING")
}

func findDebugMicroservice(clt *client.Client, agent *client.AgentInfo) (*client.MicroserviceInfo, error) {
	appName := "system-" + agent.Name
	list, err := clt.GetSystemMicroservicesByApplication(appName)
	if err != nil {
		if isMissingApplicationError(err) {
			return nil, nil
		}
		return nil, err
	}

	for i := range list.Microservices {
		msvc := &list.Microservices[i]
		if msvc.AgentUUID != agent.UUID {
			continue
		}
		if isDebugMicroserviceName(msvc.Name, agent.Name, agent.UUID) {
			return msvc, nil
		}
	}

	return nil, nil
}

func isMissingApplicationError(err error) bool {
	var notFound *client.NotFoundError
	if errors.As(err, &notFound) {
		return true
	}
	return strings.Contains(err.Error(), "Invalid application id")
}

func attachDebugExec(clt *client.Client, agent *client.AgentInfo, image *string) error {
	err := clt.AttachExecToAgent(&client.AttachExecToAgentRequest{
		UUID:  agent.UUID,
		Image: image,
	})
	if err == nil {
		return nil
	}
	if isDebugExecAlreadyProvisioned(err) {
		return nil
	}
	return fmt.Errorf("failed to provision fog debug exec: %w", err)
}

func isDebugExecAlreadyProvisioned(err error) bool {
	var conflict *client.ConflictError
	if errors.As(err, &conflict) {
		return true
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "already") ||
		strings.Contains(errStr, "conflict") ||
		strings.Contains(errStr, "exists")
}

func ensureDebugExecReady(clt *client.Client, agentName string, image *string) (*client.MicroserviceInfo, error) {
	agent, err := clt.GetAgentByName(agentName)
	if err != nil {
		return nil, err
	}
	if !isAgentRunning(agent) {
		return nil, util.NewError(ErrMsgAgentNotRunning)
	}

	msvc, err := findDebugMicroservice(clt, agent)
	if err != nil {
		return nil, err
	}

	switch {
	case msvc == nil:
		util.PrintNotify(fmt.Sprintf(
			"Debug microservice not found. Provisioning fog debug exec for Agent %s...",
			agentName,
		))
		if err := attachDebugExec(clt, agent, image); err != nil {
			return nil, err
		}
		util.PrintNotify("Waiting for debug container to start...")
	case !isMicroserviceRunning(msvc):
		util.PrintNotify("Waiting for debug container to start...")
	default:
		return msvc, nil
	}

	return waitForDebugMicroserviceRunning(clt, agent)
}

func waitForDebugMicroserviceRunning(clt *client.Client, agent *client.AgentInfo) (*client.MicroserviceInfo, error) {
	startingNotified := false

	for attempt := 0; attempt < debugPollMaxAttempts; attempt++ {
		msvc, err := findDebugMicroservice(clt, agent)
		if err != nil {
			return nil, err
		}
		if msvc != nil {
			if isMicroserviceRunning(msvc) {
				util.PrintNotify("Debug container is running. Connecting to terminal...")
				return msvc, nil
			}
			if !startingNotified {
				util.PrintNotify("Debug container is starting...")
				startingNotified = true
			}
		}

		time.Sleep(debugPollInterval)
	}

	return nil, util.NewError(ErrMsgDebugContainerTimeout)
}
