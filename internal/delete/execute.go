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

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	deleteagent "github.com/eclipse-iofog/iofogctl/v2/internal/delete/agent"
	deleteapplication "github.com/eclipse-iofog/iofogctl/v2/internal/delete/application"
	deletecatalogitem "github.com/eclipse-iofog/iofogctl/v2/internal/delete/catalogitem"
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

var kindOrder = []config.Kind{
	config.CatalogItemKind,
	config.MicroserviceKind,
	config.ApplicationKind,
	config.RegistryKind,
	config.AgentKind,
	config.KubernetesControllerKind,
	config.RemoteControllerKind,
	config.LocalControllerKind,
	config.KubernetesControlPlaneKind,
	config.RemoteControlPlaneKind,
	config.LocalControlPlaneKind,
	config.VolumeKind,
}

var kindHandlers = map[config.Kind]func(execute.KindHandlerOpt) (execute.Executor, error){
	config.ApplicationKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deleteapplication.NewExecutor(opt.Namespace, opt.Name)
	},
	config.MicroserviceKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletemicroservice.NewExecutor(opt.Namespace, opt.Name)
	},
	config.KubernetesControlPlaneKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecontrolplane.NewExecutor(opt.Namespace, opt.Name, false)
	},
	config.RemoteControlPlaneKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecontrolplane.NewExecutor(opt.Namespace, opt.Name, false)
	},
	config.LocalControlPlaneKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecontrolplane.NewExecutor(opt.Namespace, opt.Name, false)
	},
	config.KubernetesControllerKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecontroller.NewExecutor(opt.Namespace, opt.Name, false)
	},
	config.RemoteControllerKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecontroller.NewExecutor(opt.Namespace, opt.Name, false)
	},
	config.LocalControllerKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecontroller.NewExecutor(opt.Namespace, opt.Name, false)
	},
	config.AgentKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deleteagent.NewExecutor(opt.Namespace, opt.Name, false, false)
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
