package deletelocalcontrolplane

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployairgap "github.com/eclipse-iofog/iofogctl/internal/deploy/airgap"
	deploylocalcontrolplane "github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane/local"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func teardownEdgeletControlPlane(namespace string, cp *rsc.LocalControlPlane, name string) error {
	util.SpinStart("Deleting Control Plane")

	agentUUID := resolveSystemAgentUUID(namespace, name)
	edgelet, err := deploylocalcontrolplane.BuildEdgeletForTeardown(cp, namespace, name, agentUUID)
	if err != nil {
		return err
	}

	// Remove the system agent from the controller while it is still reachable.
	if agentUUID != "" {
		deleteSystemAgentFromController(namespace, name, agentUUID)
	}

	install.Verbose("Deprovisioning edgelet on " + name)
	if err := edgelet.Deprovision(); err != nil {
		util.PrintNotify(fmt.Sprintf("Could not deprovision edgelet: %v", err))
	}

	install.Verbose("Deleting edgelet control plane on " + name)
	if err := edgelet.DeleteControlPlane(); err != nil {
		util.PrintNotify(fmt.Sprintf("Could not delete edgelet control plane: %v", err))
	}

	clientutil.InvalidateAgentCache(namespace)

	install.Verbose("Uninstalling edgelet from " + name)
	if err := edgelet.Uninstall(true); err != nil {
		util.PrintNotify(fmt.Sprintf("Could not uninstall edgelet: %v", err))
	}

	if err := cleanLocalEdgeletMicroserviceContainers(cp); err != nil {
		util.PrintNotify(fmt.Sprintf("Could not clean microservice containers: %v", err))
	}

	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}
	_ = ns.DeleteAgent(name)
	clientutil.InvalidateAgentCache(namespace)
	return config.Flush()
}

func deleteSystemAgentFromController(namespace, name, agentUUID string) {
	ctrl, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		util.PrintNotify(fmt.Sprintf("Could not delete Agent %s from the Controller: %v", name, err))
		return
	}
	if err := ctrl.DeleteAgent(agentUUID); err != nil && !isIgnorableControllerAgentDeleteError(err) {
		util.PrintNotify(fmt.Sprintf("Could not delete Agent %s from the Controller: %v", name, err))
	}
}

func resolveSystemAgentUUID(namespace, name string) string {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return ""
	}
	if agent, err := ns.GetAgent(name); err == nil && agent.GetUUID() != "" {
		return agent.GetUUID()
	}
	backendAgents, err := clientutil.GetBackendAgents(namespace)
	if err != nil {
		return ""
	}
	for idx := range backendAgents {
		if backendAgents[idx].Name == name {
			return backendAgents[idx].UUID
		}
	}
	return ""
}

func cleanLocalEdgeletMicroserviceContainers(cp *rsc.LocalControlPlane) error {
	sys := cp.SystemAgent
	var cfg *rsc.AgentConfiguration
	var pkg rsc.Package
	if sys != nil {
		cfg = sys.AgentConfiguration
		pkg = sys.Package
	}
	cfg = deployairgap.EnsureAgentConfig(cfg)
	installCfg := deployairgap.EdgeletInstallConfig(deployairgap.LocalEdgeletHostOS(), cfg, pkg)
	client, err := install.NewLocalContainerClientFromEdgeletCfg(installCfg)
	if err != nil {
		return err
	}
	containers, err := client.ListContainers()
	if err != nil {
		return err
	}
	for idx := range containers {
		container := &containers[idx]
		for _, containerName := range container.Names {
			if strings.HasPrefix(containerName, "/iofog_") {
				if errClean := client.CleanContainerByID(container.ID); errClean != nil {
					util.PrintNotify(fmt.Sprintf("Could not clean Microservice container: %v", errClean))
				}
			}
		}
	}
	return nil
}

func isIgnorableControllerAgentDeleteError(err error) bool {
	if err == nil {
		return false
	}
	if util.IsNotFoundError(err) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") ||
		strings.Contains(msg, "notfound") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connect: connection refused") ||
		strings.Contains(msg, "network is unreachable") ||
		strings.Contains(msg, "no route to host")
}

func isLocalEdgeletControlPlane(cp rsc.ControlPlane) bool {
	localCP, ok := cp.(*rsc.LocalControlPlane)
	return ok && localCP.SystemAgent != nil
}

// ControlPlaneSystemAgentName returns the agent name tied to a local edgelet control plane.
func ControlPlaneSystemAgentName(cp *rsc.LocalControlPlane) string {
	if controllers := cp.GetControllers(); len(controllers) > 0 {
		return controllers[0].GetName()
	}
	return "local"
}
