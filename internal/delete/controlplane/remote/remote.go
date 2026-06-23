package deleteremotecontrolplane

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/internal/trust"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor struct {
	namespace string
}

type hostTeardownExecutor struct {
	namespace string
	cp        *rsc.RemoteControlPlane
	ctrl      *rsc.RemoteController
}

func NewExecutor(namespace string) (execute.Executor, error) {
	return &Executor{namespace: namespace}, nil
}

func (exe *Executor) GetName() string {
	return "Delete Control Plane"
}

func (exe *Executor) Execute() (err error) {
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
		return util.NewError("Could not convert Control Plane to Remote Control Plane")
	}

	controllers := controlPlane.GetControllers()
	executors := make([]execute.Executor, len(controllers))
	for idx := range controllers {
		controller, ok := controllers[idx].(*rsc.RemoteController)
		if !ok {
			return util.NewInternalError("Could not convert Controller to Remote Controller")
		}
		executors[idx] = newHostTeardownExecutor(exe.namespace, controlPlane, controller)
	}

	if err := runExecutors(executors); err != nil {
		return err
	}

	for idx := range controllers {
		_ = ns.DeleteAgent(controllers[idx].GetName())
	}

	if err := trust.RemoveCA(exe.namespace); err != nil {
		return err
	}

	ns.DeleteControlPlane()
	return config.Flush()
}

func newHostTeardownExecutor(namespace string, cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController) *hostTeardownExecutor {
	return &hostTeardownExecutor{
		namespace: namespace,
		cp:        cp,
		ctrl:      ctrl,
	}
}

func (exe *hostTeardownExecutor) GetName() string {
	return exe.ctrl.Name
}

func (exe *hostTeardownExecutor) Execute() error {
	return TeardownRemoteControllerHost(exe.namespace, exe.cp, exe.ctrl)
}

func runExecutors(executors []execute.Executor) error {
	if errs, _ := execute.ForParallel(executors); len(errs) > 0 {
		return execute.CoalesceErrors(errs)
	}
	return nil
}
