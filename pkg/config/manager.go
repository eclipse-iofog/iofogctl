package config

import (
	"io/ioutil"
	"github.com/eclipse-iofog/cli/pkg/util"
	"fmt"
	"github.com/spf13/viper"
	homedir "github.com/mitchellh/go-homedir"
	yaml "gopkg.in/yaml.v2"
)

var filename string

// SetFile sets the config filename from root command
func SetFile(file string){
	filename = file
}

// Manager export
type Manager struct {
	configuration configuration
	namespaceIndex map[string]int // For O(1) time lookups of namespaces
	controllerIndex map[string][2]int // For O(1) time lookups of controllers
	agentIndex map[string][2]int // For O(1) time lookups of agents
}

// NewManager export
func NewManager() *Manager {
	manager := new(Manager)
	manager.namespaceIndex = make(map[string]int)
	manager.controllerIndex = make(map[string][2]int)
	manager.agentIndex = make(map[string][2]int)

	// Read the file and unmarshall contents
	if filename != "" {
		// Use config file from the flag.
		viper.SetConfigFile(filename)
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
		err = util.NewNotFoundError(namespace)
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
		err = util.NewNotFoundError(namespace)
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
		err = util.NewNotFoundError(name)
		return 
	}
	namespace.Name = manager.configuration.Namespaces[idx].Name
	return
}

// GetController export
func (manager *Manager) GetController(namespace, name string) (controller Controller, err error){
	idxs, exists := manager.controllerIndex[namespace + name]
	if !exists {
		err = util.NewNotFoundError(namespace + "/" + name)
		return
	}

	controller = manager.configuration.Namespaces[idxs[0]].Controllers[idxs[1]]
	return
}

// GetAgent export
func (manager *Manager) GetAgent(namespace, name string) (agent Agent, err error){
	idxs, exists := manager.agentIndex[namespace + name]
	if !exists {
		err = util.NewNotFoundError(namespace + "/" + name)
		return
	}

	agent = manager.configuration.Namespaces[idxs[0]].Agents[idxs[1]]
	return
}

// DeleteAgent export
func (manager *Manager) DeleteAgent(namespace, name string) (err error) {
	// Check exists
	idxs, exists := manager.agentIndex[namespace + name]
	if !exists {
		err = util.NewNotFoundError(namespace + "/" + name)
		return
	}
	// Perform deletion
	nsIdx := idxs[0]
	ns := &manager.configuration.Namespaces[nsIdx]
	delIdx := idxs[1]
	ns.Agents = append(ns.Agents[:delIdx], ns.Agents[delIdx+1:]...)
	for _, a := range ns.Agents {
		print(a.Name)
	}
	
	// Delete entry from index
	delete(manager.agentIndex, namespace + name)
	// Update index entries for elements after deleted element in the array
	for idx, agent := range ns.Agents[delIdx:] {
		manager.agentIndex[namespace + agent.Name] = [2]int{nsIdx, idx}
	}

	// TODO: (Serge) Replace panic with logic to reset configuration datastructure
	// Write to file
	marshal, err := yaml.Marshal(&manager.configuration)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(filename, marshal, 0644)
	if err != nil {
		panic(err)
	}

	return
}