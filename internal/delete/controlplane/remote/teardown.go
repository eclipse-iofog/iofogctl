package deleteremotecontrolplane

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployremotecontrolplane "github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane/remote"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type edgeletHostTeardown interface {
	Deprovision() error
	DeleteControlPlane() error
	Uninstall(removeData bool) error
}

var (
	buildRemoteEdgeletFn = deployremotecontrolplane.BuildRemoteEdgelet
	deleteSystemAgentFn  = deleteSystemAgentFromController
	edgeletDeprovisionFn = func(e edgeletHostTeardown) error { return e.Deprovision() }
	edgeletDeleteCPFn    = func(e edgeletHostTeardown) error { return e.DeleteControlPlane() }
	edgeletUninstallFn   = func(e edgeletHostTeardown) error { return e.Uninstall(true) }
)

// TeardownRemoteControllerHost removes edgelet and control plane workloads from one remote host.
func TeardownRemoteControllerHost(namespace string, cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController) error {
	util.SpinStart("Deleting Control Plane on " + ctrl.Name)

	agentUUID := resolveSystemAgentUUID(namespace, ctrl.Name)
	edgelet, err := buildRemoteEdgeletFn(cp, ctrl, agentUUID)
	if err != nil {
		return err
	}

	if agentUUID != "" {
		deleteSystemAgentFn(namespace, ctrl.Name, agentUUID)
	}

	install.Verbose("Deprovisioning edgelet on " + ctrl.Name)
	if err := edgeletDeprovisionFn(edgelet); err != nil {
		util.PrintNotify(fmt.Sprintf("Could not deprovision edgelet: %v", err))
	}

	install.Verbose("Deleting edgelet control plane on " + ctrl.Name)
	if err := edgeletDeleteCPFn(edgelet); err != nil {
		util.PrintNotify(fmt.Sprintf("Could not delete edgelet control plane: %v", err))
	}

	clientutil.InvalidateAgentCache(namespace)

	install.Verbose("Uninstalling edgelet from " + ctrl.Name)
	if err := edgeletUninstallFn(edgelet); err != nil {
		util.PrintNotify(fmt.Sprintf("Could not uninstall edgelet: %v", err))
	}

	return nil
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
