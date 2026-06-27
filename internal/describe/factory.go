package describe

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Resource   string
	Name       string
	Namespace  string
	Filename   string
	IsDetached bool
	Version    string
}

func NewExecutor(opt *Options) (execute.Executor, error) {
	switch opt.Resource {
	case "namespace":
		return newNamespaceExecutor(opt.Namespace, opt.Filename), nil
	case "controlplane":
		return newControlPlaneExecutor(opt.Namespace, opt.Filename), nil
	case "controller":
		return newControllerExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "agent":
		return newAgentExecutor(opt.Namespace, opt.Name, opt.Filename, opt.IsDetached), nil
	case "registry":
		return newRegistryExecutor(opt.Namespace, opt.Name, opt.Filename)
	case "agent-config":
		return newAgentConfigExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "microservice":
		return newMicroserviceExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "system-microservice":
		return newSystemMicroserviceExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "application-template":
		return newApplicationTemplateExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "application":
		return newApplicationExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "volume":
		return newVolumeExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "secret":
		return newSecretExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "configmap":
		return newConfigMapExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "service":
		return newServiceExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "volume-mount":
		return newVolumeMountExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "certificate":
		return newCertificateExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "role":
		return newRoleExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "rolebinding":
		return newRoleBindingExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "serviceaccount":
		return newServiceAccountExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	default:
		return nil, util.NewInputError(fmt.Sprintf("Unknown resources: %s", opt.Resource))
	}
}
