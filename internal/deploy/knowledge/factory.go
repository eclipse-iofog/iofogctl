package deployknowledge

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

// knowledgeSpec is the kind: Knowledge spec. registry is an alias of registryId (P11-YAML-1).
type knowledgeSpec struct {
	Repo       string   `yaml:"repo"`
	Revision   string   `yaml:"revision,omitempty"`
	RegistryID int      `yaml:"registryId,omitempty"`
	Registry   string   `yaml:"registry,omitempty"`
	Files      []string `yaml:"files,omitempty"`
	Format     string   `yaml:"format,omitempty"`
}

type remoteExecutor struct {
	namespace  string
	name       string
	spec       knowledgeSpec
	registryID int
}

func (exe *remoteExecutor) GetName() string {
	return exe.name
}

func (exe *remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying knowledge %s", exe.GetName()))
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	_, err = clt.GetKnowledge(exe.name)
	if err != nil {
		if !clientutil.IsClientNotFoundError(err) {
			return err
		}
		return exe.createKnowledge(clt)
	}
	return exe.updateKnowledge(clt)
}

func (exe *remoteExecutor) createKnowledge(clt *client.Client) error {
	_, err := clt.CreateKnowledge(&client.KnowledgeCreateRequest{
		Name:       exe.name,
		Repo:       exe.spec.Repo,
		RegistryID: exe.registryID,
		Revision:   exe.spec.Revision,
		Files:      exe.spec.Files,
		Format:     exe.spec.Format,
	})
	return err
}

func (exe *remoteExecutor) updateKnowledge(clt *client.Client) error {
	repo := exe.spec.Repo
	revision := exe.spec.Revision
	format := exe.spec.Format
	registryID := exe.registryID
	_, err := clt.UpdateKnowledge(exe.name, &client.KnowledgeUpdateRequest{
		Repo:       &repo,
		Revision:   &revision,
		RegistryID: &registryID,
		Files:      exe.spec.Files,
		Format:     &format,
	})
	return err
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return exe, err
	}
	if len(ns.GetControllers()) == 0 {
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Knowledge")
	}

	spec, err := parseKnowledgeSpec(opt.Yaml)
	if err != nil {
		return nil, err
	}
	if len(opt.Name) > 0 {
		if err := util.IsLowerAlphanumeric("Knowledge", opt.Name); err != nil {
			return nil, err
		}
	} else {
		return nil, util.NewInputError("Name must be specified")
	}

	registryID, err := resolveKnowledgeRegistryID(spec)
	if err != nil {
		return nil, err
	}
	if err := validateKnowledgeSpec(spec); err != nil {
		return nil, err
	}

	return &remoteExecutor{
		namespace:  opt.Namespace,
		name:       opt.Name,
		spec:       spec,
		registryID: registryID,
	}, nil
}

func parseKnowledgeSpec(raw []byte) (knowledgeSpec, error) {
	var spec knowledgeSpec
	if err := yaml.UnmarshalStrict(raw, &spec); err != nil {
		return spec, util.NewUnmarshalError(err.Error())
	}
	return spec, nil
}

func resolveKnowledgeRegistryID(spec knowledgeSpec) (int, error) {
	alias := strings.TrimSpace(spec.Registry)
	var fromAlias int
	var hasAlias bool
	if alias != "" {
		id, err := clientutil.ResolveRegistryID(alias)
		if err != nil {
			return 0, err
		}
		fromAlias = id
		hasAlias = true
	}
	if spec.RegistryID > 0 && hasAlias && spec.RegistryID != fromAlias {
		return 0, util.NewInputError("spec.registry and spec.registryId must refer to the same registry")
	}
	if hasAlias {
		return fromAlias, nil
	}
	if spec.RegistryID > 0 {
		return spec.RegistryID, nil
	}
	return 0, util.NewInputError("spec.registry or spec.registryId must be specified")
}

func validateKnowledgeSpec(spec knowledgeSpec) error {
	if strings.TrimSpace(spec.Repo) == "" {
		return util.NewInputError("repo must be specified")
	}
	return nil
}
