/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	homedir "github.com/mitchellh/go-homedir"
	yaml "gopkg.in/yaml.v2"
)

var (
	conf               configuration // struct that file is unmarshalled into
	configFolder       string        // config directory
	configFilename     string        // config file name
	namespaceDirectory string        // Path of namespace directory
	namespaces         map[string]*Namespace
	// TODO: Replace sync.Mutex with chan impl (if its worth the code)
	mux = &sync.Mutex{}
)

const (
	defaultDirname   = ".iofog/"
	namespaceDirname = "namespaces/"
	defaultFilename  = "config.yaml"
	// DefaultConfigPath is used if user does not specify a config file path
	DefaultConfigPath = "~/" + defaultDirname
)

// Init initializes config, namespace and unmarshalls the files
func Init(dirPath, namespace string) {
	namespaces = make(map[string]*Namespace)

	// Format file path
	dirPath, err := util.FormatPath(dirPath)
	util.Check(err)

	if dirPath == "" {
		// Find home directory.
		home, err := homedir.Dir()
		util.Check(err)
		configFolder = path.Join(home, defaultDirname)
	} else {
		fi, err := os.Stat(dirPath)
		util.Check(err)
		if fi.IsDir() {
			// it's a directory
			configFolder = dirPath
		} else {
			// it's not a directory
			util.Check(util.NewInputError(fmt.Sprintf("The specified config folder [%s] is not a valid folder", dirPath)))
		}
	}

	// Set default filename if necessary
	filename := path.Join(configFolder, defaultFilename)
	configFilename = filename
	namespaceDirectory = path.Join(configFolder, namespaceDirname)

	// Check file exists
	if _, err := os.Stat(configFilename); os.IsNotExist(err) {
		err = os.MkdirAll(configFolder, 0755)
		util.Check(err)

		// Create default file
		conf.DefaultNamespace = "default"
		err = FlushConfig()
		util.Check(err)
	}

	// Unmarshall the config file
	err = util.UnmarshalYAML(configFilename, &conf)
	util.Check(err)

	// Check namespace dir exists
	if namespace == "" {
		namespace = conf.DefaultNamespace
	}
	conf.CurrentNamespace = namespace
	namespaceFilename := getNamespaceFile(namespace)
	if _, err := os.Stat(namespaceFilename); os.IsNotExist(err) {
		err = os.MkdirAll(namespaceDirectory, 0755)
		util.Check(err)

		// Create default file
		if namespace == "default" {
			err = AddNamespace(namespace, util.NowUTC())
			util.Check(err)
		}
	}
}

func SetDefaultNamespace(name string) (err error) {
	if name == conf.DefaultNamespace {
		return
	}
	// Check exists
	for _, n := range conf.Namespaces {
		if n == name {
			conf.CurrentNamespace = name
			conf.DefaultNamespace = name
			// Unmarshall the namespace file
			namespaces[name] = &Namespace{}
			err = util.UnmarshalYAML(getNamespaceFile(name), namespaces[name])
			return
		}
	}
	return util.NewNotFoundError(name)
}

// GetNamespaces returns all namespaces in config
func GetNamespaces() (namespaces []string) {
	return conf.Namespaces
}

func getNamespace(name string) (*Namespace, error) {
	if name == "" {
		name = conf.CurrentNamespace
	}
	namespace, ok := namespaces[name]
	if !ok {

		namespaces[name] = &Namespace{}
		if err := util.UnmarshalYAML(getNamespaceFile(name), namespaces[name]); err != nil {
			delete(namespaces, name)
			return nil, err
		}
		return namespaces[name], nil
	}
	return namespace, nil
}

// GetAgents returns all agents within the namespace
func GetAgents(namespace string) ([]Agent, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return ns.Agents, nil
}

// GetControllers returns all controllers within the namespace
func GetControllers(namespace string) ([]Controller, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return ns.ControlPlane.Controllers, nil
}

// GetConnectors returns all controllers within the namespace
func GetConnectors(namespace string) ([]Connector, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return ns.Connectors, nil
}

