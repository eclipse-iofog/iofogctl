package deployserviceaccount

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
	serviceAccount rsc.ServiceAccount
	namespace      string
	appName        string
	name           string
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying serviceaccount %s", exe.GetName()))
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	if _, err = clt.GetServiceAccount(exe.appName, exe.name); err != nil {
		return exe.createServiceAccount(clt)
	}
	return exe.updateServiceAccount(clt)
}

func (exe *executor) createServiceAccount(clt *client.Client) error {
	req := &client.ServiceAccountCreateRequest{
		Name:            exe.name,
		ApplicationName: exe.appName,
		RoleRef:         toClientRoleRef(exe.serviceAccount.RoleRef),
	}
	_, err := clt.CreateServiceAccount(req)
	return err
}

func (exe *executor) updateServiceAccount(clt *client.Client) error {
	req := client.ServiceAccountUpdateRequest{
		Name:    exe.name,
		RoleRef: toClientRoleRef(exe.serviceAccount.RoleRef),
	}
	_, err := clt.UpdateServiceAccount(exe.appName, exe.name, &req)
	return err
}

func toClientRoleRef(ref rsc.RoleRef) client.RoleRef {
	return client.RoleRef{
		Kind:     ref.Kind,
		Name:     ref.Name,
		APIGroup: ref.APIGroup,
	}
}

func NewExecutor(opt Options) (execute.Executor, error) {
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}
	if len(ns.GetControllers()) == 0 {
		return nil, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying ServiceAccounts")
	}

	var sa rsc.ServiceAccount
	if err = yaml.UnmarshalStrict(opt.Yaml, &sa); err != nil {
		return nil, util.NewUnmarshalError(err.Error())
	}

	name := opt.Name
	if name == "" {
		name = sa.Name
	}
	if name == "" {
		return nil, util.NewInputError("Name must be specified")
	}
	if sa.ApplicationName == "" {
		return nil, util.NewInputError("ServiceAccount must have applicationName (application-scoped)")
	}

	return &executor{
		namespace:      opt.Namespace,
		serviceAccount: sa,
		appName:        sa.ApplicationName,
		name:           name,
	}, nil
}
