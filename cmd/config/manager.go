package config

import (
	"github.com/eclipse-iofog/cli/cmd/util"
	"fmt"
	"github.com/spf13/viper"
	homedir "github.com/mitchellh/go-homedir"
)

// Manager export
type Manager struct {
	configuration configuration
	filename string
	namespaceIndex map[string]*namespace
	controllerIndex map[string]*Controller
	agentIndex map[string]*Agent
}

// NewManager export
func NewManager(filename string) *Manager {
	manager := new(Manager)
	manager.namespaceIndex = make(map[string]*namespace)
	manager.controllerIndex = make(map[string]*Controller)
	manager.agentIndex = make(map[string]*Agent)
	manager.filename = filename

	// Read the file and unmarshall contents
	if manager.filename != "" {
		// Use config file from the flag.
		viper.SetConfigFile(manager.filename)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		util.Check(err)

		// Search config in home directory with name ".cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	util.Check(err)
	fmt.Println("Using config file:", viper.ConfigFileUsed())

	// Unmarshall the configuration
	err = viper.Unmarshal(&manager.configuration)
	util.Check(err)

	// Update Indexes
	for _, ns := range manager.configuration.Namespaces {
		manager.namespaceIndex[ns.Name] = &ns
		for _, ctrl := range ns.Controllers {
			manager.controllerIndex[ns.Name + ctrl.Name] = &ctrl
		}
		for _, agnt := range ns.Agents {
			manager.agentIndex[ns.Name + agnt.Name] = &agnt
		}
	}

	return manager
}

// GetNamespaces export
func (manager *Manager) GetNamespaces() (namespaces []Namespace) {
	for _, ns := range manager.configuration.Namespaces {
		newNamespace := Namespace{Name: ns.Name}
		namespaces = append(namespaces, newNamespace)
	}
	return 
}

// GetAgents export
func (manager *Manager) GetAgents(namespace string) (agents []Agent, err error) {
	// Get the requested namespace
	ns, exists := manager.namespaceIndex[namespace]
	if !exists {
		err = util.NewNotFound(namespace)
		return
	}

	// Copy the agents
	copy(agents, ns.Agents)

	return
}

// GetControllers export
func (manager *Manager) GetControllers(namespace string) (controllers []Controller, err error) {
	// Get the requested namespace
	ns, exists := manager.namespaceIndex[namespace]
	if !exists {
		err = util.NewNotFound(namespace)
		return
	}

	// Copy the controllers
	copy(controllers, ns.Controllers)

	return
}

// GetNamespace export
func (manager *Manager) GetNamespace(name string) (namespace Namespace, err error){
	ns, exists := manager.namespaceIndex[name]
	if !exists {
		err = util.NewNotFound(name)
		return 
	}
	namespace.Name = ns.Name
	return
}

// GetController export
func (manager *Manager) GetController(namespace, name string) (controller Controller, err error){
	ctrl, exists := manager.controllerIndex[namespace + name]
	if !exists {
		err = util.NewNotFound(namespace + "/" + name)
		return
	}

	controller = *ctrl
	return
}

// GetAgent export
func (manager *Manager) GetAgent(namespace, name string) (agent Agent, err error){
	agnt, exists := manager.agentIndex[namespace + name]
	if !exists {
		err = util.NewNotFound(namespace + "/" + name)
		return
	}

	agent = *agnt
	return
}