package describe

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type modelExecutor struct {
	namespace string
	name      string
	filename  string
}

type modelDescribeSpec struct {
	UUID       string   `yaml:"uuid,omitempty"`
	Repo       string   `yaml:"repo"`
	Revision   string   `yaml:"revision,omitempty"`
	RegistryID int      `yaml:"registryId"`
	Files      []string `yaml:"files,omitempty"`
	Format     string   `yaml:"format,omitempty"`
}

type linkedAgentsStatus struct {
	LinkedAgents []string `yaml:"linkedAgents,omitempty"`
}

func newModelExecutor(namespace, name, filename string) *modelExecutor {
	return &modelExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *modelExecutor) GetName() string {
	return exe.name
}

func (exe *modelExecutor) Execute() error {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	model, err := clt.GetModel(exe.name)
	if err != nil {
		return err
	}

	var linkedAgents []string
	link, err := clt.GetModelLink(exe.name)
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

	return printDescribeHeader(exe.filename, modelHeader(exe.namespace, exe.name, model, linkedAgents))
}

func modelHeader(namespace, name string, model *client.Model, linkedAgents []string) config.Header {
	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.ModelKind,
		Metadata: config.HeaderMetadata{
			Namespace: namespace,
			Name:      name,
		},
		Spec: modelDescribeSpec{
			UUID:       model.UUID,
			Repo:       model.Repo,
			Revision:   model.Revision,
			RegistryID: model.RegistryID,
			Files:      model.Files,
			Format:     model.Format,
		},
	}
	if len(linkedAgents) > 0 {
		header.Status = linkedAgentsStatus{LinkedAgents: linkedAgents}
	}
	return header
}

func printDescribeHeader(filename string, header config.Header) error {
	if filename == "" {
		return util.Print(header)
	}
	return util.FPrint(header, filename)
}
