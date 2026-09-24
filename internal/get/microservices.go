package get

import (
	"fmt"
	"math"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type microserviceExecutor struct {
	namespace  string
	client     *client.Client
	msvcPerID  map[string]*client.MicroserviceInfo
	agentPerID map[string]*client.AgentInfo
}

func newMicroserviceExecutor(namespace string) *microserviceExecutor {
	a := &microserviceExecutor{}
	a.namespace = namespace
	a.msvcPerID = make(map[string]*client.MicroserviceInfo)
	a.agentPerID = make(map[string]*client.AgentInfo)
	return a
}

func (exe *microserviceExecutor) init() (err error) {
	exe.client, err = clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		if rsc.IsNoControlPlaneError(err) {
			return nil
		}
		return
	}
	listMsvcs, err := exe.client.GetAllMicroservices()
	if err != nil {
		return err
	}
	for i := 0; i < len(listMsvcs.Microservices); i++ {
		exe.msvcPerID[listMsvcs.Microservices[i].UUID] = &listMsvcs.Microservices[i]
	}

	listAgents, err := exe.client.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return err
	}
	for i := 0; i < len(listAgents.Agents); i++ {
		exe.agentPerID[listAgents.Agents[i].UUID] = &listAgents.Agents[i]
	}
	return
}

func (exe *microserviceExecutor) GetName() string {
	return ""
}

func (exe *microserviceExecutor) Execute() error {
	// Fetch data
	if err := exe.init(); err != nil {
		return err
	}
	printNamespace(exe.namespace)
	table := exe.generateMicroserviceOutput()
	return print(table)
}

func (exe *microserviceExecutor) generateMicroserviceOutput() (table [][]string) {
	// Generate table and headers
	table = make([][]string, len(exe.msvcPerID)+1)
	headers := []string{"MICROSERVICE", "STATUS", "AGENT", "NATS ACCESS", "VOLUMES", "PORTS"}
	table[0] = append(table[0], headers...)

	// Populate rows
	count := 0
	for _, ms := range exe.msvcPerID {
		if util.IsSystemMsvc(ms) {
			continue
		}

		volumes := ""
		for idx, volume := range ms.Volumes {
			if idx == 0 {
				volumes += fmt.Sprintf("%s:%s", volume.HostDestination, volume.ContainerDestination)
			} else {
				volumes += fmt.Sprintf(", %s:%s", volume.HostDestination, volume.ContainerDestination)
			}
		}
		ports := ""
		for idx, port := range ms.Ports {
			if idx == 0 {
				ports += fmt.Sprintf("%v:%v", port.External, port.Internal)
			} else {
				ports += fmt.Sprintf(", %v:%v", port.External, port.Internal)
			}
		}
		agent, ok := exe.agentPerID[ms.AgentUUID]
		var agentName string
		if !ok {
			agentName = "-"
		} else {
			agentName = agent.Name
		}
		status := formatMicroserviceGetStatus(ms.Status)
		natsAccess := "false"
		if ms.NatsConfig != nil && ms.NatsConfig.NatsAccess {
			natsAccess = "true"
		}

		row := []string{
			ms.Name,
			status,
			agentName,
			natsAccess,
			volumes,
			ports,
		}
		table[count+1] = append(table[count+1], row...)
		count++
	}

	return table
}

// formatMicroserviceGetStatus builds the STATUS cell. lastErrorAt is omitted (describe only).
// Empty lastError + restartCount 0 (healthy / older Controller) stays quiet.
func formatMicroserviceGetStatus(st client.MicroserviceStatusInfo) string {
	status := st.Status
	switch status {
	case "":
		status = "-"
	case "PULLING":
		if st.Percentage > 0 {
			status = fmt.Sprintf("%s (%d%s)", st.Status, int(math.Round(st.Percentage)), "%")
		}
	}
	if st.ErrorMessage != "" {
		msg := st.ErrorMessage
		if strings.Contains(msg, "invalid mount config for type \"bind\"") {
			msg = "Volume missing"
		} else if strings.Contains(msg, "runtime create failed") {
			msg = "Error starting container"
		}
		status = fmt.Sprintf("%s (%s)", st.Status, msg)
	}

	var extras []string
	if st.RestartCount > 0 {
		extras = append(extras, fmt.Sprintf("restarts: %d", st.RestartCount))
	}
	if st.LastError != "" && st.LastError != st.ErrorMessage {
		extras = append(extras, fmt.Sprintf("lastError: %s", st.LastError))
	}
	if len(extras) > 0 {
		status = fmt.Sprintf("%s (%s)", status, strings.Join(extras, ", "))
	}
	return status
}
