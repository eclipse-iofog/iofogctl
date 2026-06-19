package get

import (
	"strconv"

	// "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type serviceExecutor struct {
	namespace string
}

func newServiceExecutor(namespace string) *serviceExecutor {
	a := &serviceExecutor{}
	a.namespace = namespace
	return a
}

func (exe *serviceExecutor) Execute() error {
	printNamespace(exe.namespace)
	table, err := generateServicesOutput(exe.namespace)
	if err != nil {
		return err
	}
	return print(table)
}

func (exe *serviceExecutor) GetName() string {
	return ""
}

func generateServicesOutput(namespace string) ([][]string, error) {
	// Init remote resources
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return nil, err
	}

	serviceList, err := clt.ListServices()
	if err != nil {
		return nil, err
	}

	// Generate table and headers
	table := make([][]string, len(serviceList.Services)+1)
	headers := []string{
		"SERVICE",
		"TYPE",
		"RESOURCE",
		"TARGET PORT",
		// "SERVICE PORT",
		// "K8S TYPE",
		"BRIDGE PORT",
		// "DEFAULT BRIDGE",
		// "SERVICE ENDPOINT",
		"STATUS",
	}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, service := range serviceList.Services {
		// Handle empty values
		// servicePort := ""
		// if service.ServicePort != 0 {
		// 	servicePort = strconv.Itoa(service.ServicePort)
		// }

		// k8sType := ""
		// if service.K8sType != "" {
		// 	k8sType = service.K8sType
		// }

		// serviceEndpoint := ""
		// if service.ServiceEndpoint != "" {
		// 	serviceEndpoint = service.ServiceEndpoint
		// }

		row := []string{
			service.Name,
			service.Type,
			service.Resource,
			strconv.Itoa(service.TargetPort),
			// servicePort,
			// k8sType,
			strconv.Itoa(service.BridgePort),
			// service.DefaultBridge,
			// serviceEndpoint,
			service.ProvisioningStatus,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	return table, nil
}
