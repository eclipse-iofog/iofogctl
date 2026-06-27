package reconcileservice

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const serviceHubReadyMessage = "Service hub provisioned; edge bridge listeners on tagged fogs may still be converging."

type executor struct {
	namespace string
	name      string
}

func NewExecutor(namespace, name string) execute.Executor {
	return &executor{namespace: namespace, name: name}
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) Execute() error {
	util.SpinStart(fmt.Sprintf("Reconciling service %s", exe.name))
	if err := clientutil.ReconcileServiceByNameAndWait(exe.namespace, exe.name); err != nil {
		return err
	}
	util.PrintInfo(serviceHubReadyMessage)
	return nil
}
