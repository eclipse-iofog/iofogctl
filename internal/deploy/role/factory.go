package deployrole

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type executor struct {
	role      rsc.Role
	namespace string
	name      string
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying role %s", exe.GetName()))
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	if _, err = clt.GetRole(exe.name); err != nil {
		return exe.createRole(clt)
	}
	return exe.updateRole(clt)
}

func (exe *executor) createRole(clt *client.Client) error {
	req := &client.RoleCreateRequest{
		Name:  exe.name,
		Kind:  exe.role.Kind,
		Rules: toClientRules(exe.role.Rules),
	}
	_, err := clt.CreateRole(req)
	return err
}

func (exe *executor) updateRole(clt *client.Client) error {
	name := exe.name
	req := client.RoleUpdateRequest{
		Name:  &name,
		Kind:  exe.role.Kind,
		Rules: toClientRules(exe.role.Rules),
	}
	_, err := clt.UpdateRole(exe.name, &req)
	return err
}

func toClientRules(rules []rsc.RBACRule) []client.RBACRule {
	out := make([]client.RBACRule, len(rules))
	for i, r := range rules {
		out[i] = client.RBACRule{
			APIGroups:     r.APIGroups,
			Resources:     r.Resources,
			Verbs:         r.Verbs,
			ResourceNames: r.ResourceNames,
		}
	}
	return out
}

func NewExecutor(opt Options) (execute.Executor, error) {
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}
	if len(ns.GetControllers()) == 0 {
		return nil, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Roles")
	}

	var role rsc.Role
	if err = yaml.UnmarshalStrict(opt.Yaml, &role); err != nil {
		return nil, util.NewUnmarshalError(err.Error())
	}

	name := opt.Name
	if name == "" {
		name = role.Name
	}
	if name == "" {
		return nil, util.NewInputError("Name must be specified")
	}

	return &executor{
		namespace: opt.Namespace,
		role:      role,
		name:      name,
	}, nil
}
