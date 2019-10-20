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
	"fmt"

	apps "github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	deployagent "github.com/eclipse-iofog/iofogctl/internal/deploy/agent"
	deployapplication "github.com/eclipse-iofog/iofogctl/internal/deploy/application"
	deployconnector "github.com/eclipse-iofog/iofogctl/internal/deploy/connector"
	deploycontroller "github.com/eclipse-iofog/iofogctl/internal/deploy/controller"
	deploycontrolplane "github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane"
	deploymicroservice "github.com/eclipse-iofog/iofogctl/internal/deploy/microservice"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
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

// Execute deploy from yaml file
func Execute(opt *Options) (err error) {
	executorsMap, err := execute.GetExecutorsFromYAML(opt.InputFile, opt.Namespace, kindHandlers)
	if err != nil {
		return err
	}

	// Execute in parallel by priority order
	// Connector cannot be deployed in parallel

	// Controlplane
	if err = execute.RunExecutors(executorsMap[apps.ControlPlaneKind], "deploy control plane"); err != nil {
		return
	}

	// Controller
	if err = execute.RunExecutors(executorsMap[apps.ControllerKind], "deploy controller"); err != nil {
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
		if err = execute.RunExecutors(executorsMap[kindOrder[idx]], fmt.Sprintf("deploy %s", kindOrder[idx])); err != nil {
			return
		}
	}

	return nil
}
