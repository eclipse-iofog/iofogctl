package get

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type runtimeClassExecutor struct {
	namespace      string
	runtimeClasses []client.RuntimeClass
}

func newRuntimeClassExecutor(namespace string) *runtimeClassExecutor {
	return &runtimeClassExecutor{namespace: namespace}
}

func (exe *runtimeClassExecutor) GetName() string {
	return ""
}

func (exe *runtimeClassExecutor) Execute() error {
	if err := exe.init(); err != nil {
		return err
	}
	printNamespace(exe.namespace)
	return print(generateRuntimeClassOutput(exe.runtimeClasses))
}

func (exe *runtimeClassExecutor) init() error {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		if rsc.IsNoControlPlaneError(err) {
			return nil
		}
		return err
	}

	resp, err := clt.ListRuntimeClasses()
	if err != nil {
		return err
	}
	exe.runtimeClasses = resp.RuntimeClasses
	return nil
}

func generateRuntimeClassOutput(runtimeClasses []client.RuntimeClass) [][]string {
	table := make([][]string, len(runtimeClasses)+1)
	table[0] = []string{"NAME", "HANDLER"}
	for idx, runtimeClass := range runtimeClasses {
		table[idx+1] = []string{
			runtimeClass.Name,
			runtimeClass.Handler,
		}
	}
	return table
}
