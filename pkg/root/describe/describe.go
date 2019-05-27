package describe

import (
	"os"
	yaml "gopkg.in/yaml.v2"
	"github.com/eclipse-iofog/cli/pkg/config"
	"github.com/eclipse-iofog/cli/pkg/util"
)
type describe struct {
	configManager *config.Manager
}

func new() *describe {
	d := &describe{}
	d.configManager = config.NewManager()
	return d
}

func (describe *describe) execute(resource, namespace, name string) error {

	switch resource {

	case "namespace":
		namespace, err := describe.configManager.GetNamespace(name)
		if err != nil {
			return err
		}
		if err = print(namespace); err != nil {
			return err
		}

	case "controller":
		controller, err := describe.configManager.GetController(namespace, name)
		if err != nil {
			return err
		}
		if err = print(controller); err != nil {
			return err
		}

	case "agent":
		agent, err := describe.configManager.GetAgent(namespace, name)
		if err != nil {
			return err
		}
		if err = print(agent); err != nil {
			return err
		}

	case "microservice":
		//microservices, err := describe.configManager.GetMicroservice(namespace, name)

	default:
		msg := "Unknown resource: '" + resource + "'"
		return util.NewInputError(msg)
	}

	return nil
}

func print(obj interface{}) error {
	marshal, err := yaml.Marshal(&obj)
	if err != nil {
		return err
	} 
	_, err = os.Stdout.Write(marshal)
	if err != nil {
		return err
	} 
	return nil
}