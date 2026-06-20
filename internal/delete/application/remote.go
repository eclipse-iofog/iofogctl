package deleteapplication

import (
	"errors"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor struct {
	namespace string
	name      string
	client    *client.Client
	flow      *client.FlowInfo
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

func (exe *Executor) init() (err error) {
	exe.client, err = clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return
	}
	return
}

// Execute deletes application by deleting its associated flow
func (exe *Executor) Execute() (err error) {
	util.SpinStart("Deleting Application")
	if err := exe.init(); err != nil {
		return err
	}

	err = exe.client.DeleteApplication(exe.name)
	// If notfound error, try legacy
	notFoundError := &client.NotFoundError{}
	if errors.As(err, &notFoundError) {
		return exe.deleteLegacy()
	}
	return err
}