// GetNamespace returns the namespace
func GetNamespace(namespace string) (Namespace, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return Namespace{}, err
	}
	return *ns, nil
}

// GetCurrentNamespace return the current namespace
func GetCurrentNamespace() Namespace {
	return *namespaces[conf.CurrentNamespace]
}

// GetControlPlane returns a control plane within a namespace
func GetControlPlane(namespace string) (ControlPlane, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return ControlPlane{}, err
	}
	return ns.ControlPlane, nil
}

// GetController returns a single controller within the current
func GetController(namespace, name string) (controller Controller, err error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return
	}
	for _, ctrl := range ns.ControlPlane.Controllers {
		if ctrl.Name == name {
			controller = ctrl
			return
		}
	}

	err = util.NewNotFoundError(namespace + "/" + name)
	return
}

// GetConnector returns a single connector within a namespace
func GetConnector(namespace, name string) (connector Connector, err error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return
	}
	for _, cnct := range ns.Connectors {
		if cnct.Name == name {
			connector = cnct
			return
		}
	}

	err = util.NewNotFoundError(namespace + "/" + name)
	return
}

// GetAgent returns a single agent within a namespace
func GetAgent(namespace, name string) (agent Agent, err error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return
	}
	for _, ag := range ns.Agents {
		if ag.Name == name {
			agent = ag
			return
		}
	}

	err = util.NewNotFoundError(namespace + "/" + name)
	return
}

// AddNamespace adds a new namespace to the config
func AddNamespace(name, created string) error {
	if name == "" {
		name = conf.CurrentNamespace
	}
	// Check collision
	for _, n := range conf.Namespaces {
		if n == name {
			return util.NewConflictError(name)
		}
	}

	newNamespace := Namespace{
		Name:    name,
		Created: created,
	}
	mux.Lock()
	conf.Namespaces = append(conf.Namespaces, name)
	err := FlushConfig()
	mux.Unlock()
	if err != nil {
		return err
	}

	// Write namespace file
	// Marshal the runtime data
	marshal, err := yaml.Marshal(&newNamespace)
	if err != nil {
		return err
	}
	// Overwrite the file
	err = ioutil.WriteFile(getNamespaceFile(name), marshal, 0644)
	if err != nil {
		return err
	}
	namespaces[name] = &newNamespace
	return nil
}

// UpdateConnector overwrites Control Plane in the namespace
func UpdateControlPlane(namespace string, controlPlane ControlPlane) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	mux.Lock()
	ns.ControlPlane = controlPlane
	mux.Unlock()
	return nil
}

// Overwrites or creates new controller to the namespace
func UpdateController(namespace string, controller Controller) error {
	// Update existing controller if exists
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	for idx := range ns.ControlPlane.Controllers {
		if ns.ControlPlane.Controllers[idx].Name == controller.Name {
			mux.Lock()
			ns.ControlPlane.Controllers[idx] = controller
			mux.Unlock()
			return nil
		}
	}
	// Add new controller
	return AddController(namespace, controller)
}

// Overwrites or creates new connector to the namespace
func UpdateConnector(namespace string, connector Connector) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	// Update existing connector if exists
	for idx := range ns.Connectors {
		if ns.Connectors[idx].Name == connector.Name {
			mux.Lock()
			ns.Connectors[idx] = connector
			mux.Unlock()
			return nil
		}
	}
	// Add new connector
	return AddConnector(namespace, connector)
}

// Overwrites or creates new agent to the namespace
func UpdateAgent(namespace string, agent Agent) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	// Update existing agent if exists
	for idx := range ns.Agents {
		if ns.Agents[idx].Name == agent.Name {
			mux.Lock()
			ns.Agents[idx] = agent
			mux.Unlock()
			return nil
		}
	}
	// Add new agent
	return AddAgent(namespace, agent)
}

// AddController adds a new controller to the current namespace
func AddController(namespace string, controller Controller) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	_, err = GetController(namespace, controller.Name)
	if err == nil {
		return util.NewConflictError(namespace + "/" + controller.Name)
	}

	// Append the controller
	mux.Lock()
	ns.ControlPlane.Controllers = append(ns.ControlPlane.Controllers, controller)
	mux.Unlock()

	return nil
}

