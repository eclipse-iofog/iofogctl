package client

import (
	"errors"
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

const runtimeClassLink400Prefix = "cannot attach runtimeclass: RuntimeClass can only be linked to edgelet agents"

// AgentUUIDsFromNames resolves agent names to Controller fog UUIDs.
func AgentUUIDsFromNames(clt *client.Client, names []string) ([]string, error) {
	uuids := make([]string, 0, len(names))
	for _, name := range names {
		agentInfo, err := clt.GetAgentByName(name)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, agentInfo.UUID)
	}
	return uuids, nil
}

// AgentNamesFromUUIDs maps fog UUIDs to agent names. Unknown UUIDs are kept as-is.
func AgentNamesFromUUIDs(agents []client.AgentInfo, uuids []string) []string {
	if len(uuids) == 0 {
		return nil
	}
	nameByUUID := make(map[string]string, len(agents))
	for _, agent := range agents {
		if agent.UUID != "" {
			nameByUUID[agent.UUID] = agent.Name
		}
	}
	names := make([]string, 0, len(uuids))
	for _, uuid := range uuids {
		if name := nameByUUID[uuid]; name != "" {
			names = append(names, name)
			continue
		}
		names = append(names, uuid)
	}
	return names
}

// PrefixRuntimeClassLink400 adds a stable CLI prefix when Controller rejects a
// RuntimeClass link with HTTP 400 (non-edgelet fog). Other errors pass through
// so 409 bodies with blocking microservice UUIDs remain visible.
func PrefixRuntimeClassLink400(err error) error {
	if err == nil {
		return nil
	}
	var httpErr *client.HTTPError
	if errors.As(err, &httpErr) && httpErr.Code == 400 {
		return fmt.Errorf("%s\n%w", runtimeClassLink400Prefix, err)
	}
	return err
}
