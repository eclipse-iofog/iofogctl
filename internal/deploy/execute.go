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
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployagent "github.com/eclipse-iofog/iofogctl/internal/deploy/agent"
	deployagentconfig "github.com/eclipse-iofog/iofogctl/internal/deploy/agent_config"
	deployapplication "github.com/eclipse-iofog/iofogctl/internal/deploy/application"
	deploycatalogitem "github.com/eclipse-iofog/iofogctl/internal/deploy/catalog_item"
	deploycontroller "github.com/eclipse-iofog/iofogctl/internal/deploy/controller"
	deploycontrolplane "github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane"
	deploymicroservice "github.com/eclipse-iofog/iofogctl/internal/deploy/microservice"
	deployregistry "github.com/eclipse-iofog/iofogctl/internal/deploy/registry"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

var kindOrder = []apps.Kind{
	// Deploy Agents after Control Plane
	// apps.ControlPlaneKind,
	// apps.ControllerKind,
	// apps.AgentKind,
	// config.AgentConfigKind,
	config.RegistryKind,
	config.CatalogItemKind,
	apps.ApplicationKind,
	apps.MicroserviceKind,
}

type Options struct {
	Namespace string
	InputFile string
}

func deployCatalogItem(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploycatalogitem.NewExecutor(deploycatalogitem.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployApplication(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployapplication.NewExecutor(deployapplication.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployMicroservice(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploymicroservice.NewExecutor(deploymicroservice.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployControlPlane(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploycontrolplane.NewExecutor(deploycontrolplane.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployAgent(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployagent.NewExecutor(deployagent.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployAgentConfig(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployagentconfig.NewExecutor(deployagentconfig.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployController(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploycontroller.NewExecutor(deploycontroller.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployRegistry(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployregistry.NewExecutor(deployregistry.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

var kindHandlers = map[apps.Kind]func(execute.KindHandlerOpt) (execute.Executor, error){
	apps.ApplicationKind:   deployApplication,
	config.CatalogItemKind: deployCatalogItem,
	apps.MicroserviceKind:  deployMicroservice,
	apps.ControlPlaneKind:  deployControlPlane,
	apps.AgentKind:         deployAgent,
	config.AgentConfigKind: deployAgentConfig,
	apps.ControllerKind:    deployController,
	config.RegistryKind:    deployRegistry,
}

// Execute deploy from yaml file
func Execute(opt *Options) (err error) {
	executorsMap, err := execute.GetExecutorsFromYAML(opt.InputFile, opt.Namespace, kindHandlers)
	if err != nil {
		return err
	}

	// Create any AgentConfig executor missing
	// Each Agent requires a corresponding Agent Config to be created with Controller
	for _, agentGenericExecutor := range executorsMap[apps.AgentKind] {
		agentExecutor, ok := agentGenericExecutor.(deployagent.AgentDeployExecutor)
		if !ok {
			return util.NewInternalError("Could not convert agent deploy executor\n")
		}
		found := false
		host := agentExecutor.GetHost()
		for _, configGenericExecutor := range executorsMap[config.AgentConfigKind] {
			configExecutor, ok := configGenericExecutor.(deployagentconfig.AgentConfigExecutor)
			if !ok {
				return util.NewInternalError("Could not convert agent config executor\n")
			}
			if agentExecutor.GetName() == configExecutor.GetName() {
				found = true
				configExecutor.SetHost(host)
				break
			}
		}
		if !found {
			executorsMap[config.AgentConfigKind] = append(executorsMap[config.AgentConfigKind], deployagentconfig.NewRemoteExecutor(
				agentExecutor.GetName(),
				config.AgentConfiguration{
					Name: agentExecutor.GetName(),
					AgentConfiguration: client.AgentConfiguration{
						Host: &host,
					},
				},
				opt.Namespace,
			))
		}
	}

	// Execute in parallel by priority order

	// Controlplane
	if err = execute.RunExecutors(executorsMap[apps.ControlPlaneKind], "deploy control plane"); err != nil {
		return
	}

	// Controller
	if err = execute.RunExecutors(executorsMap[apps.ControllerKind], "deploy controller"); err != nil {
		return
	}

	// Agent config are the representation of agents in Controller. They need to be deployed sequentially because of router dependencies
	for idx := range executorsMap[config.AgentConfigKind] {
		if err = executorsMap[config.AgentConfigKind][idx].Execute(); err != nil {
			return err
		}
	}

	// Agents are the actual remote host installation, they can be installed in parallel
	if err = execute.RunExecutors(executorsMap[apps.AgentKind], "deploy agent"); err != nil {
		return
	}

	// CatalogItem, Application, Microservice
	for idx := range kindOrder {
		if err = execute.RunExecutors(executorsMap[kindOrder[idx]], fmt.Sprintf("deploy %s", kindOrder[idx])); err != nil {
			return
		}
	}

	return nil
}
