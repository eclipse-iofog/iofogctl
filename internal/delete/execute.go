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

package delete

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	deleteagent "github.com/eclipse-iofog/iofogctl/v3/internal/delete/agent"
	deleteapplication "github.com/eclipse-iofog/iofogctl/v3/internal/delete/application"
	deletecatalogitem "github.com/eclipse-iofog/iofogctl/v3/internal/delete/catalogitem"
	deletecontroller "github.com/eclipse-iofog/iofogctl/v3/internal/delete/controller"
	deletek8scontrolplane "github.com/eclipse-iofog/iofogctl/v3/internal/delete/controlplane/k8s"
	deletelocalcontrolplane "github.com/eclipse-iofog/iofogctl/v3/internal/delete/controlplane/local"
	deleteremotecontrolplane "github.com/eclipse-iofog/iofogctl/v3/internal/delete/controlplane/remote"
	deletemicroservice "github.com/eclipse-iofog/iofogctl/v3/internal/delete/microservice"
	deleteregistry "github.com/eclipse-iofog/iofogctl/v3/internal/delete/registry"
	deletevolume "github.com/eclipse-iofog/iofogctl/v3/internal/delete/volume"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
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
	config.RemoteAgentKind,
	config.LocalAgentKind,
	config.RemoteControllerKind,
	config.LocalControllerKind,
	config.KubernetesControlPlaneKind,
	config.RemoteControlPlaneKind,
	config.LocalControlPlaneKind,
	config.VolumeKind,
}

var kindHandlers = map[config.Kind]func(*execute.KindHandlerOpt) (execute.Executor, error){
	config.ApplicationKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deleteapplication.NewExecutor(opt.Namespace, opt.Name)
	},
	config.MicroserviceKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletemicroservice.NewExecutor(opt.Namespace, opt.Name)
	},
	config.KubernetesControlPlaneKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletek8scontrolplane.NewExecutor(opt.Namespace)
	},
	config.RemoteControlPlaneKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deleteremotecontrolplane.NewExecutor(opt.Namespace)
	},
	config.LocalControlPlaneKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletelocalcontrolplane.NewExecutor(opt.Namespace)
	},
	config.RemoteControllerKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecontroller.NewExecutor(opt.Namespace, opt.Name)
	},
	config.LocalControllerKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecontroller.NewExecutor(opt.Namespace, opt.Name)
	},
	config.RemoteAgentKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deleteagent.NewExecutor(opt.Namespace, opt.Name, false, false)
	},
	config.LocalAgentKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deleteagent.NewExecutor(opt.Namespace, opt.Name, false, false)
	},
	config.CatalogItemKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecatalogitem.NewExecutor(opt.Namespace, opt.Name)
	},
	config.RegistryKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deleteregistry.NewExecutor(opt.Namespace, opt.Name)
	},
	config.VolumeKind: func(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
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
		if errs := execute.RunExecutors(executorsMap[kindOrder[idx]], fmt.Sprintf("delete %s", kindOrder[idx])); len(errs) > 0 {
			for _, err := range errs {
				if _, ok := err.(*util.NotFoundError); !ok {
					return execute.CoalesceErrors(errs)
				}
				util.PrintNotify(fmt.Sprintf("Warning: %s %s.", kindOrder[idx], err.Error()))
			}
		}
	}

	return nil
}