// AddConnector adds a new connector to the namespace
func AddConnector(namespace string, connector Connector) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	_, err = GetConnector(namespace, connector.Name)
	if err == nil {
		return util.NewConflictError(namespace + "/" + connector.Name)
	}

	// Append the connector
	mux.Lock()
	ns.Connectors = append(ns.Connectors, connector)
	mux.Unlock()

	return nil
}

// AddAgent adds a new agent to the namespace
func AddAgent(namespace string, agent Agent) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	_, err = GetAgent(namespace, agent.Name)
	if err == nil {
		return util.NewConflictError(namespace + "/" + agent.Name)
	}

	// Append the controller
	mux.Lock()
	ns.Agents = append(ns.Agents, agent)
	mux.Unlock()

	return nil
}

// getNamespaceFile helper function that returns the full path to a namespace file
func getNamespaceFile(name string) string {
	return path.Join(namespaceDirectory, name+".yaml")
}

// DeleteNamespace removes a namespace including all the resources within it
func DeleteNamespace(name string) error {
	for idx := range conf.Namespaces {
		if conf.Namespaces[idx] == name {
			mux.Lock()
			conf.Namespaces = append(conf.Namespaces[:idx], conf.Namespaces[idx+1:]...)
			delete(namespaces, name)
			// Remove namespace file
			err := os.Remove(getNamespaceFile(name))
			mux.Unlock()
			return err
		}
	}

	return nil
}

func DeleteControlPlane(namespace string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	mux.Lock()
	ns.ControlPlane = ControlPlane{}
	mux.Unlock()
	return nil
}

// DeleteController deletes a controller from a namespace
func DeleteController(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	for idx := range ns.ControlPlane.Controllers {
		if ns.ControlPlane.Controllers[idx].Name == name {
			mux.Lock()
			ns.ControlPlane.Controllers = append(ns.ControlPlane.Controllers[:idx], ns.ControlPlane.Controllers[idx+1:]...)
			mux.Unlock()
			return nil
		}
	}

	return util.NewNotFoundError(ns.Name + "/" + name)
}

// DeleteConnector deletes a connector from a namespace
func DeleteConnector(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	for idx := range ns.Connectors {
		if ns.Connectors[idx].Name == name {
			mux.Lock()
			ns.Connectors = append(ns.Connectors[:idx], ns.Connectors[idx+1:]...)
			mux.Unlock()
			return nil
		}
	}

	return util.NewNotFoundError(ns.Name + "/" + name)
}

// DeleteAgent deletes an agent from a namespace
func DeleteAgent(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	for idx := range ns.Agents {
		if ns.Agents[idx].Name == name {
			mux.Lock()
			ns.Agents = append(ns.Agents[:idx], ns.Agents[idx+1:]...)
			mux.Unlock()
			return nil
		}
	}

	return util.NewNotFoundError(ns.Name + "/" + name)
}

// FlushConfig will write over the config file based on the runtime data of all namespaces
func FlushConfig() (err error) {
	// Marshal the runtime data
	marshal, err := yaml.Marshal(&conf)
	if err != nil {
		return
	}
	// Overwrite the file
	err = ioutil.WriteFile(configFilename, marshal, 0644)
	if err != nil {
		return
	}
	return
}

// Flush will write over the namespace file based on the runtime data
func Flush() (err error) {
	for _, ns := range namespaces {
		// Marshal the runtime data
		marshal, err := yaml.Marshal(ns)
		if err != nil {
			return err
		}
		// Overwrite the file
		err = ioutil.WriteFile(getNamespaceFile(ns.Name), marshal, 0644)
		if err != nil {
			return err
		}
	}
	return
}

// NewRandomUser creates a new config user
func NewRandomUser() IofogUser {
	return IofogUser{
		Name:     "N" + util.RandomString(10, util.AlphaLower),
		Surname:  "S" + util.RandomString(10, util.AlphaLower),
		Email:    util.RandomString(5, util.AlphaLower) + "@domain.com",
		Password: util.RandomString(10, util.AlphaNum),
	}
}
