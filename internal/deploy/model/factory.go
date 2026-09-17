package deploymodel

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

// modelSpec is the kind: Model spec. registry is an alias of registryId (P10-MODEL-YAML-1).
type modelSpec struct {
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
	spec       modelSpec
	registryID int
}

func (exe *remoteExecutor) GetName() string {
	return exe.name
}

func (exe *remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying model %s", exe.GetName()))
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	_, err = clt.GetModel(exe.name)
	if err != nil {
		if !clientutil.IsClientNotFoundError(err) {
			return err
		}
		return exe.createModel(clt)
	}
	return exe.updateModel(clt)
}

func (exe *remoteExecutor) createModel(clt *client.Client) error {
	_, err := clt.CreateModel(&client.ModelCreateRequest{
		Name:       exe.name,
		Repo:       exe.spec.Repo,
		RegistryID: exe.registryID,
		Revision:   exe.spec.Revision,
		Files:      exe.spec.Files,
		Format:     exe.spec.Format,
	})
	return err
}

func (exe *remoteExecutor) updateModel(clt *client.Client) error {
	repo := exe.spec.Repo
	revision := exe.spec.Revision
	format := exe.spec.Format
	registryID := exe.registryID
	_, err := clt.UpdateModel(exe.name, &client.ModelUpdateRequest{
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
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Models")
	}

	spec, err := parseModelSpec(opt.Yaml)
	if err != nil {
		return nil, err
	}
	if len(opt.Name) > 0 {
		if err := util.IsLowerAlphanumeric("Model", opt.Name); err != nil {
			return nil, err
		}
	} else {
		return nil, util.NewInputError("Name must be specified")
	}

	registryID, err := resolveModelRegistryID(spec)
	if err != nil {
		return nil, err
	}
	if err := validateModelSpec(spec); err != nil {
		return nil, err
	}

	return &remoteExecutor{
		namespace:  opt.Namespace,
		name:       opt.Name,
		spec:       spec,
		registryID: registryID,
	}, nil
}

func parseModelSpec(raw []byte) (modelSpec, error) {
	var spec modelSpec
	if err := yaml.UnmarshalStrict(raw, &spec); err != nil {
		return spec, util.NewUnmarshalError(err.Error())
	}
	return spec, nil
}

func resolveModelRegistryID(spec modelSpec) (int, error) {
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

func validateModelSpec(spec modelSpec) error {
	if strings.TrimSpace(spec.Repo) == "" {
		return util.NewInputError("repo must be specified")
	}
	return nil
}
