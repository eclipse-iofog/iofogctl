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
	namespaceIndex map[string]int
	controllerIndex map[string][2]int
	agentIndex map[string][2]int
}

// NewManager export
func NewManager(filename string) *Manager {
	manager := new(Manager)
	manager.namespaceIndex = make(map[string]int)
	manager.controllerIndex = make(map[string][2]int)
	manager.agentIndex = make(map[string][2]int)
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
	for nsIdx, ns := range manager.configuration.Namespaces {
		manager.namespaceIndex[ns.Name] = nsIdx
		for ctrlIdx, ctrl := range ns.Controllers {
			manager.controllerIndex[ns.Name + ctrl.Name] = [2]int{nsIdx, ctrlIdx}
		}
		for agntIdx, agnt := range ns.Agents {
			manager.agentIndex[ns.Name + agnt.Name] = [2]int{nsIdx, agntIdx}
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
	idx, exists := manager.namespaceIndex[namespace]
	if !exists {
		err = util.NewNotFound(namespace)
		return
	}

	// Copy the agents
	copy(agents, manager.configuration.Namespaces[idx].Agents)

	return
}

// GetControllers export
func (manager *Manager) GetControllers(namespace string) (controllers []Controller, err error) {
	// Get the requested namespace
	idx, exists := manager.namespaceIndex[namespace]
	if !exists {
		err = util.NewNotFound(namespace)
		return
	}

	// Copy the controllers
	copy(controllers, manager.configuration.Namespaces[idx].Controllers)

	return
}

// GetNamespace export
func (manager *Manager) GetNamespace(name string) (namespace Namespace, err error){
	idx, exists := manager.namespaceIndex[name]
	if !exists {
		err = util.NewNotFound(name)
		return 
	}
	namespace.Name = manager.configuration.Namespaces[idx].Name
	return
}

// GetController export
func (manager *Manager) GetController(namespace, name string) (controller Controller, err error){
	idxs, exists := manager.controllerIndex[namespace + name]
	if !exists {
		err = util.NewNotFound(namespace + "/" + name)
		return
	}

	controller = manager.configuration.Namespaces[idxs[0]].Controllers[idxs[1]]
	return
}

// GetAgent export
func (manager *Manager) GetAgent(namespace, name string) (agent Agent, err error){
	idxs, exists := manager.agentIndex[namespace + name]
	if !exists {
		err = util.NewNotFound(namespace + "/" + name)
		return
	}

	agent = manager.configuration.Namespaces[idxs[0]].Agents[idxs[1]]
	return
}