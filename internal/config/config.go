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

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	homedir "github.com/mitchellh/go-homedir"
	yaml "gopkg.in/yaml.v2"
)

// struct that file is unmarshalled into
var conf configuration

// Name of file
var configFilename string

const defaultDirname = ".iofog/"
const defaultFilename = "config.yaml"

// DefaultConfigPath is used if user does not specify a config file path
const DefaultConfigPath = "~/" + defaultDirname + defaultFilename

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
		defaultData := []byte(`namespaces:
- name: default
  controllers: []
  agents: []
  microservices: []
  created: ` + util.NowUTC())
		err := ioutil.WriteFile(configFilename, defaultData, 0644)
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
			return ns.Controllers, nil
		}
	}
	return nil, util.NewNotFoundError(namespace)
}

// GetMicroservices returns all microservices within a namespace
func GetMicroservices(namespace string) ([]Microservice, error) {
	for _, ns := range conf.Namespaces {
		if ns.Name == namespace {
			return ns.Microservices, nil
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

// GetController returns a single controller within a namespace
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

// GetMicroservice returns a single microservice within a namespace
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
	conf.Namespaces = append(conf.Namespaces, newNamespace)
	return nil
}

// Overwrites or creates new controller to the namespace
func UpdateController(namespace string, controller Controller) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	// Update existing controller if exists
	for idx := range ns.Controllers {
		if ns.Controllers[idx].Name == controller.Name {
			ns.Controllers[idx] = controller
			return nil
		}
	}
	// Add new controller
	AddController(namespace, controller)

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
			ns.Agents[idx] = agent
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
	ns.Controllers = append(ns.Controllers, controller)

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
	ns.Agents = append(ns.Agents, agent)

	return nil
}

// AddMicroservice adds a new microservice to the namespace
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

	return nil
}

// DeleteNamespace removes a namespace including all the resources within it
func DeleteNamespace(name string) error {
	for idx := range conf.Namespaces {
		if conf.Namespaces[idx].Name == name {
			conf.Namespaces = append(conf.Namespaces[:idx], conf.Namespaces[idx+1:]...)
			return nil
		}
	}

	return nil
}

// DeleteController deletes a controller from a namespace
func DeleteController(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}

	for idx := range ns.Controllers {
		if ns.Controllers[idx].Name == name {
			ns.Controllers = append(ns.Controllers[:idx], ns.Controllers[idx+1:]...)
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
			ns.Agents = append(ns.Agents[:idx], ns.Agents[idx+1:]...)
			return nil
		}
	}

	return util.NewNotFoundError(namespace + "/" + name)
}

// DeleteMicroservice deletes a microservice from a namespace
func DeleteMicroservice(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}

	for idx := range ns.Microservices {
		if ns.Microservices[idx].Name == name {
			ns.Microservices = append(ns.Microservices[:idx], ns.Microservices[idx+1:]...)
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
