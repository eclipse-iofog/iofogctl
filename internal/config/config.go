/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
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

	configv1 "github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	homedir "github.com/mitchellh/go-homedir"
	yaml "gopkg.in/yaml.v2"
)

var (
	conf               configuration
	configFolder       string // config directory
	configFilename     string // config file name
	namespaceDirectory string // Path of namespace directory
	namespaces         map[string]*rsc.Namespace
	// TODO: Replace sync.Mutex with chan impl (if its worth the code)
)

const (
	apiVersionGroup      = "iofog.org"
	latestVersion        = "v2"
	LatestAPIVersion     = apiVersionGroup + "/" + latestVersion
	defaultDirname       = ".iofog/" + latestVersion
	namespaceDirname     = "namespaces/"
	defaultFilename      = "config.yaml"
	configV2             = "iofogctl/v2"
	configV1             = "iofogctl/v1"
	CurrentConfigVersion = configV2
	detachedNamespace    = "_detached"
)

// Init initializes config, namespace and unmarshalls the files
func Init(configFolderArg string) {
	namespaces = make(map[string]*rsc.Namespace)

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
		err = flushShared()
		util.Check(err)
	}

	// Unmarshall the config file
	confHeader := iofogctlConfig{}
	err = util.UnmarshalYAML(configFilename, &confHeader)
	util.Check(err)

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
	if err = yaml.Unmarshal(bytes, &c); err != nil {
		return
	}
	return
}

func getNamespaceFromHeader(header iofogctlNamespace) (n rsc.Namespace, err error) {
	switch header.APIVersion {
	case CurrentConfigVersion:
		{
			// All good
			break
		}
	case configV1:
		{
			err = util.NewError("Namespace file is out of date.")
			return
		}
	default:
		return n, util.NewInputError("Invalid iofogctl config version")
	}
	bytes, err := yaml.Marshal(header.Spec)
	if err != nil {
		return
	}
	if err = yaml.Unmarshal(bytes, &n); err != nil {
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

func getNamespaceYAMLFile(ns *rsc.Namespace) ([]byte, error) {
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

func flushShared() error {
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

// Flush will write namespace files to disk
func Flush() error {
	return flushNamespaces()
}

func FlushShared() error {
	return flushShared()
}

type v1NamespaceSpecContent struct {
	Name         string                `yaml:"name,omitempty"`
	ControlPlane configv1.ControlPlane `yaml:"controlPlane,omitempty"`
	Agents       []configv1.Agent      `yaml:"agents,omitempty"`
	Created      string                `yaml:"created,omitempty"`
	Connectors   []configv1.Connector  `yaml:"connectors,omitempty"`
}

func ValidateHeader(header Header) error {
	if header.APIVersion != LatestAPIVersion {
		return util.NewInputError(fmt.Sprintf("Unsupported YAML API version %s.\nPlease use version %s. See iofog.org for specification details.", header.APIVersion, LatestAPIVersion))
	}
	return nil
}
