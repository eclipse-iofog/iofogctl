package get

import (
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func NewExecutor(resourceType, namespace string, showDetached bool, resourceName string) (execute.Executor, error) {
	switch resourceType {
	case "namespaces":
		return newNamespaceExecutor(), nil
	case "all":
		return newAllExecutor(namespace), nil
	case "controllers":
		return newControllerExecutor(namespace), nil
	case "agents":
		return newAgentExecutor(namespace, showDetached), nil
	case "microservices":
		return newMicroserviceExecutor(namespace), nil
	case "system-microservices":
		return newSystemMicroserviceExecutor(namespace), nil
	case "application-templates":
		return newApplicationTemplateExecutor(namespace), nil
	case "applications":
		return newApplicationExecutor(namespace), nil
	case "system-applications":
		return newSystemApplicationExecutor(namespace), nil
	case "catalog":
		return newCatalogExecutor(namespace), nil
	case "registries":
		return newRegistryExecutor(namespace), nil
	case "volumes":
		return newVolumeExecutor(namespace), nil
	case "secrets":
		return newSecretExecutor(namespace), nil
	case "configmaps":
		return newConfigmapExecutor(namespace), nil
	case "services":
		return newServiceExecutor(namespace), nil
	case "volume-mounts":
		return newVolumeMountExecutor(namespace), nil
	case "certificates":
		return newCertificateExecutor(namespace), nil
	case "roles":
		return newRoleExecutor(namespace), nil
	case "rolebindings":
		return newRoleBindingExecutor(namespace), nil
	case "serviceaccounts":
		return newServiceAccountExecutor(namespace), nil
	case "nats-accounts":
		return newNatsAccountExecutor(namespace), nil
	case "nats-users":
		return newNatsUserExecutor(namespace), nil
	case "nats-account-rules":
		return newNatsAccountRuleExecutor(namespace), nil
	case "nats-user-rules":
		return newNatsUserRuleExecutor(namespace), nil
	case "auth-groups":
		return newAuthGroupsExecutor(namespace), nil
	case "auth-group":
		if resourceName == "" {
			return nil, util.NewInputError("get auth-group requires a name")
		}
		return newAuthGroupExecutor(namespace, resourceName), nil
	default:
		msg := "Unknown resource: '" + resourceType + "'"
		return nil, util.NewInputError(msg)
	}
}
