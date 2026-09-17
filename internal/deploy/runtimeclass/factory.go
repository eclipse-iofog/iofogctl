package deployruntimeclass

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
	FullYAML  []byte
	Name      string
}

type remoteExecutor struct {
	namespace string
	name      string
	handler   string
}

func (exe *remoteExecutor) GetName() string {
	return exe.name
}

func (exe *remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying RuntimeClass %s", exe.GetName()))
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	_, err = clt.GetRuntimeClass(exe.name)
	if err != nil {
		if !clientutil.IsClientNotFoundError(err) {
			return err
		}
		return exe.createRuntimeClass(clt)
	}
	return exe.updateRuntimeClass(clt)
}

func (exe *remoteExecutor) createRuntimeClass(clt *client.Client) error {
	_, err := clt.CreateRuntimeClass(&client.RuntimeClassCreateRequest{
		Name:    exe.name,
		Handler: exe.handler,
	})
	return err
}

func (exe *remoteExecutor) updateRuntimeClass(clt *client.Client) error {
	_, err := clt.UpdateRuntimeClass(exe.name, &client.RuntimeClassUpdateRequest{
		Handler: exe.handler,
	})
	return err
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return exe, err
	}
	if len(ns.GetControllers()) == 0 {
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying RuntimeClasses")
	}

	if len(opt.Name) == 0 {
		return nil, util.NewInputError("Name must be specified")
	}
	if err := util.IsLowerAlphanumeric("RuntimeClass", opt.Name); err != nil {
		return nil, err
	}

	handler, err := parseRuntimeClassHandler(opt.FullYAML, opt.Yaml)
	if err != nil {
		return nil, err
	}

	return &remoteExecutor{
		namespace: opt.Namespace,
		name:      opt.Name,
		handler:   handler,
	}, nil
}

// parseRuntimeClassHandler reads handler from the document root (P10-RC-YAML-1).
// FullYAML is preferred; spec YAML is a fallback when dispatch stuffed handler there.
func parseRuntimeClassHandler(fullYAML, specYAML []byte) (string, error) {
	if handler := handlerFromYAML(fullYAML); handler != "" {
		return handler, nil
	}
	if handler := handlerFromYAML(specYAML); handler != "" {
		return handler, nil
	}
	return "", util.NewInputError("handler must be specified")
}

func handlerFromYAML(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	var doc struct {
		Handler string `yaml:"handler"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return ""
	}
	return strings.TrimSpace(doc.Handler)
}
