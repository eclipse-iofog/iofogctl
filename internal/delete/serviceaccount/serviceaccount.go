package deleteserviceaccount

import (
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor struct {
	namespace string
	appName   string
	name      string
}

// parseServiceAccountName returns (appName, name). Name must be "appName/name" (application-scoped).
func parseServiceAccountName(arg string) (appName, name string) {
	if idx := strings.Index(arg, "/"); idx >= 0 {
		return arg[:idx], arg[idx+1:]
	}
	return "", arg
}

func NewExecutor(namespace, name string) (execute.Executor, error) {
	appName, saName := parseServiceAccountName(name)
	return &Executor{
		namespace: namespace,
		appName:   appName,
		name:      saName,
	}, nil
}

func (exe *Executor) GetName() string {
	if exe.appName != "" {
		return exe.appName + "/" + exe.name
	}
	return exe.name
}

func (exe *Executor) Execute() error {
	if exe.appName == "" {
		return util.NewInputError("ServiceAccount is application-scoped: use APPLICATION_NAME/SERVICE_ACCOUNT_NAME (e.g. myapp/my-sa)")
	}
	util.SpinStart("Deleting ServiceAccount")
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}
	return clt.DeleteServiceAccount(exe.appName, exe.name)
}
