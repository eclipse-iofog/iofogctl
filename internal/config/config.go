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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
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
	configV2             = "iofogctl/v2"
	configV1             = "iofogctl/v1"
	CurrentConfigVersion = configV2
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

	// Map old config to new config file system
	for _, ns := range oldConfig.Namespaces {
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
func Init(configFolderArg string) {
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

	// Check config file already exists
	if _, err := os.Stat(configFilename); os.IsNotExist(err) {
		err = os.MkdirAll(configFolder, 0755)
		util.Check(err)

		// Create default config file
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
	namespaceFilename := getNamespaceFile("default")
	if _, err := os.Stat(namespaceFilename); os.IsNotExist(err) {
		err = os.MkdirAll(namespaceDirectory, 0755)
		util.Check(err)

		// Create default namespace file
		if err = AddNamespace("default", util.NowUTC()); err != nil {
			util.Check(errors.New("Could not initialize default namespace configuration"))
		}
	}
}

func SetDefaultNamespace(name string) (err error) {
	if name == conf.DefaultNamespace {
		return
	}
	// Check exists
	for _, n := range GetNamespaces() {
		if n == name {
			conf.DefaultNamespace = name
			return
		}
	}
	return util.NewNotFoundError(name)
}

// GetNamespaces returns all namespaces in config
func GetNamespaces() (namespaces []string) {
	files, err := ioutil.ReadDir(namespaceDirectory)
	util.Check(err)

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	for _, file := range files {
		name := util.Before(file.Name(), ".yaml")
		namespaces = append(namespaces, name)
	}
	return
}

func GetDefaultNamespaceName() string {
	return conf.DefaultNamespace
}

func getNamespace(name string) (*Namespace, error) {
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

// GetNamespace returns the namespace
func GetNamespace(namespace string) (Namespace, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return Namespace{}, err
	}
	return *ns, nil
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
	// Check collision
	for _, n := range GetNamespaces() {
		if n == name {
			return util.NewConflictError(name)
		}
	}

	newNamespace := Namespace{
		Name:    name,
		Created: created,
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

// UpdateControlPlane overwrites Control Plane in the namespace
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
	if name == "default" {
		return util.NewInputError("Cannot delete namespace named \"default\"")
	}

	// Reset default namespace if required
	if name == conf.DefaultNamespace {
		err1 := SetDefaultNamespace("default")
		err2 := FlushConfig()
		if err1 != nil || err2 != nil {
			return errors.New("Failed to delete namespace " + name + " which is configured as default")
		}
	}

	filename := getNamespaceFile(name)
	if err := os.Remove(filename); err != nil {
		return util.NewNotFoundError("Could not delete namespace file " + filename)
	}

	delete(namespaces, name)

	return nil
}

// RenameNamespace renames a namespace
func RenameNamespace(name, newName string) error {

	if name == conf.DefaultNamespace {
		util.PrintError("Cannot rename default namespaces, please choose a different namespace to rename")
		return util.NewInputError("Cannot find valid namespace with name: " + name)
	}
	ns, err := getNamespace(name)
	if err != nil {
		util.PrintError("Could not find namespace " + name)
		return err
	}
	ns.Name = newName
	err = os.Rename(getNamespaceFile(name), getNamespaceFile(newName))
	if err != nil {
		return err
	}
	err = FlushConfig()
	if err != nil {
		return err
	}
	return Flush()
}

func ClearNamespace(namespace string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	mux.Lock()
	defer mux.Unlock()
	ns.ControlPlane = ControlPlane{}
	ns.Agents = []Agent{}
	return FlushConfig()
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

func updateNamespaceToV2(header iofogctlNamespace) (iofogctlNamespace, error) {
	type v1SpecContent struct {
		Name         string        `yaml:"name,omitempty"`
		ControlPlane ControlPlane  `yaml:"controlPlane,omitempty"`
		Agents       []Agent       `yaml:"agents,omitempty"`
		Created      string        `yaml:"created,omitempty"`
		Connectors   []interface{} `yaml:"connectors,omitempty"`
	}
	header.APIVersion = configV2
	bytes, err := yaml.Marshal(header.Spec)
	v1Spec := v1SpecContent{}
	if err != nil {
		return header, err
	}
	if err = yaml.UnmarshalStrict(bytes, &v1Spec); err != nil {
		return header, err
	}

	v2Spec := Namespace{
		Name:         v1Spec.Name,
		ControlPlane: v1Spec.ControlPlane,
		Agents:       v1Spec.Agents,
		Created:      v1Spec.Created,
	}

	header.Spec = v2Spec

	return header, nil
}

func updateConfigToV2(header iofogctlConfig) (iofogctlConfig, error) {
	header.APIVersion = configV2
	return header, nil
}

func getConfigFromHeader(header iofogctlConfig) (c configuration, err error) {
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
	case configV1:
		{
			headerV2, err := updateConfigToV2(header)
			if err != nil {
				return c, err
			}
			return getConfigFromHeader(headerV2)
		}
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
	if c.DetachedResources.Agents == nil {
		c.DetachedResources.Agents = make(map[string]Agent)
	}
	return
}

func getNamespaceFromHeader(header iofogctlNamespace) (n Namespace, err error) {
	switch header.APIVersion {
	case CurrentConfigVersion:
		{
			// All good
			break
		}
	case configV1:
		{
			headerV2, err := updateNamespaceToV2(header)
			if err != nil {
				return n, err
			}
			return getNamespaceFromHeader(headerV2)
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

func GetDetachedAgent(name string) (Agent, error) {
	if agent, found := conf.DetachedResources.Agents[name]; found {
		return agent, nil
	}

	return Agent{}, util.NewNotFoundError(name)
}

func AttachAgent(namespace, name, UUID string) error {
	agent, err := GetDetachedAgent(name)
	if err != nil {
		return err
	}
	delete(conf.DetachedResources.Agents, name)
	if err = FlushConfig(); err != nil {
		return err
	}
	agent.UUID = UUID
	return UpdateAgent(namespace, agent)
}

func DetachAgent(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	for idx := range ns.Agents {
		if ns.Agents[idx].Name == name {
			mux.Lock()
			detachedAgent := ns.Agents[idx]
			detachedAgent.UUID = ""
			ns.Agents = append(ns.Agents[:idx], ns.Agents[idx+1:]...)
			conf.DetachedResources.Agents[detachedAgent.Name] = detachedAgent
			mux.Unlock()
			return FlushConfig()
		}
	}

	return util.NewNotFoundError(ns.Name + "/" + name)
}

func GetDetachedResources() DetachedResources {
	return conf.DetachedResources
}

func RenameDetachedAgent(oldName, newName string) error {
	agent, err := GetDetachedAgent(oldName)
	if err != nil {
		return err
	}
	agent.Name = newName
	delete(conf.DetachedResources.Agents, oldName)
	conf.DetachedResources.Agents[newName] = agent
	return FlushConfig()
}

func DeleteDetachedResources() error {
	conf.DetachedResources = DetachedResources{
		Agents: make(map[string]Agent),
	}

	return FlushConfig()
}

func DeleteDetachedAgent(name string) error {
	if _, err := GetDetachedAgent(name); err != nil {
		return err
	}
	delete(conf.DetachedResources.Agents, name)
	return FlushConfig()
}

func UpdateDetachedAgent(agent Agent) error {
	if _, err := GetDetachedAgent(agent.Name); err != nil {
		return err
	}
	conf.DetachedResources.Agents[agent.Name] = agent
	return FlushConfig()
}
