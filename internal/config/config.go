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
	"sync"

	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
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
	detachedNamespace    = "_detached"
)

type v1NamespaceSpecContent struct {
	Name         string        `yaml:"name,omitempty"`
	ControlPlane ControlPlane  `yaml:"controlPlane,omitempty"`
	Agents       []Agent       `yaml:"agents,omitempty"`
	Created      string        `yaml:"created,omitempty"`
	Connectors   []interface{} `yaml:"connectors,omitempty"`
}

func updateConfigToK8sStyle() error {
	// Previous config structure
	type OldConfig struct {
		Namespaces []v1NamespaceSpecContent `yaml:"namespaces"`
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
		namespaceHeader := iofogctlNamespace{
			Header{
				Kind:       IofogctlNamespaceKind,
				APIVersion: configV1,
				Metadata: HeaderMetadata{
					Name: ns.Name,
				},
				Spec: ns,
			},
		}
		bytes, err := yaml.Marshal(namespaceHeader)
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
		err = flushConfig()
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
	initNamespaces := []string{"default", detachedNamespace}
	flush := false
	for _, initNamespace := range initNamespaces {
		nsFile := getNamespaceFile(initNamespace)
		if _, err := os.Stat(nsFile); os.IsNotExist(err) {
			flush = true
			err = os.MkdirAll(namespaceDirectory, 0755)
			util.Check(err)

			// Create default namespace file
			if err = AddNamespace(initNamespace, util.NowUTC()); err != nil {
				util.Check(errors.New("Could not initialize " + initNamespace + " configuration"))
			}
		}
	}
	if flush {
		err = flushNamespaces()
		util.Check(err)
	}
}

// getNamespaceFile helper function that returns the full path to a namespace file
func getNamespaceFile(name string) string {
	return path.Join(namespaceDirectory, name+".yaml")
}

func updateNamespaceToV2(header iofogctlNamespace) (iofogctlNamespace, error) {
	type v1NamespaceSpecContent struct {
		Name         string        `yaml:"name,omitempty"`
		ControlPlane ControlPlane  `yaml:"controlPlane,omitempty"`
		Agents       []Agent       `yaml:"agents,omitempty"`
		Created      string        `yaml:"created,omitempty"`
		Connectors   []interface{} `yaml:"connectors,omitempty"`
	}
	header.APIVersion = configV2
	bytes, err := yaml.Marshal(header.Spec)
	v1Spec := v1NamespaceSpecContent{}
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

func flushNamespaces() error {
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
	return nil
}

func flushConfig() error {
	// Marshal the runtime data
	marshal, err := getConfigYAMLFile(conf)
	if err != nil {
		return nil
	}
	// Overwrite the file
	err = ioutil.WriteFile(configFilename, marshal, 0644)
	if err != nil {
		return nil
	}
	return nil
}

// Flush will write namespace and configuration files to disk
func Flush() (err error) {
	// Flush namespace files
	if err = flushNamespaces(); err != nil {
		return
	}
	// Flush configuration e.g. default namespace
	return flushConfig()
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
