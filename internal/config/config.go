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

	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	homedir "github.com/mitchellh/go-homedir"
	yaml "gopkg.in/yaml.v2"
)

var pkg struct {
	conf                 configuration
	configFolder         string // config directory
	configFilename       string // config file name
	namespaceDirectory   string // Path of namespace directory
	namespaces           map[string]*rsc.Namespace
	apiVersionGroup      string
	latestVersion        string
	latestAPIVersion     string
	defaultDirname       string
	namespaceDirname     string
	defaultFilename      string
	configV3             string
	configV2             string
	configV1             string
	currentConfigVersion string
	detachedNamespace    string
}

func init() {
	pkg.apiVersionGroup = "iofog.org"
	pkg.latestVersion = "v3"
	pkg.latestAPIVersion = pkg.apiVersionGroup + "/" + pkg.latestVersion
	pkg.defaultDirname = ".iofog/" + pkg.latestVersion
	pkg.namespaceDirname = "namespaces/"
	pkg.defaultFilename = "config.yaml"
	pkg.configV3 = "iofogctl/v3"
	pkg.configV2 = "iofogctl/v2"
	pkg.configV1 = "iofogctl/v1"
	pkg.currentConfigVersion = pkg.configV3
	pkg.detachedNamespace = "_detached"
}

func APIVersion() string {
	return pkg.latestAPIVersion
}

// Init initializes config, namespace and unmarshalls the files
func Init(configFolderArg string) {
	pkg.namespaces = make(map[string]*rsc.Namespace)

	var err error
	pkg.configFolder, err = util.FormatPath(configFolderArg)
	util.Check(err)

	if pkg.configFolder == "" {
		// Find home directory.
		home, err := homedir.Dir()
		util.Check(err)
		pkg.configFolder = path.Join(home, pkg.defaultDirname)
	} else {
		dirInfo, err := os.Stat(pkg.configFolder)
		util.Check(err)
		if !dirInfo.IsDir() {
			util.Check(util.NewInputError(fmt.Sprintf("The config folder %s is not a valid directory", pkg.configFolder)))
		}
	}

	// Set default filename if necessary
	filename := path.Join(pkg.configFolder, pkg.defaultFilename)
	pkg.configFilename = filename
	pkg.namespaceDirectory = path.Join(pkg.configFolder, pkg.namespaceDirname)

	// Check config file already exists
	if _, err := os.Stat(pkg.configFilename); os.IsNotExist(err) {
		err = os.MkdirAll(pkg.configFolder, 0755)
		util.Check(err)

		// Create default config file
		pkg.conf.DefaultNamespace = "default"
		err = flushShared()
		util.Check(err)
	}

	// Unmarshall the config file
	confHeader := iofogctlConfig{}
	err = util.UnmarshalYAML(pkg.configFilename, &confHeader)
	util.Check(err)

	pkg.conf, err = getConfigFromHeader(&confHeader)
	util.Check(err)

	// Check namespace dir exists
	initNamespaces := []string{"default", pkg.detachedNamespace}
	flush := false
	for _, initNamespace := range initNamespaces {
		nsFile := getNamespaceFile(initNamespace)
		if _, err := os.Stat(nsFile); os.IsNotExist(err) {
			flush = true
			err = os.MkdirAll(pkg.namespaceDirectory, 0755)
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
	return path.Join(pkg.namespaceDirectory, name+".yaml")
}

func updateConfigToV3(header *iofogctlConfig) {
	if header != nil {
		header.APIVersion = pkg.configV3
	}
}

func getConfigFromHeader(header *iofogctlConfig) (conf configuration, err error) {
	switch header.APIVersion {
	case pkg.currentConfigVersion:
		// All good
		break
	// Example for further maintenance
	// case PreviousConfigVersion
	// 	updateFromPreviousVersion()
	// 	break
	case pkg.configV1:
		updateConfigToV3(header)
		return getConfigFromHeader(header)
	default:
		return conf, util.NewInputError("Invalid iofogctl config version")
	}
	bytes, err := yaml.Marshal(header.Spec)
	if err != nil {
		return
	}
	if err = yaml.Unmarshal(bytes, &conf); err != nil {
		return
	}
	return conf, err
}

func getNamespaceFromHeader(header *iofogctlNamespace) (ns *rsc.Namespace, err error) {
	// Check header not supported
	switch header.APIVersion {
	case pkg.currentConfigVersion:
		// All good
		break
	case pkg.configV1:
		err = util.NewError("Namespace file is out of date.")
		return
	default:
		err = util.NewInputError("Invalid iofogctl config version")
		return
	}
	// Unmarshal Namespace spec
	bytes, err := yaml.Marshal(header.Spec)
	if err != nil {
		return
	}
	ns = new(rsc.Namespace)
	if err = yaml.Unmarshal(bytes, &ns); err != nil {
		return
	}
	return
}

func getConfigYAMLFile(conf configuration) ([]byte, error) {
	confHeader := iofogctlConfig{
		Header: Header{
			Kind:       IofogctlConfigKind,
			APIVersion: pkg.currentConfigVersion,
			Spec:       conf,
		},
	}

	return yaml.Marshal(confHeader)
}

func getNamespaceYAMLFile(ns *rsc.Namespace) ([]byte, error) {
	namespaceHeader := iofogctlNamespace{
		Header{
			Kind:       IofogctlNamespaceKind,
			APIVersion: pkg.currentConfigVersion,
			Metadata: HeaderMetadata{
				Name: ns.Name,
			},
			Spec: ns,
		},
	}
	return yaml.Marshal(namespaceHeader)
}

func flushNamespaces() error {
	for _, ns := range pkg.namespaces {
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
	marshal, err := getConfigYAMLFile(pkg.conf)
	if err != nil {
		return nil
	}
	// Overwrite the file
	err = ioutil.WriteFile(pkg.configFilename, marshal, 0644)
	if err != nil {
		return nil
	}
	return nil
}

// Flush will write namespace files to disk
func Flush() error {
	return flushNamespaces()
}

func ValidateHeader(header *Header) error {
	if header.APIVersion != pkg.latestAPIVersion {
		msg := fmt.Sprintf("Unsupported YAML API version %s.\nPlease use version %s. See iofog.org for specification details.", header.APIVersion, pkg.latestAPIVersion)
		return util.NewInputError(msg)
	}
	return nil
}
