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
	"io"
	"io/ioutil"

	apps "github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployagent "github.com/eclipse-iofog/iofogctl/internal/deploy/agent"
	deployapplication "github.com/eclipse-iofog/iofogctl/internal/deploy/application"
	deployconnector "github.com/eclipse-iofog/iofogctl/internal/deploy/connector"
	deploycontroller "github.com/eclipse-iofog/iofogctl/internal/deploy/controller"
	deploycontrolplane "github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane"
	deploymicroservice "github.com/eclipse-iofog/iofogctl/internal/deploy/microservice"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

var kindOrder = []apps.Kind{
	// Connector cannot be ran in parallel.
	// apps.ControlPlaneKind,
	// apps.ControllerKind,
	// apps.ConnectorKind,
	apps.AgentKind,
	apps.ApplicationKind,
	apps.MicroserviceKind,
}

type Options struct {
	Namespace string
	InputFile string
}

func deployApplication(namespace, name string, yaml []byte) (exe execute.Executor, err error) {
	return deployapplication.NewExecutor(deployapplication.Options{Namespace: namespace, Yaml: yaml, Name: name})
}

func deployMicroservice(namespace, name string, yaml []byte) (exe execute.Executor, err error) {
	return deploymicroservice.NewExecutor(deploymicroservice.Options{Namespace: namespace, Yaml: yaml, Name: name})
}

func deployControlPlane(namespace, name string, yaml []byte) (exe execute.Executor, err error) {
	return deploycontrolplane.NewExecutor(deploycontrolplane.Options{Namespace: namespace, Yaml: yaml, Name: name})
}

func deployAgent(namespace, name string, yaml []byte) (exe execute.Executor, err error) {
	return deployagent.NewExecutor(deployagent.Options{Namespace: namespace, Yaml: yaml, Name: name})
}

func deployConnector(namespace, name string, yaml []byte) (exe execute.Executor, err error) {
	return deployconnector.NewExecutor(deployconnector.Options{Namespace: namespace, Yaml: yaml, Name: name})
}

func deployController(namespace, name string, yaml []byte) (exe execute.Executor, err error) {
	return deploycontroller.NewExecutor(deploycontroller.Options{Namespace: namespace, Yaml: yaml, Name: name})
}

var kindHandlers = map[apps.Kind]func(string, string, []byte) (execute.Executor, error){
	apps.ApplicationKind:  deployApplication,
	apps.MicroserviceKind: deployMicroservice,
	apps.ControlPlaneKind: deployControlPlane,
	apps.AgentKind:        deployAgent,
	apps.ConnectorKind:    deployConnector,
	apps.ControllerKind:   deployController,
}

func execDocument(header config.Header, namespace string) (exe execute.Executor, err error) {
	// Check namespace exists
	if len(header.Metadata.Namespace) > 0 {
		namespace = header.Metadata.Namespace
	}
	if _, err := config.GetNamespace(namespace); err != nil {
		return exe, err
	}

	subYamlBytes, err := yaml.Marshal(header.Spec)
	if err != nil {
		return exe, err
	}

	createExecutorf, found := kindHandlers[header.Kind]
	if !found {
		return exe, util.NewInputError("Invalid kind")
	}

	return createExecutorf(namespace, header.Metadata.Name, subYamlBytes)
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
	header := config.Header{
		Spec: raw,
	}

	// Generate all executors
	executorsMap := make(map[apps.Kind][]execute.Executor)
	decodeErr := dec.Decode(&header)
	for decodeErr == nil {
		exe, err := execDocument(header, namespace)
		if err != nil {
			return err
		}
		executorsMap[header.Kind] = append(executorsMap[header.Kind], exe)
		decodeErr = dec.Decode(&header)
	}

	if decodeErr != io.EOF && decodeErr != nil {
		return err
	}

	// Execute in parallel by priority order
	// Connector cannot be deployed in parallel

	// Controlplane
	if err = executeKind(executorsMap[apps.ControlPlaneKind]); err != nil {
		return
	}

	// Controller
	if err = executeKind(executorsMap[apps.ControllerKind]); err != nil {
		return
	}

	// Connector
	for idx := range executorsMap[apps.ConnectorKind] {
		if err = executorsMap[apps.ConnectorKind][idx].Execute(); err != nil {
			util.PrintNotify("Error from " + executorsMap[apps.ConnectorKind][idx].GetName() + ": " + err.Error())
			return util.NewError("Failed to deploy")
		}
	}

	// Agents, Application, Microservice
	for idx := range kindOrder {
		if err = executeKind(executorsMap[kindOrder[idx]]); err != nil {
			return
		}
	}

	return nil
}

func executeKind(executors []execute.Executor) (err error) {
	if errs, failedExes := execute.ForParallel(executors); len(errs) > 0 {
		for idx := range errs {
			util.PrintNotify("Error from " + failedExes[idx].GetName() + ": " + errs[idx].Error())
		}
		return util.NewError("Failed to deploy")
	}
	return nil
}
