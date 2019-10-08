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

package deploy

import (
	"bytes"
	"io/ioutil"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployagent "github.com/eclipse-iofog/iofogctl/internal/deploy/agent"
	deployapplication "github.com/eclipse-iofog/iofogctl/internal/deploy/application"
	deployconnector "github.com/eclipse-iofog/iofogctl/internal/deploy/connector"
	deploycontroller "github.com/eclipse-iofog/iofogctl/internal/deploy/controller"
	deploycontrolplane "github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane"
	deploymicroservice "github.com/eclipse-iofog/iofogctl/internal/deploy/microservice"
	deploy "github.com/eclipse-iofog/iofogctl/pkg/iofog/deploy"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	InputFile string
}

func deployApplication(namespace string, yaml []byte) error {
	if err := deployapplication.Execute(deployapplication.Options{Namespace: namespace, Yaml: yaml}); err != nil {
		return err
	}
	return nil
}

func deployMicroservice(namespace string, yaml []byte) error {
	if err := deploymicroservice.Execute(deploymicroservice.Options{Namespace: namespace, Yaml: yaml}); err != nil {
		return err
	}
	return nil
}

func deployControlPlane(namespace string, yaml []byte) error {
	if err := deploycontrolplane.Execute(deploycontrolplane.Options{Namespace: namespace, Yaml: yaml}); err != nil {
		return err
	}
	return nil
}

func deployAgent(namespace string, yaml []byte) error {
	if err := deployagent.Execute(deployagent.Options{Namespace: namespace, Yaml: yaml}); err != nil {
		return err
	}
	return nil
}

func deployConnector(namespace string, yaml []byte) error {
	if err := deployconnector.Execute(deployconnector.Options{Namespace: namespace, Yaml: yaml}); err != nil {
		return err
	}
	return nil
}

func deployController(namespace string, yaml []byte) error {
	if err := deploycontroller.Execute(deploycontroller.Options{Namespace: namespace, Yaml: yaml}); err != nil {
		return err
	}
	return nil
}

var kindHandlers = map[deploy.Kind]func(string, []byte) error{
	deploy.ApplicationKind:  deployApplication,
	deploy.MicroserviceKind: deployMicroservice,
	deploy.ControlPlaneKind: deployControlPlane,
	deploy.AgentKind:        deployAgent,
	deploy.ConnectorKind:    deployConnector,
	deploy.ControllerKind:   deployController,
}

func execDocument(header deploy.Header, namespace string) error {
	// Check namespace exists
	if len(header.Metadata.Namespace) > 0 {
		namespace = header.Metadata.Namespace
	}
	if _, err := config.GetNamespace(namespace); err != nil {
		return err
	}

	subYamlBytes, err := yaml.Marshal(header.Spec)
	if err != nil {
		return err
	}

	deployf, found := kindHandlers[header.Kind]
	if !found {
		return util.NewInputError("Invalid kind")
	}

	return deployf(namespace, subYamlBytes)
}

// Execute deploy from yaml file
func Execute(opt *Options) (err error) {
	yamlFile, err := ioutil.ReadFile(opt.InputFile)
	if err != nil {
		return err
	}

	r := bytes.NewReader(yamlFile)
	dec := yaml.NewDecoder(r)

	namespace := opt.Namespace
	var raw yaml.MapSlice
	header := deploy.Header{
		Spec: raw,
	}

	for dec.Decode(&header) == nil {
		if err = execDocument(header, namespace); err != nil {
			return err
		}
	}
	return nil
}
