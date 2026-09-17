package describe

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type runtimeClassExecutor struct {
	namespace string
	name      string
	filename  string
}

func newRuntimeClassExecutor(namespace, name, filename string) *runtimeClassExecutor {
	return &runtimeClassExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *runtimeClassExecutor) GetName() string {
	return exe.name
}

func (exe *runtimeClassExecutor) Execute() error {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	runtimeClass, err := clt.GetRuntimeClass(exe.name)
	if err != nil {
		return err
	}

	var linkedAgents []string
	link, err := clt.GetRuntimeClassLink(exe.name)
	if err != nil {
		return err
	}
	if link != nil && len(link.FogUUIDs) > 0 {
		agents, listErr := clt.ListAgents(client.ListAgentsRequest{})
		if listErr != nil {
			return listErr
		}
		linkedAgents = clientutil.AgentNamesFromUUIDs(agents.Agents, link.FogUUIDs)
	}

	return printDescribeHeader(exe.filename, runtimeClassHeader(exe.namespace, exe.name, runtimeClass.Handler, linkedAgents))
}

func runtimeClassHeader(namespace, name, handler string, linkedAgents []string) config.Header {
	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.RuntimeClassKind,
		Metadata: config.HeaderMetadata{
			Namespace: namespace,
			Name:      name,
		},
		Handler: handler,
	}
	if len(linkedAgents) > 0 {
		header.Status = linkedAgentsStatus{LinkedAgents: linkedAgents}
	}
	return header
}
