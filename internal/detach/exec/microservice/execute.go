package detachexecmicroservice

import (
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Name      string
	Namespace string
	Msvc      *client.MicroserviceInfo
}

type executor struct {
	name      string
	namespace string
	msvc      *client.MicroserviceInfo
}

func NewExecutor(opt Options) execute.Executor {
	return &executor{
		name:      opt.Name,
		namespace: opt.Namespace,
		msvc:      opt.Msvc,
	}
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) Execute() error {
	util.SpinStart("Detaching Exec Session to Microservice")

	// Init client
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	appName, msvcName, err := clientutil.ParseFQName(exe.name, "Microservice")
	if err != nil {
		return err
	}

	exe.msvc, err = clt.GetMicroserviceByName(appName, msvcName)
	isSystem := false
	if err != nil {
		// Check if error indicates application not found
		if strings.Contains(err.Error(), "Invalid application id") {
			// Try system application
			exe.msvc, err = clt.GetSystemMicroserviceByName(appName, msvcName)
			if err != nil {
				return err
			}
			isSystem = true
		} else {
			// Return other types of errors
			return err
		}
	}

	// Attach Exec Session to Microservice
	req := client.DetachExecMicroserviceRequest{
		UUID: exe.msvc.UUID,
	}
	if isSystem {
		if err := clt.DetachExecSystemMicroservice(&req); err != nil {
			return err
		}
	} else {
		if err := clt.DetachExecMicroservice(&req); err != nil {
			return err
		}
	}

	return nil
}
