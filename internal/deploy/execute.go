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
	"github.com/eclipse-iofog/iofogctl/internal"
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
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/twmb/algoimpl/go/graph"
)

var kindOrder = []apps.Kind{
	// Deploy Agents after Control Plane
	// apps.ControlPlaneKind,
	// apps.ControllerKind,
	// config.AgentConfigKind,
	apps.AgentKind,
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

	// Controlplane
	if err = execute.RunExecutors(executorsMap[apps.ControlPlaneKind], "deploy control plane"); err != nil {
		return
	}

	// Controller
	if err = execute.RunExecutors(executorsMap[apps.ControllerKind], "deploy controller"); err != nil {
		return
	}

	// Agent config
	if err = deployAgentConfiguration(executorsMap[config.AgentConfigKind]); err != nil {
		return err
	}

	// Execute in parallel by priority order
	// Agents, CatalogItem, Application, Microservice
	for idx := range kindOrder {
		if err = execute.RunExecutors(executorsMap[kindOrder[idx]], fmt.Sprintf("deploy %s", kindOrder[idx])); err != nil {
			return
		}
	}

	return nil
}

func deployAgentConfiguration(executors []execute.Executor) (err error) {
	if len(executors) == 0 {
		return nil
	}

	executorsByNamespace := make(map[string][]deployagentconfig.AgentConfigExecutor)

	// Sort executors by namespace
	for idx := range executors {
		// Get a more specific executor allowing retrieval of namespace
		agentConfigExecutor, ok := (executors[idx]).(deployagentconfig.AgentConfigExecutor)
		if !ok {
			return util.NewInternalError("Could not convert node to agent config executor")
		}
		executorsByNamespace[agentConfigExecutor.GetNamespace()] = append(executorsByNamespace[agentConfigExecutor.GetNamespace()], agentConfigExecutor)
	}

	for namespace, executors := range executorsByNamespace {
		// List agents on Controller
		ctrlClient, err := internal.NewControllerClient(namespace)
		if err != nil {
			return err
		}

		listAgentReponse, err := ctrlClient.ListAgents()
		if err != nil {
			return err
		}

		// Get a map for easy access
		agentByName := make(map[string]*client.AgentInfo)
		agentByUUID := make(map[string]*client.AgentInfo)
		for idx := range listAgentReponse.Agents {
			agentByName[listAgentReponse.Agents[idx].Name] = &listAgentReponse.Agents[idx]
			agentByUUID[listAgentReponse.Agents[idx].UUID] = &listAgentReponse.Agents[idx]
		}
		// Add default router
		agentByName[iofog.VanillaRouterAgentName] = &client.AgentInfo{Name: iofog.VanillaRouterAgentName}

		// Agent config are the representation of agents in Controller. They need to be deployed sequentially because of router dependencies
		// First create the acyclic graph of dependencies
		g := graph.New(graph.Directed)
		nodeMap := make(map[string]graph.Node, 0)
		agentNodeMap := make(map[string]graph.Node, 0)

		for idx := range executors {
			// Create node
			nodeMap[executors[idx].GetName()] = g.MakeNode()
			// Make node value to be executor
			*nodeMap[executors[idx].GetName()].Value = executors[idx]
		}

		// Create connections
		for _, node := range nodeMap {
			// Get a more specific executor allowing retrieval of upstream agents
			agentConfigExecutor, ok := (*node.Value).(deployagentconfig.AgentConfigExecutor)
			if !ok {
				return util.NewInternalError("Could not convert node to agent config executor")
			}
			// Set dependencies for agent config topological sort
			configuration := agentConfigExecutor.GetConfiguration()
			dependencies := getDependencies(configuration.UpstreamRouters, configuration.NetworkRouter)
			if err = makeEdges(g, node, nodeMap, agentNodeMap, agentByName, agentByUUID, dependencies); err != nil {
				return err
			}
		}

		// Detect if there is any cyclic graph
		cyclicGraphs := g.StronglyConnectedComponents()
		for _, cyclicGraph := range cyclicGraphs {
			if len(cyclicGraph) > 1 {
				cyclicAgentsNames := []string{}
				for _, node := range cyclicGraph {
					executor := (*node.Value).(execute.Executor)
					cyclicAgentsNames = append(cyclicAgentsNames, executor.GetName())
				}
				return util.NewInputError(fmt.Sprintf("Cyclic dependencies between agent configurations: %v\n", cyclicAgentsNames))
			}
		}

		// Sort and execute
		sortedExecutors := g.TopologicalSort()
		for i := range sortedExecutors {
			executor, ok := (*sortedExecutors[i].Value).(execute.Executor)
			if !ok {
				return util.NewInternalError("Failed to convert node to executor")
			}
			if err = executor.Execute(); err != nil {
				return err
			}
		}
	}

	return nil
}

func makeEdges(g *graph.Graph, node graph.Node, nodeMap, agentNodeMap map[string]graph.Node, agentByName, agentByUUID map[string]*client.AgentInfo, dependencies []string) (err error) {
	for _, dep := range dependencies {
		dependsOnNode, found := nodeMap[dep]
		if !found {
			// This means agent is not getting deployed with this file, so it must already exist on Controller
			agent, found := agentByName[dep]
			if !found {
				return util.NewNotFoundError(fmt.Sprintf("Could not find agent %s while establishing agent dependency graph\n", dep))
			}
			dependsOnNode, found = agentNodeMap[dep]
			if !found {
				// Create empty executor
				dependsOnNode = g.MakeNode()
				emptyExecutor := execute.NewEmptyExecutor(dep)
				*dependsOnNode.Value = emptyExecutor
				// Add to agentNodeMap to avoid duplicating nodes
				agentNodeMap[dep] = dependsOnNode
			}
			if agent != nil {
				// Fill dependency graph with agents on Controller
				uuidDependencies := getDependencies(agent.UpstreamRouters, agent.NetworkRouter)
				if err = makeEdges(g, dependsOnNode, nodeMap, agentNodeMap, agentByName, agentByUUID, mapUUIDsToNames(uuidDependencies, agentByUUID)); err != nil {
					return err
				}
			}
		}
		// Edge from x -> y means that x needs to complete before y
		g.MakeEdge(dependsOnNode, node)
	}
	return nil
}

func getDependencies(upstreamRouters *[]string, networkRouter *string) []string {
	dependencies := []string{}
	if upstreamRouters != nil {
		dependencies = append(dependencies, *upstreamRouters...)
	}
	if networkRouter != nil {
		dependencies = append(dependencies, *networkRouter)
	}
	return dependencies
}

func mapUUIDsToNames(uuids []string, agentByUUID map[string]*client.AgentInfo) (names []string) {
	for _, uuid := range uuids {
		agent, found := agentByUUID[uuid]
		var name string
		if found {
			name = agent.Name
		} else {
			name = uuid
		}
		names = append(names, name)
	}
	return
}
