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

	deploy "github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
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

var kindOrder = []deploy.Kind{
	// Connector cannot be ran in parallel.
	// deploy.ControlPlaneKind,
	// deploy.ControllerKind,
	// deploy.ConnectorKind,
	deploy.AgentKind,
	deploy.ApplicationKind,
	deploy.MicroserviceKind,
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

var kindHandlers = map[deploy.Kind]func(string, string, []byte) (execute.Executor, error){
	deploy.ApplicationKind:  deployApplication,
	deploy.MicroserviceKind: deployMicroservice,
	deploy.ControlPlaneKind: deployControlPlane,
	deploy.AgentKind:        deployAgent,
	deploy.ConnectorKind:    deployConnector,
	deploy.ControllerKind:   deployController,
}

func execDocument(header deploy.Header, namespace string) (exe execute.Executor, err error) {
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
	header := deploy.Header{
		Spec: raw,
	}

	// Generate all executors
	executorsMap := make(map[deploy.Kind][]execute.Executor)
	for dec.Decode(&header) == nil {
		exe, err := execDocument(header, namespace)
		if err != nil {
			return err
		}
		executorsMap[header.Kind] = append(executorsMap[header.Kind], exe)
	}

	// Execute in parallel by priority order
	// Connector cannot be deployed in parallel

	// Controlplane
	if err = executeKind(executorsMap[deploy.ControlPlaneKind]); err != nil {
		return
	}

	// Controller
	if err = executeKind(executorsMap[deploy.ControllerKind]); err != nil {
		return
	}

	// Connector
	for idx := range executorsMap[deploy.ConnectorKind] {
		if err = executorsMap[deploy.ConnectorKind][idx].Execute(); err != nil {
			util.PrintNotify("Error from " + executorsMap[deploy.ConnectorKind][idx].GetName() + ": " + err.Error())
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
