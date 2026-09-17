package get

import (
	"strconv"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type microserviceTemplateExecutor struct {
	namespace string
	templates []client.MicroserviceTemplate
}

func newMicroserviceTemplateExecutor(namespace string) *microserviceTemplateExecutor {
	return &microserviceTemplateExecutor{namespace: namespace}
}

func (exe *microserviceTemplateExecutor) GetName() string {
	return ""
}

func (exe *microserviceTemplateExecutor) Execute() error {
	if err := exe.init(); err != nil {
		return err
	}
	printNamespace(exe.namespace)
	return print(generateMicroserviceTemplateOutput(exe.templates))
}

func (exe *microserviceTemplateExecutor) init() error {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		if rsc.IsNoControlPlaneError(err) {
			return nil
		}
		return err
	}

	resp, err := clt.ListMicroserviceTemplates()
	if err != nil {
		return err
	}
	exe.templates = resp.MicroserviceTemplates
	return nil
}

func generateMicroserviceTemplateOutput(templates []client.MicroserviceTemplate) [][]string {
	table := make([][]string, len(templates)+1)
	table[0] = []string{"NAME", "DESCRIPTION", "VARIABLES"}
	for idx, template := range templates {
		table[idx+1] = []string{
			template.Name,
			template.Description,
			strconv.Itoa(len(template.Variables)),
		}
	}
	return table
}
