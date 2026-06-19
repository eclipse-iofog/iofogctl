package deleteedgeresource

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type executor struct {
	namespace string
	name      string
	version   string
}

func (exe executor) GetName() string {
	return fmt.Sprintf("%s/%s", exe.name, exe.version)
}

func (exe executor) Execute() (err error) {
	if _, err = config.GetNamespace(exe.namespace); err != nil {
		return
	}

	// Connect to Controller
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return
	}

	if err = clt.DeleteEdgeResource(exe.name, exe.version); err != nil {
		return
	}
	return
}

func NewExecutor(namespace, name, version string) (exe execute.Executor) {
	return executor{
		namespace: namespace,
		name:      name,
		version:   version,
	}
}
