package deployagent

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type localExecutor struct {
	isSystem  bool
	namespace string
	agent     *rsc.LocalAgent
	edgelet   edgeletAgent
}

func newLocalExecutor(namespace string, agent *rsc.LocalAgent, isSystem bool) (*localExecutor, error) {
	if err := checkLocalAgentPortAvailable(isSystem); err != nil {
		return nil, err
	}
	if err := ensureLocalAgentHost(agent); err != nil {
		return nil, err
	}
	agent.Config = deployairgap.EnsureAgentConfig(agent.Config)
	return &localExecutor{
		isSystem:  isSystem,
		namespace: namespace,
		agent:     agent,
	}, nil
}

func (exe *localExecutor) getEdgelet() (edgeletAgent, error) {
	if exe.edgelet != nil {
		return exe.edgelet, nil
	}

	cfg := deployairgap.EdgeletInstallConfig(deployairgap.LocalEdgeletHostOS(), exe.agent.Config, exe.agent.Package)
	edgelet, err := install.NewLocalEdgelet(exe.agent.Name, exe.agent.UUID, cfg)
	if err != nil {
		return nil, err
	}
	exe.edgelet = edgelet
	return edgelet, nil
}

func (exe *localExecutor) ProvisionAgent() (string, error) {
	edgelet, err := exe.getEdgelet()
	if err != nil {
		return "", err
	}

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return "", err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return "", err
	}

	controllerEndpoint := exe.agent.GetControllerEndpoint()
	if controllerEndpoint == "" {
		controllerEndpoint, err = controlPlane.GetEndpoint()
		if err != nil {
			return "", err
		}
	}

	user := install.IofogUser(controlPlane.GetUser())
	user.Password = controlPlane.GetUser().GetRawPassword()
	return edgelet.Configure(controllerEndpoint, user)
}

func (exe *localExecutor) GetName() string {
	return exe.agent.Name
}

func (exe *localExecutor) Execute() error {
	exe.agent.Config = deployairgap.EnsureAgentConfig(exe.agent.Config)
	deployairgap.ResolveAgentDeployment(exe.agent.Config, exe.agent.Package.Container.Image)

	edgelet, err := exe.getEdgelet()
	if err != nil {
		return err
	}

	if err := customizeEdgeletProcedures(edgelet, exe.agent.Scripts); err != nil {
		return err
	}
	if err := applyEdgeletPackage(edgelet, exe.agent.Package); err != nil {
		return err
	}

	util.SpinStart("Installing edgelet")
	if err := edgelet.Bootstrap(); err != nil {
		return err
	}

	exe.agent.Host = exe.agent.GetHost()
	return nil
}
