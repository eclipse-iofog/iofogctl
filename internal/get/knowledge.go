package get

import (
	"strconv"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type knowledgeExecutor struct {
	namespace string
}

func newKnowledgeExecutor(namespace string) *knowledgeExecutor {
	return &knowledgeExecutor{namespace: namespace}
}

func (exe *knowledgeExecutor) GetName() string {
	return ""
}

func (exe *knowledgeExecutor) Execute() error {
	printNamespace(exe.namespace)
	table, err := generateKnowledgeOutput(exe.namespace)
	if err != nil {
		return err
	}
	return print(table)
}

func generateKnowledgeOutput(namespace string) ([][]string, error) {
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		if rsc.IsNoControlPlaneError(err) {
			return formatKnowledgeOutput(nil), nil
		}
		return nil, err
	}

	resp, err := clt.ListKnowledge()
	if err != nil {
		return nil, err
	}
	return formatKnowledgeOutput(resp.Knowledge), nil
}

func formatKnowledgeOutput(items []client.Knowledge) [][]string {
	table := make([][]string, len(items)+1)
	table[0] = []string{"NAME", "REPO", "REGISTRY_ID", "FORMAT", "REVISION"}
	for idx, item := range items {
		table[idx+1] = []string{
			item.Name,
			item.Repo,
			strconv.Itoa(item.RegistryID),
			item.Format,
			item.Revision,
		}
	}
	return table
}
