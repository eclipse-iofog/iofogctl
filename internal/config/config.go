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
	conf               configuration
	configFolder       string // config directory
	configFilename     string // config file name
	namespaceDirectory string // Path of namespace directory
	namespaces         map[string]*Namespace
	// TODO: Replace sync.Mutex with chan impl (if its worth the code)
	mux = &sync.Mutex{}
)

const (
	defaultDirname       = ".iofog/"
	namespaceDirname     = "namespaces/"
	defaultFilename      = "config.yaml"
	CurrentConfigVersion = "iofogctl/v1"
)

func updateConfigToK8sStyle() error {
	// Previous config structure
	type OldConfig struct {
		Namespaces []Namespace `yaml:"namespaces"`
	}

	// Get config files
	configFileName := path.Join(configFolder, "config.yaml")
	configSaveFileName := path.Join(configFolder, "config.yaml.save")

	// Create namespaces folder
	namespaceDirectory := path.Join(configFolder, "namespaces")
	err := os.MkdirAll(namespaceDirectory, 0755)
	util.Check(err)

	// Read previous config
	r, err := ioutil.ReadFile(configFileName)
	util.Check(err)

	oldConfig := OldConfig{}
	newConfig := configuration{DefaultNamespace: "default"}
	configHeader := iofogctlConfig{}
	err = yaml.UnmarshalStrict(r, &oldConfig)
	if err != nil {
		if err2 := yaml.UnmarshalStrict(r, &configHeader); err2 != nil {
			util.Check(err)
		}
		return nil
	}

	// Map old config to new confi file system
	for _, ns := range oldConfig.Namespaces {
		// Add namespace to list
		newConfig.Namespaces = append(newConfig.Namespaces, ns.Name)

		// Write namespace config file
		bytes, err := getNamespaceYAMLFile(&ns)
		util.Check(err)
		configFile := getNamespaceFile(ns.Name)
		err = ioutil.WriteFile(configFile, bytes, 0644)
		util.Check(err)
	}

	// Write old config save file
	err = ioutil.WriteFile(configSaveFileName, r, 0644)
	util.Check(err)

	// Write new config file
	bytes, err := getConfigYAMLFile(newConfig)
	util.Check(err)
	err = ioutil.WriteFile(configFileName, bytes, 0644)
	util.Check(err)

	util.PrintInfo(fmt.Sprintf("Your config file has successfully been updated, the previous config file has been saved under %s", configSaveFileName))
	return nil
}

// Init initializes config, namespace and unmarshalls the files
func Init(namespace, configFolderArg string) {
	namespaces = make(map[string]*Namespace)

	var err error
	configFolder, err = util.FormatPath(configFolderArg)
	util.Check(err)

	if configFolder == "" {
		// Find home directory.
		home, err := homedir.Dir()
		util.Check(err)
		configFolder = path.Join(home, defaultDirname)
	} else {
		dirInfo, err := os.Stat(configFolder)
		util.Check(err)
		if dirInfo.IsDir() == false {
			util.Check(util.NewInputError(fmt.Sprintf("The config folder %s is not a valid directory", configFolder)))
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
	confHeader := iofogctlConfig{}
	err = util.UnmarshalYAML(configFilename, &confHeader)
	// Warn user about possible update
	if err != nil {
		if err = updateConfigToK8sStyle(); err != nil {
			util.Check(util.NewInternalError(fmt.Sprintf("Failed to update iofogctl configuration. Error: %v", err)))
		}
		err = util.UnmarshalYAML(configFilename, &confHeader)
		util.Check(err)
	}

	conf, err = getConfigFromHeader(confHeader)
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
		namespaceHeader := iofogctlNamespace{}
		if err := util.UnmarshalYAML(getNamespaceFile(name), &namespaceHeader); err != nil {
			if os.IsNotExist(err) {
				return nil, util.NewNotFoundError(name)
			}
			return nil, err
		}
		ns, err := getNamespaceFromHeader(namespaceHeader)
		if err != nil {
			return nil, err
		}
		namespaces[name] = &ns
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
	ns, _ := getNamespace(conf.CurrentNamespace)
	return *ns
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
	marshal, err := getNamespaceYAMLFile(&newNamespace)
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
	if name == "" {
		name = conf.CurrentNamespace
	}
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

	return util.NewNotFoundError("Could not find namespace " + name)
}

// RenameNamespace renames a namespace
func RenameNamespace(name, newName string) error {
	if name == "" {
		name = conf.CurrentNamespace
	}
	if name == conf.DefaultNamespace {
		util.PrintError("Cannot rename default namespaces, please choose a different namespace to rename")
		return util.NewInputError("Cannot find valid namespace with name: " + name)
	}
	for idx, ns := range conf.Namespaces {
		if ns == name {
			mux.Lock()
			defer mux.Unlock()
			// Rename namespace file
			conf.Namespaces[idx] = newName
			err := os.Rename(getNamespaceFile(name), getNamespaceFile(newName))
			if err != nil {
				return err
			}
			err = FlushConfig()
			return err
		}
	}

	return util.NewNotFoundError("Could not find namespace " + name)
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
	marshal, err := getConfigYAMLFile(conf)
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

func getConfigFromHeader(header iofogctlConfig) (c configuration, err error) {
	if header.APIVersion != CurrentConfigVersion {
		return c, util.NewInputError("Invalid iofogctl config version")
	}
	switch header.APIVersion {
	case CurrentConfigVersion:
		{
			// All good
			break
		}
	// Example for further maintenance
	// case PreviousConfigVersion {
	// 	updateFromPreviousVersion()
	// 	break
	// }
	default:
		return c, util.NewInputError("Invalid iofogctl config version")
	}
	bytes, err := yaml.Marshal(header.Spec)
	if err != nil {
		return
	}
	if err = yaml.UnmarshalStrict(bytes, &c); err != nil {
		return
	}
	return
}

func getNamespaceFromHeader(header iofogctlNamespace) (n Namespace, err error) {
	if header.Kind != IofogctlNamespaceKind {
		return n, util.NewInputError("Invalid namespace kind")
	}
	switch header.APIVersion {
	case CurrentConfigVersion:
		{
			// All good
			break
		}
	// Example for further maintenance
	// case PreviousConfigVersion {
	// 	updateFromPreviousVersion()
	// 	break
	// }
	default:
		return n, util.NewInputError("Invalid iofogctl config version")
	}
	bytes, err := yaml.Marshal(header.Spec)
	if err != nil {
		return
	}
	if err = yaml.UnmarshalStrict(bytes, &n); err != nil {
		return
	}
	return
}

func getConfigYAMLFile(conf configuration) ([]byte, error) {
	confHeader := iofogctlConfig{
		Header: Header{
			Kind:       IofogctlConfigKind,
			APIVersion: CurrentConfigVersion,
			Spec:       conf,
		},
	}

	return yaml.Marshal(confHeader)
}

func getNamespaceYAMLFile(ns *Namespace) ([]byte, error) {
	namespaceHeader := iofogctlNamespace{
		Header{
			Kind:       IofogctlNamespaceKind,
			APIVersion: CurrentConfigVersion,
			Metadata: HeaderMetadata{
				Name: ns.Name,
			},
			Spec: ns,
		},
	}
	return yaml.Marshal(namespaceHeader)
}

// Flush will write over the namespace file based on the runtime data
func Flush() (err error) {
	for _, ns := range namespaces {
		// Marshal the runtime data
		marshal, err := getNamespaceYAMLFile(ns)
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
