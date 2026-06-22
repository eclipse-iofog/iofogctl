package deletevolumemount

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

// GetName returns application name
func (exe *Executor) GetName() string {
	return exe.name
}

// Execute deletes application by deleting its associated application
func (exe *Executor) Execute() error {
	util.SpinStart("Deleting Volume Mount")
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	return clt.DeleteVolumeMount(exe.name)
}
