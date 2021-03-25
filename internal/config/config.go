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
	"fmt"
	"io/ioutil"
	"os"
	"path"

	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	homedir "github.com/mitchellh/go-homedir"
	yaml "gopkg.in/yaml.v2"
)

type versionKey = string
type namespaceKey = string
type versionedNamespaceIndex map[versionKey]map[namespaceKey]*rsc.Namespace

var pkg struct {
	conf              configuration
	configFolder      string // config directory
	configFilename    string // config file name
	nsDirs            map[versionKey]string
	nsIndex           versionedNamespaceIndex
	apiVersionGroup   string
	latestVersion     string
	supportedVersions []string
	latestAPIVersion  string
	rootDir           string
	defaultDirname    string
	namespaceDirname  string
	defaultFilename   string
	apiV3             string
	apiV2             string
	apiV1             string
	detachedNamespace string
}

func init() {
	pkg.nsIndex = versionedNamespaceIndex{}
	pkg.nsDirs = map[versionKey]string{}
	pkg.apiVersionGroup = "iofog.org"
	pkg.latestVersion = "v3"
	pkg.supportedVersions = []string{"v3", "v2"}
	pkg.latestAPIVersion = pkg.apiVersionGroup + "/" + pkg.latestVersion
	pkg.rootDir = ".iofog/"
	pkg.defaultDirname = pkg.rootDir + pkg.latestVersion
	pkg.namespaceDirname = "namespaces/"
	pkg.defaultFilename = "config.yaml"
	pkg.apiV3 = "iofogctl/v3"
	pkg.apiV2 = "iofogctl/v2"
	pkg.apiV1 = "iofogctl/v1"
	pkg.detachedNamespace = "_detached"
}

func APIVersion() string {
	return pkg.latestAPIVersion
}

// Init initializes config, namespace and unmarshalls the files
func Init() error {
	for _, vers := range pkg.supportedVersions {
		pkg.nsIndex[vers] = map[namespaceKey]*rsc.Namespace{}
	}

	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	pkg.configFolder = path.Join(home, pkg.defaultDirname)

	// Set default filename if necessary
	filename := path.Join(pkg.configFolder, pkg.defaultFilename)
	pkg.configFilename = filename
	for _, vers := range pkg.supportedVersions {
		pkg.nsDirs[vers] = path.Join(pkg.rootDir, vers, pkg.namespaceDirname)
	}

	// Check config file already exists
	if _, err := os.Stat(pkg.configFilename); os.IsNotExist(err) {
		if err := os.MkdirAll(pkg.configFolder, 0755); err != nil {
			return err
		}

		// Create default config file
		pkg.conf.DefaultNamespace = "default"
		if err := flushShared(); err != nil {
			return err
		}
	}

	// Unmarshall the config file
	confHeader := iofogctlConfig{}
	if err := util.UnmarshalYAML(pkg.configFilename, &confHeader); err != nil {
		return err
	}

	pkg.conf, err = getConfigFromHeader(&confHeader)
	if err != nil {
		return err
	}

	// Check namespace dirs exists
	initNamespaces := []string{"default", pkg.detachedNamespace}
	flush := false
	for _, initNamespace := range initNamespaces {
		for _, vers := range pkg.supportedVersions {
			nsFile := getNamespaceFile(initNamespace, vers)
			if _, err := os.Stat(nsFile); os.IsNotExist(err) {
				flush = true
				if err := os.MkdirAll(getNamespaceDir(vers), 0755); err != nil {
					return err
				}

				// Create default namespace file
				if err := addNamespace(initNamespace, util.NowUTC(), vers); err != nil {
					return fmt.Errorf("Could not initialize %s configuration", initNamespace)
				}
			} else {
				if initNamespace == "default" && vers != pkg.latestVersion {
					// Move old default ns
					if err := os.Rename(nsFile, getNamespaceFile("default"+vers, vers)); err != nil {
						return fmt.Errorf("Failed to rename %s default namespace", vers)
					}
				}
			}
		}
	}
	if flush {
		if err := flushNamespaces(); err != nil {
			return err
		}
	}
	return nil
}

func getNamespaceDir(version string) string {
	return path.Join(pkg.rootDir, version, pkg.namespaceDirname)
}

// getNamespaceFile helper function that returns the full path to a namespace file
func getNamespaceFile(name, version string) string {
	return path.Join(getNamespaceDir(version), name+".yaml")
}

func getConfigFromHeader(header *iofogctlConfig) (conf configuration, err error) {
	switch header.APIVersion {
	default:
		return conf, util.NewInputError("Invalid iofogctl config version")
	case pkg.apiV3:
		fallthrough
	case pkg.apiV2:
		fallthrough
	case pkg.apiV1:
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
	notSupported := "Config version %s is not supported"
	switch header.APIVersion {
	case pkg.apiV2:
		fallthrough
	case pkg.apiV3:
		// All good
		break
	case pkg.apiV1:
		err = util.NewError(fmt.Sprintf(notSupported, pkg.apiV1))
		return
	default:
		err = util.NewInputError(fmt.Sprintf(notSupported, header.APIVersion))
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

func getConfigYAMLFile(conf configuration, apiVersion string) ([]byte, error) {
	confHeader := iofogctlConfig{
		Header: Header{
			Kind:       IofogctlConfigKind,
			APIVersion: apiVersion,
			Spec:       conf,
		},
	}

	return yaml.Marshal(confHeader)
}

func getNamespaceYAMLFile(ns *rsc.Namespace, apiVersion string) ([]byte, error) {
	namespaceHeader := iofogctlNamespace{
		Header{
			Kind:       IofogctlNamespaceKind,
			APIVersion: apiVersion,
			Metadata: HeaderMetadata{
				Name: ns.Name,
			},
			Spec: ns,
		},
	}
	return yaml.Marshal(namespaceHeader)
}

func flushNamespaces() error {
	for vers, namespaces := range pkg.nsIndex {
		for _, ns := range namespaces {
			// Marshal the runtime data
			marshal, err := getNamespaceYAMLFile(ns, vers)
			if err != nil {
				return err
			}
			// Overwrite the file
			err = ioutil.WriteFile(getNamespaceFile(ns.Name, vers), marshal, 0644)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func flushShared() error {
	// Marshal the runtime data
	marshal, err := getConfigYAMLFile(pkg.conf, pkg.latestAPIVersion)
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
