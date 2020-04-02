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
	"sync"

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

func getNamespaceFromHeader(header iofogctlNamespace) (n rsc.Namespace, err error) {
	switch header.APIVersion {
	case CurrentConfigVersion:
		{
			// All good
			break
		}
	case configV1:
		{
			msg := `An older YAML version has been detected in Namespace %s.
  You will only be able to view this Namespace from your current version of iofogctl.
  Redeploy the corresponding ECN with your current version of iofogctl to gain full control.`
			util.PrintNotify(fmt.Sprintf(msg, header.Metadata.Name))
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

type v1NamespaceSpecContent struct {
	Name         string                `yaml:"name,omitempty"`
	ControlPlane configv1.ControlPlane `yaml:"controlPlane,omitempty"`
	Agents       []configv1.Agent      `yaml:"agents,omitempty"`
	Created      string                `yaml:"created,omitempty"`
	Connectors   []configv1.Connector  `yaml:"connectors,omitempty"`
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

func updateNamespaceToV2(header iofogctlNamespace) (iofogctlNamespace, error) {
	header.APIVersion = configV2
	bytes, err := yaml.Marshal(header.Spec)
	nsV1 := configv1.Namespace{}
	if err != nil {
		return header, err
	}
	if err = yaml.UnmarshalStrict(bytes, &nsV1); err != nil {
		return header, err
	}

	// Namespace
	ns := rsc.Namespace{
		Name:    nsV1.Name,
		Created: nsV1.Created,
	}

	// Finish
	if len(nsV1.ControlPlane.Controllers) == 0 {
		header.Spec = ns
		return header, nil
	}

	// Get Controller to determine Control Plane type
	controllerV1 := nsV1.ControlPlane.Controllers[0]

	// Local Agents
	if util.IsLocalHost(controllerV1.Host) {
		for _, agentV1 := range nsV1.Agents {
			agent := rsc.LocalAgent{
				Name:    agentV1.Name,
				UUID:    agentV1.UUID,
				Created: agentV1.Created,
				Container: rsc.Container{
					Image:       agentV1.Container.Image,
					Credentials: rsc.Credentials(agentV1.Container.Credentials),
				},
			}
			if err := ns.AddAgent(&agent); err != nil {
				return header, err
			}
		}
	} else {
		// Remote Agents
		for _, agentV1 := range nsV1.Agents {
			agent := rsc.RemoteAgent{
				Name:    agentV1.Name,
				Host:    agentV1.Host,
				SSH:     rsc.SSH(agentV1.SSH),
				UUID:    agentV1.UUID,
				Created: agentV1.Created,
				Package: rsc.Package(agentV1.Package),
			}
			if err := ns.AddAgent(&agent); err != nil {
				return header, err
			}
		}
	}

	// Local Control Plane
	if util.IsLocalHost(controllerV1.Host) {
		ns.LocalControlPlane = &rsc.LocalControlPlane{
			Controller: &rsc.LocalController{
				Name:     controllerV1.Name,
				Endpoint: controllerV1.Endpoint,
				Created:  controllerV1.Created,
				Container: rsc.Container{
					Image:       controllerV1.Container.Image,
					Credentials: rsc.Credentials(controllerV1.Container.Credentials),
				},
			},
		}
		header.Spec = ns
		return header, nil
	}

	// K8s control plane
	if controllerV1.Kube.Config != "" {
		for idx, controllerV1 := range nsV1.ControlPlane.Controllers {
			if ns.KubernetesControlPlane == nil {
				ns.KubernetesControlPlane = &rsc.KubernetesControlPlane{
					KubeConfig: controllerV1.Kube.Config,
					Services: rsc.Services{
						Controller: rsc.Service{
							Type: controllerV1.Kube.ServiceType,
							IP:   controllerV1.Kube.StaticIP,
						},
					},
					Replicas: rsc.Replicas{
						Controller: int32(controllerV1.Kube.Replicas),
					},
					Images: rsc.KubeImages{
						Controller: controllerV1.Container.Image,
						Operator:   controllerV1.Kube.Images.Operator,
						Kubelet:    controllerV1.Kube.Images.Kubelet,
					},
					Endpoint: controllerV1.Endpoint,
				}
				pod := rsc.KubernetesController{
					PodName:  fmt.Sprintf("kubernetes-%d", idx+1),
					Endpoint: controllerV1.Endpoint,
					Created:  controllerV1.Created,
				}
				if err := ns.KubernetesControlPlane.AddController(&pod); err != nil {
					return header, err
				}
			}
		}
		header.Spec = ns
		return header, nil
	}

	// Remote Control Plane
	for _, controllerV1 := range nsV1.ControlPlane.Controllers {
		if ns.RemoteControlPlane == nil {
			ns.RemoteControlPlane = new(rsc.RemoteControlPlane)
		}
		ctrl := rsc.RemoteController{
			Name:     controllerV1.Name,
			Host:     controllerV1.Host,
			SSH:      rsc.SSH(controllerV1.SSH),
			Endpoint: controllerV1.Endpoint,
			Created:  controllerV1.Created,
			Package:  rsc.Package(controllerV1.Package),
		}
		if err := ns.RemoteControlPlane.AddController(&ctrl); err != nil {
			return header, err
		}
	}

	header.Spec = ns
	return header, nil
}
