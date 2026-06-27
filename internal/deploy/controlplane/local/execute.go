package deploylocalcontrolplane

import (
	"context"
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/auth"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/internal/trust"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Namespace string
	Yaml      []byte
	FullYAML  []byte
	Name      string
}

type localControlPlaneExecutor struct {
	controlPlane *rsc.LocalControlPlane
	namespace    string
	name         string
}

func (exe localControlPlaneExecutor) Execute() (err error) {
	util.SpinStart(fmt.Sprintf("Deploying controlplane %s", exe.GetName()))

	if ca := exe.controlPlane.GetTrustCA(); ca != "" {
		if err := trust.StoreCA(exe.namespace, ca); err != nil {
			return err
		}
	}

	edgelet, err := installHostEdgelet(exe.controlPlane, exe.name)
	if err != nil {
		return err
	}

	translateOpts := TranslateOptions{
		Name:      exe.name,
		Namespace: exe.namespace,
	}

	registryID, err := deployPrivateEdgeletRegistry(exe.controlPlane, edgelet, translateOpts)
	if err != nil {
		return err
	}

	if err := deployEdgeletControlPlane(exe.controlPlane, edgelet, translateOpts, registryID); err != nil {
		return err
	}

	endpoint := resolveLocalControlPlaneEndpoint(exe.controlPlane)
	if err := persistLocalControllerStub(exe.controlPlane, exe.name, endpoint); err != nil {
		return err
	}
	if err := rsc.BackfillConsoleURL(exe.controlPlane); err != nil {
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

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	ns.SetControlPlane(exe.controlPlane)
	if err := config.Flush(); err != nil {
		return err
	}

	if err := deployLocalSystemAgent(exe.namespace, exe.controlPlane, exe.name, edgelet); err != nil {
		return err
	}

	return nil
}

func (exe localControlPlaneExecutor) GetName() string {
	return exe.name
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	_, err = config.GetNamespace(opt.Namespace)
	if err != nil {
		return
	}

	if len(opt.FullYAML) > 0 {
		if err = rsc.ValidateLocalControlPlaneMetadata(opt.FullYAML); err != nil {
			return
		}
	}
	controlPlane, err := rsc.UnmarshallLocalControlPlane(opt.Yaml)
	if err != nil {
		return
	}

	return localControlPlaneExecutor{
		controlPlane: &controlPlane,
		namespace:    opt.Namespace,
		name:         opt.Name,
	}, nil
}
