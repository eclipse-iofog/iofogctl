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
	"io/ioutil"
	"os"
	"sync"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	homedir "github.com/mitchellh/go-homedir"
	yaml "gopkg.in/yaml.v2"
)

var (
	conf           configuration // struct that file is unmarshalled into
	configFilename string        // Name of file
	// TODO: Replace sync.Mutex with chan impl (if its worth the code)
	mux = &sync.Mutex{}
)

const (
	defaultDirname  = ".iofog/"
	defaultFilename = "config.yaml"
	// DefaultConfigPath is used if user does not specify a config file path
	DefaultConfigPath = "~/" + defaultDirname + defaultFilename
)

// Init initializes config and unmarshalls the file
func Init(filename string) {
	// Format file path
	filename, err := util.FormatPath(filename)
	util.Check(err)

	// Set default filename if necessary
	var homeDirname string
	if filename == "" {
		// Find home directory.
		home, err := homedir.Dir()
		util.Check(err)
		homeDirname = home + "/" + defaultDirname
		filename = homeDirname + defaultFilename
	}
	configFilename = filename

	// Check file exists
	if _, err := os.Stat(configFilename); os.IsNotExist(err) {
		err = os.MkdirAll(homeDirname, 0755)
		util.Check(err)

		// Create default file
		err = AddNamespace("default", util.NowUTC())
		util.Check(err)
		err = Flush()
		util.Check(err)
	}

	// Unmarshall the file
	err = util.UnmarshalYAML(configFilename, &conf)
	util.Check(err)
}

// GetNamespaces returns all namespaces in config
func GetNamespaces() (namespaces []Namespace) {
	return conf.Namespaces
}

// GetAgents returns all agents within a namespace
func GetAgents(namespace string) ([]Agent, error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == namespace {
			return ns.Agents, nil
		}
	}
	return nil, util.NewNotFoundError(namespace)
}

// GetControllers returns all controllers within a namespace
func GetControllers(namespace string) ([]Controller, error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == namespace {
			return ns.ControlPlane.Controllers, nil
		}
	}
	return nil, util.NewNotFoundError(namespace)
}

// GetConnectors returns all controllers within a namespace
func GetConnectors(namespace string) ([]Connector, error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == namespace {
			return ns.Connectors, nil
		}
	}
	return nil, util.NewNotFoundError(namespace)
}

// GetNamespace returns a single namespace
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

// GetControlPlane returns a control plane within a namespace
func GetControlPlane(namespace string) (controlplane ControlPlane, err error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == namespace {
			controlplane = ns.ControlPlane
			return
		}
	}
	err = util.NewNotFoundError(namespace + " Control Plane")
	return
}

// GetController returns a single controller within a namespace
func GetController(namespace, name string) (controller Controller, err error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == namespace {
			for _, ctrl := range ns.ControlPlane.Controllers {
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

// GetConnector returns a single connector within a namespace
func GetConnector(namespace, name string) (connector Connector, err error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == namespace {
			for _, cnct := range ns.Connectors {
				if cnct.Name == name {
					connector = cnct
					return
				}
			}
		}
	}
	err = util.NewNotFoundError(namespace + "/" + name)
	return
}

// GetAgent returns a single agent within a namespace
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

// AddNamespace adds a new namespace to the config
func AddNamespace(name, created string) error {
	// Check collision
	_, err := GetNamespace(name)
	if err == nil {
		return util.NewConflictError(name)
	}

	newNamespace := Namespace{
		Name:    name,
		Created: created,
	}
	mux.Lock()
	conf.Namespaces = append(conf.Namespaces, newNamespace)
	mux.Unlock()
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
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	// Update existing controller if exists
	for idx := range ns.ControlPlane.Controllers {
		if ns.ControlPlane.Controllers[idx].Name == controller.Name {
			mux.Lock()
			ns.ControlPlane.Controllers[idx] = controller
			mux.Unlock()
			return nil
		}
	}
	// Add new controller
	AddController(namespace, controller)

	return nil
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
	AddConnector(namespace, connector)

	return nil
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
	AddAgent(namespace, agent)

	return nil
}

// AddController adds a new controller to the namespace
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
	mux.Lock()
	ns.ControlPlane.Controllers = append(ns.ControlPlane.Controllers, controller)
	mux.Unlock()

	return nil
}

// AddConnector adds a new connector to the namespace
func AddConnector(namespace string, connector Connector) error {
	_, err := GetConnector(namespace, connector.Name)
	if err == nil {
		return util.NewConflictError(namespace + "/" + connector.Name)
	}

	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}

	// Append the connector
	mux.Lock()
	ns.Connectors = append(ns.Connectors, connector)
	mux.Unlock()

	return nil
}

// AddAgent adds a new agent to the namespace
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
	mux.Lock()
	ns.Agents = append(ns.Agents, agent)
	mux.Unlock()

	return nil
}

// DeleteNamespace removes a namespace including all the resources within it
func DeleteNamespace(name string) error {
	for idx := range conf.Namespaces {
		if conf.Namespaces[idx].Name == name {
			mux.Lock()
			conf.Namespaces = append(conf.Namespaces[:idx], conf.Namespaces[idx+1:]...)
			mux.Unlock()
			return nil
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

	return util.NewNotFoundError(namespace + "/" + name)
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

	return util.NewNotFoundError(namespace + "/" + name)
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

	return util.NewNotFoundError(namespace + "/" + name)
}

// getNamespace is a helper function to find a namespace and reference it directly
func getNamespace(name string) (*Namespace, error) {
	for idx := range conf.Namespaces {
		if conf.Namespaces[idx].Name == name {
			return &conf.Namespaces[idx], nil
		}
	}
	return nil, util.NewNotFoundError(name)
}

// Flush will write over the config file based on the runtime data of all namespaces
func Flush() (err error) {
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

// NewRandomUser creates a new config user
func NewRandomUser() IofogUser {
	return IofogUser{
		Name:     "N" + util.RandomString(10, util.AlphaLower),
		Surname:  "S" + util.RandomString(10, util.AlphaLower),
		Email:    util.RandomString(5, util.AlphaLower) + "@domain.com",
		Password: util.RandomString(10, util.AlphaNum),
	}
}
