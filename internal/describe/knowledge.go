package describe

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type knowledgeExecutor struct {
	namespace string
	name      string
	filename  string
}

type knowledgeDescribeSpec struct {
	UUID       string   `yaml:"uuid,omitempty"`
	Repo       string   `yaml:"repo"`
	Revision   string   `yaml:"revision,omitempty"`
	RegistryID int      `yaml:"registryId"`
	Files      []string `yaml:"files,omitempty"`
	Format     string   `yaml:"format,omitempty"`
}

func newKnowledgeExecutor(namespace, name, filename string) *knowledgeExecutor {
	return &knowledgeExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *knowledgeExecutor) GetName() string {
	return exe.name
}

func (exe *knowledgeExecutor) Execute() error {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	knowledge, err := clt.GetKnowledge(exe.name)
	if err != nil {
		return err
	}

	var linkedAgents []string
	link, err := clt.GetKnowledgeLink(exe.name)
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

	return printDescribeHeader(exe.filename, knowledgeHeader(exe.namespace, exe.name, knowledge, linkedAgents))
}

func knowledgeHeader(namespace, name string, knowledge *client.Knowledge, linkedAgents []string) config.Header {
	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.KnowledgeKind,
		Metadata: config.HeaderMetadata{
			Namespace: namespace,
			Name:      name,
		},
		Spec: knowledgeDescribeSpec{
			UUID:       knowledge.UUID,
			Repo:       knowledge.Repo,
			Revision:   knowledge.Revision,
			RegistryID: knowledge.RegistryID,
			Files:      knowledge.Files,
			Format:     knowledge.Format,
		},
	}
	if len(linkedAgents) > 0 {
		header.Status = linkedAgentsStatus{LinkedAgents: linkedAgents}
	}
	return header
}
