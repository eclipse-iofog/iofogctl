package deletemicroservicetemplate

import (
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor struct {
	namespace string
	name      string
}

func NewExecutor(namespace, name string) (execute.Executor, error) {
	exe := &Executor{
		namespace: namespace,
		name:      name,
	}

	return exe, nil
}

func (exe *Executor) GetName() string {
	return exe.name
}

func (exe *Executor) Execute() error {
	util.SpinStart("Deleting Microservice Template")
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	return clt.DeleteMicroserviceTemplate(exe.name)
}
