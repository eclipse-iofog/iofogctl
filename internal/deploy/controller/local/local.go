package deploylocalcontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type localExecutor struct {
	namespace string
	ctrl      *rsc.LocalController
	ctrlPlane *rsc.LocalControlPlane
}

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	controller, err := rsc.UnmarshallLocalController(opt.Yaml)
	if err != nil {
		return
	}

	if len(opt.Name) > 0 {
		controller.Name = opt.Name
	}

	if err = Validate(&controller); err != nil {
		return
	}

	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}

	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		return
	}
	controlPlane, ok := baseControlPlane.(*rsc.LocalControlPlane)
	if !ok {
		err = util.NewError("Could not convert Control Plane to Local Control Plane")
		return
	}

	return NewExecutorWithoutParsing(opt.Namespace, controlPlane, &controller)
}

func NewExecutorWithoutParsing(namespace string, controlPlane *rsc.LocalControlPlane, controller *rsc.LocalController) (exe execute.Executor, err error) {
	_, err = config.GetNamespace(namespace)
	if err != nil {
		return
	}
	if err := util.IsLowerAlphanumeric("Controller", controller.GetName()); err != nil {
		return nil, err
	}

	return &localExecutor{
		namespace: namespace,
		ctrl:      controller,
		ctrlPlane: controlPlane,
	}, nil
}

func (exe *localExecutor) GetName() string {
	return exe.ctrl.Name
}

func (exe *localExecutor) Execute() error {
	return util.NewInputError("LocalController add-on deploy via edgelet is not implemented yet")
}

func Validate(ctrl rsc.Controller) error {
	if err := util.IsLowerAlphanumeric("Controller", ctrl.GetName()); err != nil {
		return err
	}
	return nil
}
