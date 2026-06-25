package connectremotecontrolplane

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	connectcontrolplane "github.com/eclipse-iofog/iofogctl/internal/connect/controlplane"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type remoteExecutor struct {
	controlPlane *rsc.RemoteControlPlane
	namespace    string
	caFile       string
	caB64        string
}

func NewManualExecutor(namespace, name, endpoint, email, password, caFile, caB64 string) (execute.Executor, error) {
	fmtEndpoint, err := formatEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	host := fmtEndpoint.Hostname()
	formatedEndpoint, err := util.GetControllerEndpoint(fmtEndpoint.String())
	if err != nil {
		return nil, err
	}
	controlPlane := &rsc.RemoteControlPlane{
		IofogUser: rsc.IofogUser{Email: email, Password: password},
		Controllers: []rsc.RemoteController{
			{
				Name:     name,
				Endpoint: formatedEndpoint,
				Host:     host,
			},
		},
	}

	return newRemoteExecutor(controlPlane, namespace, caFile, caB64), nil
}

func NewExecutor(namespace, name string, yaml []byte, kind config.Kind, caFile, caB64 string) (execute.Executor, error) {
	// Read the input file
	controlPlane, err := rsc.UnmarshallRemoteControlPlane(yaml)
	if err != nil {
		return nil, err
	}

	if err := validate(&controlPlane); err != nil {
		return nil, err
	}

	// In YAML, the endpoint will come through Host variable
	for _, baseController := range controlPlane.GetControllers() {
		controller, ok := baseController.(*rsc.RemoteController)
		if !ok {
			return nil, util.NewError("Could not convert Controller to Remote Controller")
		}
		fmtEndpoint, err := formatEndpoint(controlPlane.Controllers[0].Host)
		if err != nil {
			return nil, err
		}
		host := fmtEndpoint.Hostname()
		formatedEndpoint, err := util.GetControllerEndpoint(fmtEndpoint.String())
		if err != nil {
			return nil, err
		}
		controller.Endpoint = formatedEndpoint
		controller.Host = host
		if err := controlPlane.UpdateController(controller); err != nil {
			return nil, err
		}
	}

	return newRemoteExecutor(&controlPlane, namespace, caFile, caB64), nil
}

func newRemoteExecutor(controlPlane *rsc.RemoteControlPlane, namespace, caFile, caB64 string) *remoteExecutor {
	r := &remoteExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
		caFile:       caFile,
		caB64:        caB64,
	}
	return r
}

func (exe *remoteExecutor) GetName() string {
	return "Remote Control Plane"
}

func (exe *remoteExecutor) Execute() (err error) {
	// Establish connection
	controllers := exe.controlPlane.GetControllers()
	if len(controllers) == 0 {
		return util.NewError("Control Plane in Namespace " + exe.namespace + " has no Controllers. Try deploying a Control Plane to this Namespace.")
	}
	endpoint, err := exe.controlPlane.GetEndpoint()
	if err != nil {
		return err
	}
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	if err := connectcontrolplane.PrepareTrust(exe.namespace, exe.controlPlane, exe.caFile, exe.caB64); err != nil {
		return err
	}
	err = connectcontrolplane.Connect(exe.controlPlane, endpoint, exe.namespace, ns)
	if err != nil {
		return err
	}

	ns.SetControlPlane(exe.controlPlane)
	return config.Flush()
}

func formatEndpoint(endpoint string) (*url.URL, error) {
	// Ensure protocol
	if !strings.Contains(endpoint, "://") {
		endpoint = fmt.Sprintf("http://%s", endpoint)
	}
	URL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	// Ensure port for http
	if !strings.Contains(URL.Host, ":") && URL.Scheme != "https" {
		URL.Host += ":51121"
	}
	return URL, nil
}

func validate(controlPlane rsc.ControlPlane) (err error) {
	// Validate user
	user := controlPlane.GetUser()
	if user.Email == "" {
		return util.NewInputError("To connect, Control Plane Iofog User must contain non-empty value in email field")
	}
	// Validate Controllers
	if len(controlPlane.GetControllers()) == 0 {
		err = util.NewInputError("Control Plane must have at least one Controller instance specified.")
		return
	}
	for _, ctrl := range controlPlane.GetControllers() {
		if err = util.IsLowerAlphanumeric("Controller", ctrl.GetName()); err != nil {
			return
		}
	}

	return
}
