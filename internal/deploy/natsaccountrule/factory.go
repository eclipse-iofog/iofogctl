package deploynatsaccountrule

import (
	"bytes"
	"errors"
	"fmt"

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

type executor struct {
	namespace string
	name      string
	specYAML  []byte // spec-only payload sent to API (controller expects rule payload, not full doc)
	fullYAML  []byte // full doc used only for validation
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying NATS account rule %s", exe.GetName()))
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(exe.specYAML)
	_, err = clt.UpdateNatsAccountRuleFromYAML(exe.name, reader)
	if err == nil {
		return nil
	}
	notFoundError := &client.NotFoundError{}
	if !errors.As(err, &notFoundError) {
		return err
	}

	reader = bytes.NewReader(exe.specYAML)
	_, err = clt.CreateNatsAccountRuleFromYAML(reader)
	return err
}

func NewExecutor(opt Options) (execute.Executor, error) {
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return nil, err
	}
	if len(controlPlane.GetControllers()) == 0 {
		return nil, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying NATS account rules")
	}

	// Validate YAML shape using strict decode against expected top-level fields.
	var validateDoc struct {
		APIVersion string                 `yaml:"apiVersion"`
		Kind       config.Kind            `yaml:"kind"`
		Metadata   config.HeaderMetadata  `yaml:"metadata"`
		Spec       map[string]interface{} `yaml:"spec"`
	}
	if err = yaml.UnmarshalStrict(opt.FullYAML, &validateDoc); err != nil {
		return nil, util.NewUnmarshalError(err.Error())
	}

	// Controller expects full document: apiVersion, kind, metadata, and spec (metadata and spec required).
	specPayload := opt.Yaml
	var specMap map[interface{}]interface{}
	if err = yaml.Unmarshal(opt.Yaml, &specMap); err == nil {
		doc := map[interface{}]interface{}{
			"apiVersion": util.GetCliApiVersion(),
			"kind":       "NatsAccountRule",
			"metadata":   map[interface{}]interface{}{"name": opt.Name},
			"spec":       specMap,
		}
		if specPayload, err = yaml.Marshal(doc); err != nil {
			specPayload = opt.Yaml
		}
	}

	return &executor{
		namespace: opt.Namespace,
		name:      opt.Name,
		specYAML:  specPayload,
		fullYAML:  opt.FullYAML,
	}, nil
}
