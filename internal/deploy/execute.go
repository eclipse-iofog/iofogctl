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

package deploy

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	deployagent "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/agent"
	deployagentconfig "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/agentconfig"
	deployapplication "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/application"
	deployapplicationtemplate "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/applicationtemplate"
	deploycatalogitem "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/catalogitem"
	deploylocalcontroller "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/controller/local"
	deployremotecontroller "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/controller/remote"
	deployk8scontrolplane "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/controlplane/k8s"
	deploylocalcontrolplane "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/controlplane/local"
	deployremotecontrolplane "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/controlplane/remote"
	deployedgeresource "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/edgeresource"
	deploymicroservice "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/microservice"
	deployregistry "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/registry"
	deployroute "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/route"
	deployvolume "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/volume"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/twmb/algoimpl/go/graph"
)

var kindOrder = []config.Kind{
	config.RemoteAgentKind,
	config.LocalAgentKind,
	config.EdgeResourceKind,
	config.ApplicationTemplateKind,
	config.VolumeKind,
	config.RegistryKind,
	config.CatalogItemKind,
	config.ApplicationKind,
	config.MicroserviceKind,
	config.RouteKind,
}

type Options struct {
	Namespace string
	InputFile string
}

func deployEdgeResource(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployedgeresource.NewExecutor(deployedgeresource.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployCatalogItem(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploycatalogitem.NewExecutor(deploycatalogitem.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployApplicationTemplate(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployapplicationtemplate.NewExecutor(deployapplicationtemplate.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployApplication(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployapplication.NewExecutor(deployapplication.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployMicroservice(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploymicroservice.NewExecutor(deploymicroservice.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployKubernetesControlPlane(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployk8scontrolplane.NewExecutor(deployk8scontrolplane.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployRemoteControlPlane(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployremotecontrolplane.NewExecutor(deployremotecontrolplane.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployLocalControlPlane(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploylocalcontrolplane.NewExecutor(deploylocalcontrolplane.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployRemoteController(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployremotecontroller.NewExecutor(deployremotecontroller.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployLocalController(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploylocalcontroller.NewExecutor(deploylocalcontroller.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployRemoteAgent(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployagent.NewRemoteExecutorYAML(deployagent.Options{Namespace: opt.Namespace, Tags: opt.Tags, Yaml: opt.YAML, Name: opt.Name})
}

func deployLocalAgent(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployagent.NewLocalExecutorYAML(deployagent.Options{Namespace: opt.Namespace, Tags: opt.Tags, Yaml: opt.YAML, Name: opt.Name})
}

func deployAgentConfig(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployagentconfig.NewExecutor(deployagentconfig.Options{Namespace: opt.Namespace, Tags: opt.Tags, Yaml: opt.YAML, Name: opt.Name})
}

func deployRegistry(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployregistry.NewExecutor(deployregistry.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployVolume(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployvolume.NewExecutor(deployvolume.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployRoute(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployroute.NewExecutor(deployroute.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

var kindHandlers = map[config.Kind]func(*execute.KindHandlerOpt) (execute.Executor, error){
	config.ApplicationKind:            deployApplication,
	config.ApplicationTemplateKind:    deployApplicationTemplate,
	config.MicroserviceKind:           deployMicroservice,
	config.CatalogItemKind:            deployCatalogItem,
	config.EdgeResourceKind:           deployEdgeResource,
	config.KubernetesControlPlaneKind: deployKubernetesControlPlane,
	config.RemoteControlPlaneKind:     deployRemoteControlPlane,
	config.LocalControlPlaneKind:      deployLocalControlPlane,
	config.RemoteControllerKind:       deployRemoteController,
	config.LocalControllerKind:        deployLocalController,
	config.RemoteAgentKind:            deployRemoteAgent,
	config.LocalAgentKind:             deployLocalAgent,
	config.AgentConfigKind:            deployAgentConfig,
	config.RegistryKind:               deployRegistry,
	config.VolumeKind:                 deployVolume,
	config.RouteKind:                  deployRoute,
}

// Execute deploy from yaml file
func Execute(opt *Options) (err error) {
	executorsMap, err := execute.GetExecutorsFromYAML(opt.InputFile, opt.Namespace, kindHandlers)
	if err != nil {
		return err
	}

	// Create any AgentConfig executor missing
	// Each Agent requires a corresponding Agent Config to be created with Controller
	appendedAgentExecs := append(executorsMap[config.LocalAgentKind], executorsMap[config.RemoteAgentKind]...)
	for _, agentGenericExecutor := range appendedAgentExecs {
		agentExecutor, ok := agentGenericExecutor.(deployagent.AgentDeployExecutor)
		if !ok {
			return util.NewInternalError("Could not convert agent deploy executor\n")
		}
		found := false
		host := agentExecutor.GetHost()
		tags := agentExecutor.GetTags()
		for _, configGenericExecutor := range executorsMap[config.AgentConfigKind] {
			configExecutor, ok := configGenericExecutor.(deployagentconfig.AgentConfigExecutor)
			if !ok {
				return util.NewInternalError("Could not convert agent config executor\n")
			}
			if agentExecutor.GetName() == configExecutor.GetName() {
				found = true
				configExecutor.SetHost(host)
				configExecutor.SetTags(tags)
				break
			}
		}
		if !found {
			agentConfig := client.AgentConfiguration{
				Host: &host,
			}
			if util.IsLocalHost(host) { // Set de default local config to interior standalone
				upstreamRouters := []string{}
				routerMode := "interior"
				edgeRouterPort := 56721
				interRouterPort := 56722
				agentConfig.UpstreamRouters = &upstreamRouters
				agentConfig.RouterConfig = client.RouterConfig{
					RouterMode:      &routerMode,
					EdgeRouterPort:  &edgeRouterPort,
					InterRouterPort: &interRouterPort,
				}
			}
			executorsMap[config.AgentConfigKind] = append(executorsMap[config.AgentConfigKind], deployagentconfig.NewRemoteExecutor(
				agentExecutor.GetName(),
				&rsc.AgentConfiguration{
					Name:               agentExecutor.GetName(),
					AgentConfiguration: agentConfig,
				},
				opt.Namespace,
				tags,
			))
		}
	}

	// ControlPlanes (should only be 1)
	cpCount := 0
	errMsg := "Specified multiple Control Planes in a single Namespace"
	if exe, exists := executorsMap[config.KubernetesControlPlaneKind]; exists {
		if errs := execute.RunExecutors(exe, "deploy Kubernetes Control Plane"); len(errs) > 0 {
			return execute.CoalesceErrors(errs)
		}
		cpCount++
	}
	if exe, exists := executorsMap[config.RemoteControlPlaneKind]; exists {
		if cpCount > 0 {
			err = util.NewInputError(errMsg)
		}
		if errs := execute.RunExecutors(exe, "deploy Remote Control Plane"); len(errs) > 0 {
			return execute.CoalesceErrors(errs)
		}
		cpCount++
	}
	if exe, exists := executorsMap[config.LocalControlPlaneKind]; exists {
		if cpCount > 0 {
			err = util.NewInputError(errMsg)
		}
		if errs := execute.RunExecutors(exe, "deploy Local Control Plane"); len(errs) > 0 {
			return execute.CoalesceErrors(errs)
		}
	}

	// Controllers
	if errs := execute.RunExecutors(executorsMap[config.LocalControllerKind], "deploy local controller"); len(errs) > 0 {
		return execute.CoalesceErrors(errs)
	}

	// Agent config
	if err := deployAgentConfiguration(executorsMap[config.AgentConfigKind]); err != nil {
		return err
	}

	// Execute in parallel by priority order
	// Edge Resources, Agents, Volumes, CatalogItem, Application, Microservice, Route
	for idx := range kindOrder {
		if errs := execute.RunExecutors(executorsMap[kindOrder[idx]], fmt.Sprintf("deploy %s", kindOrder[idx])); len(errs) > 0 {
			return execute.CoalesceErrors(errs)
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
		if err := sortAndExecute(namespace, executors); err != nil {
			return err
		}
	}

	return nil
}

func sortAndExecute(namespace string, executors []deployagentconfig.AgentConfigExecutor) error {
	// List agents on Controller
	ctrlClient, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	listAgentReponse, err := ctrlClient.ListAgents(client.ListAgentsRequest{})
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
	nodeMap := make(map[string]graph.Node)
	agentNodeMap := make(map[string]graph.Node)

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
		if err := makeEdges(g, node, nodeMap, agentNodeMap, agentByName, agentByUUID, dependencies); err != nil {
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
		if err := executor.Execute(); err != nil {
			return err
		}
	}
	return nil
}

// TODO: Refactor this func to have struct arg
func makeEdges(g *graph.Graph, node graph.Node, nodeMap, agentNodeMap map[string]graph.Node,
	agentByName, agentByUUID map[string]*client.AgentInfo, dependencies []string) (err error) {
	for _, dep := range dependencies {
		dependsOnNode, found := nodeMap[dep]
		if !found {
			// This means agent is not getting deployed with this file, so it must already exist on Controller
			agent, found := agentByName[dep]
			if !found {
				return util.NewNotFoundError(fmt.Sprintf("Could not find Agent %s while establishing agent dependency graph\n", dep))
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
				if err := makeEdges(g, dependsOnNode, nodeMap, agentNodeMap, agentByName, agentByUUID, mapUUIDsToNames(uuidDependencies, agentByUUID)); err != nil {
					return err
				}
			}
		}
		// Edge from x -> y means that x needs to complete before y
		if err := g.MakeEdge(dependsOnNode, node); err != nil {
			return err
		}
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
