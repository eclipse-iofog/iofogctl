package deployapplication

import (
	"fmt"

	apps "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
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

type remoteExecutor struct {
	namespace   string
	application interface{}
	name        string
}

func (exe *remoteExecutor) GetName() string {
	return exe.name
}

func (exe *remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying Application %s", exe.GetName()))

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
	return apps.DeployApplication(controller, exe.application, exe.name)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	if _, err = config.GetNamespace(opt.Namespace); err != nil {
		return exe, err
	}
	// Unmarshal file
	var application interface{}
	if err = yaml.UnmarshalStrict(opt.Yaml, &application); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	return &remoteExecutor{
		namespace:   opt.Namespace,
		application: &application,
		name:        opt.Name,
	}, nil
}
