package logs

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func NewExecutor(resourceType, namespace, name string, logConfig *LogTailConfig) (execute.Executor, error) {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}
	// Use default config if nil
	if logConfig == nil {
		logConfig = DefaultLogTailConfig()
	}
	switch resourceType {
	case "controller":
		baseControlPlane, err := ns.GetControlPlane()
		if err != nil {
			return nil, util.NewError("Could not get Control Plane for namespace " + namespace)
		}
		switch controlPlane := baseControlPlane.(type) {
		case *rsc.KubernetesControlPlane:
			return newKubernetesControllerExecutor(controlPlane, namespace, name), nil
		case *rsc.RemoteControlPlane:
			return newRemoteControllerExecutor(controlPlane, namespace, name), nil
		case *rsc.LocalControlPlane:
			return newLocalControllerExecutor(controlPlane, namespace, name), nil
		}
	case "agent":
		return newAgentExecutor(namespace, name, logConfig), nil
	case "microservice":
		if len(ns.GetControllers()) == 0 {
			return nil, util.NewError("No Controllers found in namespace " + namespace)
		}
		return newRemoteMicroserviceExecutor(namespace, name, logConfig), nil
	}
	msg := "Unknown resource: '" + resourceType + "'"
	return nil, util.NewInputError(msg)
}
