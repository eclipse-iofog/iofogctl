package deleteremotecontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deletecontroller "github.com/eclipse-iofog/iofogctl/internal/delete/controller"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor struct {
	namespace string
}

func NewExecutor(namespace string) (execute.Executor, error) {
	exe := &Executor{
		namespace: namespace,
	}
	return exe, nil
}

// GetName returns application name
func (exe *Executor) GetName() string {
	return "Delete Control Plane"
}

// Execute deletes application by deleting its associated application
func (exe *Executor) Execute() (err error) {
	// Get Control Plane
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}
	controlPlane, ok := baseControlPlane.(*rsc.RemoteControlPlane)
	if !ok {
		return util.NewError("Could not Convert Controller to Remote Controller")
	}

	controllers := controlPlane.GetControllers()
	executors := make([]execute.Executor, len(controllers))
	for idx := range controllers {
		controller := controllers[idx]
		exe := deletecontroller.NewRemoteExecutor(controlPlane, exe.namespace, controller.GetName())
		executors[idx] = exe
	}

	if err := runExecutors(executors); err != nil {
		return err
	}

	// Delete Control Plane in config
	ns.DeleteControlPlane()
	return config.Flush()
}

func runExecutors(executors []execute.Executor) error {
	if errs, _ := execute.ForParallel(executors); len(errs) > 0 {
		return execute.CoalesceErrors(errs)
	}
	return nil
}
