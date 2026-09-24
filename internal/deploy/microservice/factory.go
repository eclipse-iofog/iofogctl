package deploymicroservice

import (
	"fmt"
	"strings"

	apps "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace      string
	Yaml           []byte
	Name           string
	PatchModel     bool
	PatchKnowledge bool
}

type remoteExecutor struct {
	namespace        string
	microservice     apps.Microservice
	name             string
	patchModel       bool
	patchKnowledge   bool
	catalog          client.MicroserviceCatalog
	knowledgeCatalog client.KnowledgeCatalog
}

func (exe *remoteExecutor) GetName() string {
	return exe.name
}

func (exe *remoteExecutor) Execute() error {
	if exe.patchModel || exe.patchKnowledge {
		return exe.patchCatalogs()
	}

	util.SpinStart(fmt.Sprintf("Deploying microservice %s", exe.GetName()))
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}

	// Check Controller exists
	if len(controlPlane.GetControllers()) == 0 {
		return util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Applications")
	}
	endpoint, err := controlPlane.GetEndpoint()
	if err != nil {
		return err
	}

	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	controller := apps.IofogController{
		Endpoint:     endpoint,
		Email:        controlPlane.GetUser().Email,
		Password:     controlPlane.GetUser().Password,
		Token:        clt.GetAccessToken(),
		RefreshToken: clt.GetRefreshToken(),
	}

	appName, msvcName, err := clientutil.ParseFQName(exe.name, "Microservice")
	if err != nil {
		return err
	}

	return apps.DeployMicroservice(controller, &exe.microservice, appName, msvcName, apps.WithAPIVersion(util.GetCliApiVersion()))
}

func (exe *remoteExecutor) patchCatalogs() error {
	clt, uuid, err := exe.lookupMicroservice()
	if err != nil {
		return err
	}
	// Models first to match --patch-model then --patch-knowledge flag order.
	if exe.patchModel {
		if err := exe.patchModels(clt, uuid); err != nil {
			return err
		}
	}
	if exe.patchKnowledge {
		if err := exe.patchKnowledgeCatalog(clt, uuid); err != nil {
			return err
		}
	}
	return nil
}

func (exe *remoteExecutor) lookupMicroservice() (*client.Client, string, error) {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return nil, "", err
	}
	appName, msvcName, err := clientutil.ParseFQName(exe.name, "Microservice")
	if err != nil {
		return nil, "", err
	}
	msvc, err := clt.GetMicroserviceByName(appName, msvcName)
	if err != nil {
		return nil, "", err
	}
	return clt, msvc.UUID, nil
}

func (exe *remoteExecutor) patchModels(clt *client.Client, uuid string) error {
	util.SpinStart(fmt.Sprintf("Patching models for microservice %s", exe.GetName()))
	return clt.PatchMicroserviceModels(uuid, exe.catalog)
}

func (exe *remoteExecutor) patchKnowledgeCatalog(clt *client.Client, uuid string) error {
	util.SpinStart(fmt.Sprintf("Patching knowledge for microservice %s", exe.GetName()))
	return clt.PatchMicroserviceKnowledge(uuid, exe.knowledgeCatalog)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	if _, err = config.GetNamespace(opt.Namespace); err != nil {
		return exe, err
	}
	// Unmarshal file
	var microservice apps.Microservice
	if err = yaml.UnmarshalStrict(opt.Yaml, &microservice); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	name := resolveMicroserviceName(opt.Name, opt.Yaml)

	if opt.PatchModel || opt.PatchKnowledge {
		remote := &remoteExecutor{
			namespace:      opt.Namespace,
			microservice:   microservice,
			name:           name,
			patchModel:     opt.PatchModel,
			patchKnowledge: opt.PatchKnowledge,
		}
		if opt.PatchModel {
			catalog, err := parseMicroserviceModels(opt.Yaml)
			if err != nil {
				return nil, err
			}
			if catalog == nil {
				return nil, util.NewInputError("--patch-model requires spec.models")
			}
			remote.catalog = toClientCatalog(*catalog)
		}
		if opt.PatchKnowledge {
			catalog, err := parseMicroserviceKnowledge(opt.Yaml)
			if err != nil {
				return nil, err
			}
			if catalog == nil {
				return nil, util.NewInputError("--patch-knowledge requires spec.knowledge")
			}
			remote.knowledgeCatalog = toClientKnowledgeCatalog(*catalog)
		}
		return remote, nil
	}

	return &remoteExecutor{
		namespace:    opt.Namespace,
		microservice: microservice,
		name:         name,
	}, nil
}

func resolveMicroserviceName(optName string, yamlBytes []byte) string {
	name := optName
	if strings.Contains(name, "/") {
		return name
	}
	var spec struct {
		Application string `yaml:"application,omitempty"`
	}
	if err := yaml.Unmarshal(yamlBytes, &spec); err != nil {
		return name
	}
	if spec.Application != "" {
		return spec.Application + "/" + name
	}
	return name
}

func parseMicroserviceModels(raw []byte) (*apps.MicroserviceCatalog, error) {
	var spec struct {
		Models *apps.MicroserviceCatalog `yaml:"models"`
	}
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		return nil, util.NewUnmarshalError(err.Error())
	}
	return spec.Models, nil
}

func parseMicroserviceKnowledge(raw []byte) (*apps.KnowledgeCatalog, error) {
	var spec struct {
		Knowledge *apps.KnowledgeCatalog `yaml:"knowledge"`
	}
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		return nil, util.NewUnmarshalError(err.Error())
	}
	return spec.Knowledge, nil
}

func toClientCatalog(in apps.MicroserviceCatalog) client.MicroserviceCatalog {
	out := client.MicroserviceCatalog{
		BindPath:    in.BindPath,
		Permissions: in.Permissions,
	}
	if len(in.Items) == 0 {
		return out
	}
	out.Items = make([]client.MicroserviceCatalogItem, len(in.Items))
	for i, item := range in.Items {
		out.Items[i] = client.MicroserviceCatalogItem{Name: item.Name}
	}
	return out
}

func toClientKnowledgeCatalog(in apps.KnowledgeCatalog) client.KnowledgeCatalog {
	out := client.KnowledgeCatalog{
		BindPath:    in.BindPath,
		Permissions: in.Permissions,
	}
	if len(in.Items) == 0 {
		return out
	}
	out.Items = make([]client.KnowledgeCatalogItem, len(in.Items))
	for i, item := range in.Items {
		out.Items[i] = client.KnowledgeCatalogItem{Name: item.Name}
	}
	return out
}
