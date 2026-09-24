package describe

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type microserviceTemplateExecutor struct {
	namespace string
	name      string
	filename  string
}

type microserviceTemplateVariable struct {
	Key          string      `yaml:"key"`
	Description  string      `yaml:"description,omitempty"`
	DefaultValue interface{} `yaml:"defaultValue,omitempty"`
}

type microserviceTemplateSpec struct {
	Description  string                         `yaml:"description,omitempty"`
	Variables    []microserviceTemplateVariable `yaml:"variables,omitempty"`
	Microservice interface{}                    `yaml:"microservice,omitempty"`
}

func newMicroserviceTemplateExecutor(namespace, name, filename string) *microserviceTemplateExecutor {
	return &microserviceTemplateExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *microserviceTemplateExecutor) GetName() string {
	return exe.name
}

func (exe *microserviceTemplateExecutor) Execute() error {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	template, err := clt.GetMicroserviceTemplate(exe.name)
	if err != nil {
		return err
	}

	return printDescribeHeader(exe.filename, microserviceTemplateHeader(exe.namespace, exe.name, template))
}

func microserviceTemplateHeader(namespace, name string, template *client.MicroserviceTemplate) config.Header {
	spec := microserviceTemplateSpec{
		Description:  template.Description,
		Microservice: template.Microservice,
	}
	if len(template.Variables) > 0 {
		spec.Variables = make([]microserviceTemplateVariable, 0, len(template.Variables))
		for _, variable := range template.Variables {
			spec.Variables = append(spec.Variables, microserviceTemplateVariable{
				Key:          variable.Key,
				Description:  variable.Description,
				DefaultValue: variable.DefaultValue,
			})
		}
	}
	return config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.MicroserviceTemplateKind,
		Metadata: config.HeaderMetadata{
			Namespace: namespace,
			Name:      name,
		},
		Spec: spec,
	}
}
