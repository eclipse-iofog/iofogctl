package deployagent

import (
	"context"
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	namespace string
	agent     *rsc.RemoteAgent
	edgelet   edgeletAgent
}

func newRemoteExecutor(namespace string, agent *rsc.RemoteAgent) *remoteExecutor {
	return &remoteExecutor{
		namespace: namespace,
		agent:     agent,
	}
}

func (exe *remoteExecutor) GetName() string {
	return exe.agent.Name
}

func (exe *remoteExecutor) getEdgelet() (edgeletAgent, error) {
	if exe.edgelet != nil {
		return exe.edgelet, nil
	}

	exe.agent.Config = deployairgap.EnsureAgentConfig(exe.agent.Config)
	cfg := deployairgap.EdgeletInstallConfig("linux", exe.agent.Config, exe.agent.Package)

	edgelet, err := install.NewRemoteEdgelet(
		exe.agent.SSH.User,
		exe.agent.Host,
		exe.agent.SSH.Port,
		exe.agent.SSH.KeyFile,
		exe.agent.Name,
		exe.agent.UUID,
		cfg,
	)
	if err != nil {
		return nil, err
	}
	exe.edgelet = edgelet
	return edgelet, nil
}

func (exe *remoteExecutor) ProvisionAgent() (string, error) {
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
			return "", util.NewError("Failed to retrieve Controller endpoint!")
		}
	}

	user := install.IofogUser(controlPlane.GetUser())
	user.Password = controlPlane.GetUser().GetRawPassword()
	opt, err := clientutil.ControllerClientOptions(context.Background(), exe.namespace, controllerEndpoint)
	if err != nil {
		return "", err
	}
	return edgelet.Configure(controllerEndpoint, user, opt)
}

func (exe *remoteExecutor) Execute() (err error) {
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil || len(controlPlane.GetControllers()) == 0 {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return err
	}

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

	var pendingNativeAirgap *deployairgap.AgentAirgapResult
	ctx := context.Background()

	if exe.agent.Airgap {
		var remoteControlPlane *rsc.RemoteControlPlane
		if remoteCP, ok := controlPlane.(*rsc.RemoteControlPlane); ok {
			remoteControlPlane = remoteCP
		}

		isInitial, err := deployairgap.IsInitialDeployment(exe.namespace)
		if err != nil {
			return fmt.Errorf("failed to determine deployment type: %w", err)
		}
		if remoteControlPlane == nil {
			isInitial = false
		}

		airgapPlan, err := deployairgap.PrepareAgentAirgap(exe.namespace, exe.agent, remoteControlPlane, isInitial)
		if err != nil {
			return fmt.Errorf("airgap deployment requires valid configuration: %w", err)
		}

		if deployairgap.IsNativeDeployment(airgapPlan.Options.DeploymentType) {
			plan := airgapPlan
			pendingNativeAirgap = &plan
			remoteBinPath, err := deployairgap.TransferAgentAirgapBinary(ctx, exe.namespace, exe.agent.Host, exe.agent.Package.Version, &exe.agent.SSH, airgapPlan.Platform)
			if err != nil {
				return fmt.Errorf("failed to transfer edgelet binary: %w", err)
			}
			if err := edgelet.SetAirgap(remoteBinPath); err != nil {
				return fmt.Errorf("failed to configure edgelet airgap binary: %w", err)
			}
		} else if len(airgapPlan.ImageList) > 0 {
			if err := deployairgap.TransferAgentAirgapImages(ctx, exe.namespace, exe.agent.Host, &exe.agent.SSH, airgapPlan.Platform, airgapPlan.Options, airgapPlan.ImageList); err != nil {
				return fmt.Errorf("failed to transfer airgap images: %w", err)
			}
		}

		if err := stageRemoteWasmAirgap(ctx, exe.namespace, exe.agent.Host, exe.agent.Config, &exe.agent.SSH, exe.agent.Package, edgelet); err != nil {
			return err
		}
	}

	if err := edgelet.PrepareWasm(ctx, exe.namespace); err != nil {
		return err
	}

	if err := edgelet.Bootstrap(); err != nil {
		return err
	}

	if pendingNativeAirgap != nil && len(pendingNativeAirgap.ImageList) > 0 {
		if err := deployairgap.TransferAgentAirgapImages(ctx, exe.namespace, exe.agent.Host, &exe.agent.SSH, pendingNativeAirgap.Platform, pendingNativeAirgap.Options, pendingNativeAirgap.ImageList); err != nil {
			return fmt.Errorf("failed to transfer airgap images: %w", err)
		}
	}

	return nil
}

func ValidateRemoteAgent(agent *rsc.RemoteAgent) error {
	if err := util.IsLowerAlphanumeric("Agent", agent.Name); err != nil {
		return err
	}
	if agent.Name == iofog.VanillaRouterAgentName {
		return util.NewInputError(fmt.Sprintf("%s is a reserved name and cannot be used for an Agent", iofog.VanillaRouterAgentName))
	}
	if (agent.Host != "localhost" && agent.Host != "127.0.0.1") && (agent.Host == "" || agent.SSH.User == "" || agent.SSH.KeyFile == "") {
		return util.NewInputError("For Agents you must specify non-empty values for host, user, and keyfile")
	}
	return rsc.ValidateAgentPackage("Agent", agent.Package, agent.Config)
}
