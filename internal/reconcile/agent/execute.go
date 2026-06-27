package reconcileagent

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

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
	util.SpinStart(fmt.Sprintf("Reconciling agent %s platform", exe.name))
	return clientutil.ReconcileAgentByNameAndWait(exe.namespace, exe.name)
}
