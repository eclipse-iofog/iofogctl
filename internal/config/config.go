package config

import (
	"fmt"
	"github.com/eclipse-iofog/cli/pkg/util"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// struct that file is unmarshalled into
var conf configuration

// Name of file
var configFilename string

// DefaultFilename export
const DefaultFilename = ".iofog.yaml"

// Init initializes config and unmarshalls the file
func Init(filename string) {
	// Read the file and unmarshall contents
	if filename == "" {
		// Find home directory.
		home, err := homedir.Dir()
		util.Check(err)

		configFilename = home + "/" + DefaultFilename
	} else {
		configFilename = filename
	}
	// Check file exists
	if _, err := os.Stat(configFilename); os.IsNotExist(err) {
		// Create default file
		defaultData := []byte(`namespaces:
- name: default
  controllers: []
  agents: []
  microservices: []`)
		err := ioutil.WriteFile(configFilename, defaultData, 0644)
		util.Check(err)
	}

	viper.SetConfigFile(configFilename)

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	util.Check(err)
	fmt.Println("Using config file:", viper.ConfigFileUsed())

	// Unmarshall the file
	err = viper.Unmarshal(&conf)
	util.Check(err)
}

// GetNamespaces export
func GetNamespaces() (namespaces []Namespace) {
	for _, ns := range conf.Namespaces {
		newNamespace := Namespace{Name: ns.Name}
		namespaces = append(namespaces, newNamespace)
	}
	return
}

// GetAgents export
func GetAgents(namespace string) ([]Agent, error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == namespace {
			return ns.Agents, nil
		}
	}
	return nil, util.NewNotFoundError(namespace)
}

// GetControllers export
func GetControllers(namespace string) ([]Controller, error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == namespace {
			return ns.Controllers, nil
		}
	}
	return nil, util.NewNotFoundError(namespace)
}

// GetMicroservices export
func GetMicroservices(namespace string) ([]Microservice, error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == namespace {
			return ns.Microservices, nil
		}
	}
	return nil, util.NewNotFoundError(namespace)
}

// GetNamespace export
func GetNamespace(name string) (namespace Namespace, err error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == name {
			namespace = ns
			return
		}
	}
	err = util.NewNotFoundError(name)
	return
}

// GetController export
func GetController(namespace, name string) (controller Controller, err error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == namespace {
			for _, ctrl := range ns.Controllers {
				if ctrl.Name == name {
					controller = ctrl
					return
				}
			}
		}
	}
	err = util.NewNotFoundError(namespace + "/" + name)
	return
}

// GetAgent export
func GetAgent(namespace, name string) (agent Agent, err error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == namespace {
			for _, ag := range ns.Agents {
				if ag.Name == name {
					agent = ag
					return
				}
			}
		}
	}
	err = util.NewNotFoundError(namespace + "/" + name)
	return
}

// GetMicroservice export
func GetMicroservice(namespace, name string) (microservice Microservice, err error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == namespace {
			for _, ms := range ns.Microservices {
				if ms.Name == name {
					microservice = ms
					return
				}
			}
		}
	}
	err = util.NewNotFoundError(namespace + "/" + name)
	return
}

// AddNamespace export
func AddNamespace(name string) error {
	// Check collision
	_, err := GetNamespace(name)
	if err == nil {
		return util.NewConflictError(name)
	}

	newNamespace := Namespace{Name: name}
	conf.Namespaces = append(conf.Namespaces, newNamespace)
	if err := updateFile(); err != nil {
		return err
	}
	return nil
}

// AddController export
func AddController(namespace string, controller Controller) error {
	_, err := GetController(namespace, controller.Name)
	if err == nil {
		return util.NewConflictError(namespace + "/" + controller.Name)
	}

	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}

	// Append the controller
	ns.Controllers = append(ns.Controllers, controller)

	// Write to file
	if err := updateFile(); err != nil {
		return err
	}

	return nil
}

// AddAgent export
func AddAgent(namespace string, agent Agent) error {
	_, err := GetAgent(namespace, agent.Name)
	if err == nil {
		return util.NewConflictError(namespace + "/" + agent.Name)
	}

	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}

	// Append the controller
	ns.Agents = append(ns.Agents, agent)

	// Write to file
	if err := updateFile(); err != nil {
		return err
	}

	return nil
}

// AddMicroservice export
func AddMicroservice(namespace string, microservice Microservice) error {
	_, err := GetMicroservice(namespace, microservice.Name)
	if err == nil {
		return util.NewConflictError(namespace + "/" + microservice.Name)
	}

	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}

	// Append the controller
	ns.Microservices = append(ns.Microservices, microservice)

	// Write to file
	if err := updateFile(); err != nil {
		return err
	}

	return nil
}

// DeleteNamespace export
func DeleteNamespace(name string) error {
	ns, err := getNamespace(name)
	if err != nil {
		return err
	}

	hasAgents := len(ns.Agents) > 0
	hasControllers := len(ns.Controllers) > 0
	hasMicroservices := len(ns.Microservices) > 0

	if hasAgents || hasControllers || hasMicroservices {
		return util.NewInputError("Namespace " + name + " not empty")
	}

	// Delete namespace
	for idx := range conf.Namespaces {
		if conf.Namespaces[idx].Name == name {
			conf.Namespaces = append(conf.Namespaces[:idx], conf.Namespaces[idx+1:]...)
		}
	}
	if err := updateFile(); err != nil {
		return err
	}

	return nil
}

// DeleteController export
func DeleteController(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}

	for idx := range ns.Controllers {
		if ns.Controllers[idx].Name == name {
			ns.Controllers = append(ns.Controllers[:idx], ns.Controllers[idx+1:]...)
			if err := updateFile(); err != nil {
				return err
			}
			return nil
		}
	}

	return util.NewNotFoundError(namespace + "/" + name)
}

// DeleteAgent export
func DeleteAgent(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}

	for idx := range ns.Agents {
		if ns.Agents[idx].Name == name {
			ns.Agents = append(ns.Agents[:idx], ns.Agents[idx+1:]...)

			if err := updateFile(); err != nil {
				return err
			}
			return nil
		}
	}

	return util.NewNotFoundError(namespace + "/" + name)
}

// DeleteMicroservice export
func DeleteMicroservice(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}

	for idx := range ns.Microservices {
		if ns.Microservices[idx].Name == name {
			ns.Microservices = append(ns.Microservices[:idx], ns.Microservices[idx+1:]...)
			if err := updateFile(); err != nil {
				return err
			}
			return nil
		}
	}

	return util.NewNotFoundError(namespace + "/" + name)
}

func getNamespace(name string) (*Namespace, error) {
	for idx := range conf.Namespaces {
		if conf.Namespaces[idx].Name == name {
			return &conf.Namespaces[idx], nil
		}
	}
	return nil, util.NewNotFoundError(name)
}

func updateFile() (err error) {
	marshal, err := yaml.Marshal(&conf)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(configFilename, marshal, 0644)
	if err != nil {
		return
	}
	return
}
