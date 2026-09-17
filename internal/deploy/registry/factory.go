package deployregistry

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

const (
	registryTypeOCI = "oci"
	registryTypeHF  = "hf"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type remoteExecutor struct {
	namespace string
	registry  rsc.Registry
}

func (exe *remoteExecutor) GetName() string {
	if exe.registry.URL != nil {
		return *exe.registry.URL
	}
	return ""
}

func (exe *remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying registry %s", exe.GetName()))
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	if exe.registry.ID > 0 {
		return clt.UpdateRegistry(toUpdateRequest(exe.registry))
	}

	_, err = clt.CreateRegistry(toCreateRequest(exe.registry))
	return err
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return
	}

	if len(controlPlane.GetControllers()) == 0 {
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Applications")
	}

	var registry rsc.Registry
	if err = yaml.UnmarshalStrict(opt.Yaml, &registry); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	if registry.Private == nil {
		private := false
		registry.Private = &private
	}
	if registry.Type == nil || strings.TrimSpace(*registry.Type) == "" {
		t := registryTypeOCI
		registry.Type = &t
	} else {
		t := strings.ToLower(strings.TrimSpace(*registry.Type))
		registry.Type = &t
	}

	if err := validate(registry); err != nil {
		return nil, err
	}

	return &remoteExecutor{
		registry:  registry,
		namespace: opt.Namespace,
	}, nil
}

func toCreateRequest(reg rsc.Registry) *client.RegistryCreateRequest {
	req := &client.RegistryCreateRequest{
		Type: derefString(reg.Type),
		CA:   derefString(reg.CA),
	}
	if reg.URL != nil {
		req.URL = *reg.URL
	}
	if reg.Private != nil {
		req.IsPublic = !*reg.Private
	}
	if reg.Username != nil {
		req.Username = *reg.Username
	}
	if reg.Password != nil {
		req.Password = *reg.Password
	}
	if reg.Email != nil {
		req.Email = *reg.Email
	}
	if reg.Insecure != nil {
		req.Insecure = *reg.Insecure
	}
	return req
}

func toUpdateRequest(reg rsc.Registry) client.RegistryUpdateRequest {
	return client.RegistryUpdateRequest{
		URL:      reg.URL,
		IsPublic: invertBool(reg.Private),
		Username: reg.Username,
		Email:    reg.Email,
		Password: reg.Password,
		Type:     reg.Type,
		CA:       reg.CA,
		Insecure: reg.Insecure,
		ID:       reg.ID,
	}
}

func invertBool(p *bool) *bool {
	if p == nil {
		return nil
	}
	v := !*p
	return &v
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func validate(opt rsc.Registry) error {
	if opt.URL == nil || *opt.URL == "" {
		return util.NewInputError("URL cannot be empty")
	}

	regType := derefString(opt.Type)
	if regType != registryTypeOCI && regType != registryTypeHF {
		return util.NewInputError("type must be oci or hf")
	}

	if opt.Private != nil && *opt.Private {
		switch regType {
		case registryTypeHF:
			if opt.Password == nil || *opt.Password == "" {
				return util.NewInputError("Password cannot be empty for private hf registry")
			}
		default:
			if opt.Username == nil || *opt.Username == "" || opt.Password == nil || *opt.Password == "" {
				return util.NewInputError("Username and password are required for private oci registry")
			}
		}
	}

	return nil
}
