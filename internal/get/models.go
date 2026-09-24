package get

import (
	"strconv"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type modelExecutor struct {
	namespace string
}

func newModelExecutor(namespace string) *modelExecutor {
	return &modelExecutor{namespace: namespace}
}

func (exe *modelExecutor) GetName() string {
	return ""
}

func (exe *modelExecutor) Execute() error {
	printNamespace(exe.namespace)
	table, err := generateModelsOutput(exe.namespace)
	if err != nil {
		return err
	}
	return print(table)
}

func generateModelsOutput(namespace string) ([][]string, error) {
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		if rsc.IsNoControlPlaneError(err) {
			return generateModelOutput(nil), nil
		}
		return nil, err
	}

	resp, err := clt.ListModels()
	if err != nil {
		return nil, err
	}
	return generateModelOutput(resp.Models), nil
}

func generateModelOutput(models []client.Model) [][]string {
	table := make([][]string, len(models)+1)
	table[0] = []string{"NAME", "REPO", "REGISTRY_ID", "FORMAT", "REVISION"}
	for idx, model := range models {
		table[idx+1] = []string{
			model.Name,
			model.Repo,
			strconv.Itoa(model.RegistryID),
			model.Format,
			model.Revision,
		}
	}
	return table
}
