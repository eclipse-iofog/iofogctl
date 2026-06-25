package describe

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func stringPtrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type serviceExecutor struct {
	namespace string
	name      string
	filename  string
}

func newServiceExecutor(namespace, name, filename string) *serviceExecutor {
	return &serviceExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *serviceExecutor) GetName() string {
	return exe.name
}

func (exe *serviceExecutor) Execute() error {
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Get service from Controller
	service, err := clt.GetService(exe.name)
	if err != nil {
		return err
	}

	// Convert tags to pointer
	var tags *[]string
	if len(service.Tags) > 0 {
		tags = &service.Tags
	}

	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.ServiceKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
			Tags:      tags,
		},
		Spec: rsc.ClusterService{
			Type:            service.Type,
			Resource:        service.Resource,
			TargetPort:      service.TargetPort,
			BridgePort:      service.BridgePort,
			DefaultBridge:   service.DefaultBridge,
			K8sType:         stringPtrValue(service.K8sType),
			ServiceEndpoint: service.ServiceEndpoint,
			ServicePort:     service.ServicePort,
		},
	}

	if exe.filename == "" {
		if err := util.Print(header); err != nil {
			return err
		}
	} else {
		if err := util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
