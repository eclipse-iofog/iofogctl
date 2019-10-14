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

package delete

import (
	"bytes"
	"io/ioutil"

	apps "github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deleteagent "github.com/eclipse-iofog/iofogctl/internal/delete/agent"
	deleteapplication "github.com/eclipse-iofog/iofogctl/internal/delete/application"
	deleteconnector "github.com/eclipse-iofog/iofogctl/internal/delete/connector"
	deletecontroller "github.com/eclipse-iofog/iofogctl/internal/delete/controller"
	deletecontrolplane "github.com/eclipse-iofog/iofogctl/internal/delete/controlplane"
	deletemicroservice "github.com/eclipse-iofog/iofogctl/internal/delete/microservice"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	InputFile string
}

var kindOrder = []apps.Kind{
	apps.MicroserviceKind,
	apps.ApplicationKind,
	apps.AgentKind,
	apps.ConnectorKind,
	apps.ControllerKind,
	apps.ControlPlaneKind,
}

var kindHandlers = map[apps.Kind]func(string, string) (execute.Executor, error){
	apps.ApplicationKind: func(namespace, name string) (exe execute.Executor, err error) {
		return deleteapplication.NewExecutor(namespace, name)
	},
	apps.MicroserviceKind: func(namespace, name string) (exe execute.Executor, err error) {
		return deletemicroservice.NewExecutor(namespace, name)
	},
	apps.ControlPlaneKind: func(namespace, name string) (exe execute.Executor, err error) {
		return deletecontrolplane.NewExecutor(namespace, name)
	},
	apps.AgentKind: func(namespace, name string) (exe execute.Executor, err error) {
		return deleteagent.NewExecutor(namespace, name)
	},
	apps.ConnectorKind: func(namespace, name string) (exe execute.Executor, err error) {
		return deleteconnector.NewExecutor(namespace, name)
	},
	apps.ControllerKind: func(namespace, name string) (exe execute.Executor, err error) {
		return deletecontroller.NewExecutor(namespace, name)
	},
}

func execDocument(header config.Header, namespace string) (exe execute.Executor, err error) {
	// Check namespace exists
	if len(header.Metadata.Namespace) > 0 {
		namespace = header.Metadata.Namespace
	}
	if _, err := config.GetNamespace(namespace); err != nil {
		return exe, err
	}

	createExecutorf, found := kindHandlers[header.Kind]
	if !found {
		return exe, util.NewInputError("Invalid kind")
	}

	return createExecutorf(namespace, header.Metadata.Name)
}

func Execute(opt *Options) error {
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
	for dec.Decode(&header) == nil {
		exe, err := execDocument(header, namespace)
		if err != nil {
			return err
		}
		executorsMap[header.Kind] = append(executorsMap[header.Kind], exe)
	}

	// Microservice, Application, Agent, Connector, Controller, ControlPlane
	for idx := range kindOrder {
		if err = runExecutors(executorsMap[kindOrder[idx]]); err != nil {
			return err
		}
	}

	return nil
}

func runExecutors(executors []execute.Executor) error {
	if errs, failedExes := execute.ForParallel(executors); len(errs) > 0 {
		for idx := range errs {
			util.PrintNotify("Error from " + failedExes[idx].GetName() + ": " + errs[idx].Error())
		}
		return util.NewError("Failed to delete")
	}
	return nil
}
