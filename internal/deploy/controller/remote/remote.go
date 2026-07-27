package deployremotecontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	deployremotecontrolplane "github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane/remote"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	namespace    string
	controlPlane *rsc.RemoteControlPlane
	controller   *rsc.RemoteController
}

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	controller, err := rsc.UnmarshallRemoteController(opt.Yaml)
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
	controlPlane, ok := baseControlPlane.(*rsc.RemoteControlPlane)
	if !ok {
		err = util.NewError("Could not convert Control Plane to Remote Control Plane")
		return
	}

	return NewExecutorWithoutParsing(opt.Namespace, controlPlane, &controller)
}

func newExecutor(namespace string, controlPlane *rsc.RemoteControlPlane, controller *rsc.RemoteController) *remoteExecutor {
	return &remoteExecutor{
		namespace:    namespace,
		controlPlane: controlPlane,
		controller:   controller,
	}
}

func (exe *remoteExecutor) GetName() string {
	return "Deploy Remote Controller"
}

func NewExecutorWithoutParsing(namespace string, controlPlane *rsc.RemoteControlPlane, controller *rsc.RemoteController) (exe execute.Executor, err error) {
	_, err = config.GetNamespace(namespace)
	if err != nil {
		return
	}

	if err := controller.Sanitize(); err != nil {
		return nil, err
	}

	if err := util.IsLowerAlphanumeric("Controller", controller.GetName()); err != nil {
		return nil, err
	}

	return newExecutor(namespace, controlPlane, controller), nil
}

func (exe *remoteExecutor) Execute() (err error) {
	if err = exe.controlPlane.SupportsControllerAddOn(); err != nil {
		return err
	}
	if err = exe.controlPlane.ValidateControllerAddOnDatabase(); err != nil {
		return err
	}
	if err = exe.controlPlane.ValidateControllerAddOn(exe.controller); err != nil {
		return err
	}
	if err = exe.controller.ValidateSSH(); err != nil {
		return err
	}
	if deployairgap.ControllerAirgapEnabled(exe.controlPlane, exe.controller) {
		if err = deployairgap.ValidateControllerAirgapRequirements(exe.controller); err != nil {
			return err
		}
	}

	applySystemMicroserviceDefaults(exe.controlPlane)

	edgelet, err := deployremotecontrolplane.DeployHostEdgelet(exe.controlPlane, exe.controller, exe.namespace)
	if err != nil {
		return err
	}

	translateOpts := deployremotecontrolplane.TranslateOptions{
		Name:      exe.namespace,
		Namespace: exe.namespace,
	}

	registryID, err := deployremotecontrolplane.DeployPrivateEdgeletRegistry(exe.controlPlane, edgelet, translateOpts)
	if err != nil {
		return err
	}

	if err := deployremotecontrolplane.DeployEdgeletControlPlane(exe.controlPlane, exe.controller, edgelet, translateOpts, registryID); err != nil {
		return err
	}

	endpoint, err := deployremotecontrolplane.ResolveControllerHostEndpoint(exe.controlPlane, exe.controller)
	if err != nil {
		return err
	}
	exe.controller.Endpoint = endpoint
	exe.controller.Created = util.NowUTC()

	if err := deployremotecontrolplane.DeployNextSystemAgent(exe.namespace, exe.controlPlane, exe.controller); err != nil {
		return err
	}

	if err := exe.controlPlane.UpdateController(exe.controller); err != nil {
		return err
	}

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	ns.SetControlPlane(exe.controlPlane)
	return config.Flush()
}

func applySystemMicroserviceDefaults(cp *rsc.RemoteControlPlane) {
	if cp.SystemMicroservices.Router.AMD64 == "" {
		cp.SystemMicroservices.Router.AMD64 = util.GetRouterImage()
	}
	if cp.SystemMicroservices.Router.ARM64 == "" {
		cp.SystemMicroservices.Router.ARM64 = util.GetRouterImage()
	}
	if cp.SystemMicroservices.Router.RISCV64 == "" {
		cp.SystemMicroservices.Router.RISCV64 = util.GetRouterImage()
	}
	if cp.SystemMicroservices.Router.ARM == "" {
		cp.SystemMicroservices.Router.ARM = util.GetRouterImage()
	}
	if cp.SystemMicroservices.Nats.AMD64 == "" {
		cp.SystemMicroservices.Nats.AMD64 = util.GetNatsImage()
	}
	if cp.SystemMicroservices.Nats.ARM64 == "" {
		cp.SystemMicroservices.Nats.ARM64 = util.GetNatsImage()
	}
	if cp.SystemMicroservices.Nats.RISCV64 == "" {
		cp.SystemMicroservices.Nats.RISCV64 = util.GetNatsImage()
	}
	if cp.SystemMicroservices.Nats.ARM == "" {
		cp.SystemMicroservices.Nats.ARM = util.GetNatsImage()
	}
}

func Validate(ctrl rsc.Controller) error {
	if err := util.IsLowerAlphanumeric("Controller", ctrl.GetName()); err != nil {
		return err
	}
	return nil
}
