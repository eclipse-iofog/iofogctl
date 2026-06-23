package deployremotecontrolplane

import (
	"context"
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/auth"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/internal/trust"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type remoteControlPlaneExecutor struct {
	controlPlane *rsc.RemoteControlPlane
	namespace    string
	name         string
}

type hostExecutor struct {
	namespace string
	name      string
	cp        *rsc.RemoteControlPlane
	ctrl      *rsc.RemoteController
}

func (exe hostExecutor) Execute() error {
	return deployRemoteControlPlaneHost(exe.namespace, exe.name, exe.cp, exe.ctrl)
}

func (exe hostExecutor) GetName() string {
	return exe.ctrl.Name
}

func (exe remoteControlPlaneExecutor) Execute() (err error) {
	util.SpinStart(fmt.Sprintf("Deploying controlplane %s", exe.GetName()))

	if ca := exe.controlPlane.GetTrustCA(); ca != "" {
		if err := trust.StoreCA(exe.namespace, ca); err != nil {
			return err
		}
	}

	if exe.controlPlane.Airgap {
		if err := deployairgap.ValidateControlPlaneAirgapRequirements(exe.controlPlane); err != nil {
			return err
		}
	}

	hostExecutors := make([]execute.Executor, len(exe.controlPlane.Controllers))
	for idx := range exe.controlPlane.Controllers {
		ctrl := &exe.controlPlane.Controllers[idx]
		hostExecutors[idx] = hostExecutor{
			namespace: exe.namespace,
			name:      exe.name,
			cp:        exe.controlPlane,
			ctrl:      ctrl,
		}
	}
	if err := runExecutors(hostExecutors); err != nil {
		return err
	}

	endpoint, err := exe.controlPlane.GetEndpoint()
	if err != nil {
		return err
	}

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	ns.SetControlPlane(exe.controlPlane)
	if err := config.Flush(); err != nil {
		return err
	}

	if err := trust.WaitForControllerAPI(context.Background(), exe.namespace, endpoint); err != nil {
		return err
	}

	if err := auth.EnsureIofogUserEmbedded(context.Background(), exe.namespace, endpoint, auth.EmbeddedAuthSpec{
		Mode:      exe.controlPlane.Auth.Mode,
		Bootstrap: exe.controlPlane.Auth.Bootstrap,
		User:      &exe.controlPlane.IofogUser,
	}); err != nil {
		return err
	}

	if err := deployControllerRegistry(exe.namespace, exe.controlPlane); err != nil {
		return err
	}

	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}
	if err := install.DeployGlobalCertificates(clt, globalCertificatesFromCP(exe.controlPlane)); err != nil {
		return err
	}

	if err := deployRemoteSystemAgents(exe.namespace, exe.controlPlane); err != nil {
		return err
	}

	ns.SetControlPlane(exe.controlPlane)
	return config.Flush()
}

func (exe remoteControlPlaneExecutor) GetName() string {
	return exe.name
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	_, err = config.GetNamespace(opt.Namespace)
	if err != nil {
		return
	}

	controlPlane, err := rsc.UnmarshallRemoteControlPlane(opt.Yaml)
	if err != nil {
		return
	}

	applySystemMicroserviceDefaults(&controlPlane)

	return remoteControlPlaneExecutor{
		controlPlane: &controlPlane,
		namespace:    opt.Namespace,
		name:         opt.Name,
	}, nil
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
