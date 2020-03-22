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
	"fmt"

	apps "github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	deleteagent "github.com/eclipse-iofog/iofogctl/v2/internal/delete/agent"
	deleteapplication "github.com/eclipse-iofog/iofogctl/v2/internal/delete/application"
	deletecatalogitem "github.com/eclipse-iofog/iofogctl/v2/internal/delete/catalog_item"
	deletecontroller "github.com/eclipse-iofog/iofogctl/v2/internal/delete/controller"
	deletecontrolplane "github.com/eclipse-iofog/iofogctl/v2/internal/delete/controlplane"
	deletemicroservice "github.com/eclipse-iofog/iofogctl/v2/internal/delete/microservice"
	deleteregistry "github.com/eclipse-iofog/iofogctl/v2/internal/delete/registry"
	deletevolume "github.com/eclipse-iofog/iofogctl/v2/internal/delete/volume"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
)

type Options struct {
	Namespace string
	InputFile string
	Soft      bool
}

var kindOrder = []apps.Kind{
	config.CatalogItemKind,
	apps.MicroserviceKind,
	apps.ApplicationKind,
	config.RegistryKind,
	config.AgentKind,
	config.ControllerKind,
	config.ControlPlaneKind,
	config.VolumeKind,
}

var kindHandlers = map[apps.Kind]func(execute.KindHandlerOpt) (execute.Executor, error){
	apps.ApplicationKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deleteapplication.NewExecutor(opt.Namespace, opt.Name)
	},
	apps.MicroserviceKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletemicroservice.NewExecutor(opt.Namespace, opt.Name)
	},
	config.ControlPlaneKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecontrolplane.NewExecutor(opt.Namespace, opt.Name, false)
	},
	config.AgentKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deleteagent.NewExecutor(opt.Namespace, opt.Name, false, false)
	},
	config.ControllerKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecontroller.NewExecutor(opt.Namespace, opt.Name, false)
	},
	config.CatalogItemKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecatalogitem.NewExecutor(opt.Namespace, opt.Name)
	},
	config.RegistryKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deleteregistry.NewExecutor(opt.Namespace, opt.Name)
	},
	config.VolumeKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletevolume.NewExecutor(opt.Namespace, opt.Name)
	},
}

func Execute(opt *Options) error {
	executorsMap, err := execute.GetExecutorsFromYAML(opt.InputFile, opt.Namespace, kindHandlers)
	if err != nil {
		return err
	}

	// Microservice, Application, Agent, Controller, ControlPlane
	for idx := range kindOrder {
		if err = execute.RunExecutors(executorsMap[kindOrder[idx]], fmt.Sprintf("delete %s", kindOrder[idx])); err != nil {
			return err
		}
	}

	return nil
}
