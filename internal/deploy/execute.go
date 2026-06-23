package deploy

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployagent "github.com/eclipse-iofog/iofogctl/internal/deploy/agent"
	deployagentconfig "github.com/eclipse-iofog/iofogctl/internal/deploy/agentconfig"
	deployapplication "github.com/eclipse-iofog/iofogctl/internal/deploy/application"
	deployapplicationtemplate "github.com/eclipse-iofog/iofogctl/internal/deploy/applicationtemplate"
	deploycatalogitem "github.com/eclipse-iofog/iofogctl/internal/deploy/catalogitem"
	deploycertificate "github.com/eclipse-iofog/iofogctl/internal/deploy/certificate"
	deployconfigmap "github.com/eclipse-iofog/iofogctl/internal/deploy/configmap"
	deploylocalcontroller "github.com/eclipse-iofog/iofogctl/internal/deploy/controller/local"
	deployremotecontroller "github.com/eclipse-iofog/iofogctl/internal/deploy/controller/remote"
	deployk8scontrolplane "github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane/k8s"
	deploylocalcontrolplane "github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane/local"
	deployremotecontrolplane "github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane/remote"
	deployedgeresource "github.com/eclipse-iofog/iofogctl/internal/deploy/edgeresource"
	deploymicroservice "github.com/eclipse-iofog/iofogctl/internal/deploy/microservice"
	deploynatsaccountrule "github.com/eclipse-iofog/iofogctl/internal/deploy/natsaccountrule"
	deploynatsuserrule "github.com/eclipse-iofog/iofogctl/internal/deploy/natsuserrule"
	deployofflineimage "github.com/eclipse-iofog/iofogctl/internal/deploy/offlineimage"
	deployregistry "github.com/eclipse-iofog/iofogctl/internal/deploy/registry"
	deployrole "github.com/eclipse-iofog/iofogctl/internal/deploy/role"
	deployrolebinding "github.com/eclipse-iofog/iofogctl/internal/deploy/rolebinding"
	deploysecret "github.com/eclipse-iofog/iofogctl/internal/deploy/secret"
	deployservice "github.com/eclipse-iofog/iofogctl/internal/deploy/service"
	deployserviceaccount "github.com/eclipse-iofog/iofogctl/internal/deploy/serviceaccount"
	deployvalidate "github.com/eclipse-iofog/iofogctl/internal/deploy/validate"
	deployvolume "github.com/eclipse-iofog/iofogctl/internal/deploy/volume"
	deployvolumemount "github.com/eclipse-iofog/iofogctl/internal/deploy/volumeMount"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/twmb/algoimpl/go/graph"
)

var kindOrder = []config.Kind{
	config.SecretKind,
	config.CertificateAuthorityKind,
	config.CertificateKind,
	config.ConfigMapKind,
	config.RoleKind,
	config.RoleBindingKind,
	config.ServiceAccountKind,
	config.NatsAccountRuleKind,
	config.NatsUserRuleKind,
	config.RemoteAgentKind,
	config.LocalAgentKind,
	config.EdgeResourceKind,
	config.ApplicationTemplateKind,
	config.VolumeKind,
	config.OfflineImageKind,
	config.VolumeMountKind,
	config.RegistryKind,
	config.CatalogItemKind,
	config.ApplicationKind,
	config.MicroserviceKind,
	config.ServiceKind,
}

type Options struct {
	Namespace    string
	InputFile    string
	NoCache      bool
	TransferPool int
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
	return deploylocalcontrolplane.NewExecutor(deploylocalcontrolplane.Options{Namespace: opt.Namespace, Yaml: opt.YAML, FullYAML: opt.FullYAML, Name: opt.Name})
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

func deploySecret(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploysecret.NewExecutor(deploysecret.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Data: opt.Data, Name: opt.Name})
}

func deployConfigMap(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployconfigmap.NewExecutor(deployconfigmap.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Data: opt.Data, Name: opt.Name})
}

func deployService(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployservice.NewExecutor(deployservice.Options{Namespace: opt.Namespace, Tags: opt.Tags, Yaml: opt.YAML, Name: opt.Name})
}

func deployVolumeMount(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployvolumemount.NewExecutor(deployvolumemount.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployCertificate(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploycertificate.NewExecutor(deploycertificate.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name, Kind: opt.Kind})
}

func deployRole(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployrole.NewExecutor(deployrole.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployRoleBinding(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployrolebinding.NewExecutor(deployrolebinding.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployServiceAccount(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployserviceaccount.NewExecutor(deployserviceaccount.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployNatsAccountRule(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploynatsaccountrule.NewExecutor(deploynatsaccountrule.Options{
		Namespace: opt.Namespace,
		Yaml:      opt.YAML,
		FullYAML:  opt.FullYAML,
		Name:      opt.Name,
	})
}

func deployNatsUserRule(opt *execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploynatsuserrule.NewExecutor(deploynatsuserrule.Options{
		Namespace: opt.Namespace,
		Yaml:      opt.YAML,
		FullYAML:  opt.FullYAML,
		Name:      opt.Name,
	})
}

// Execute deploy from yaml file
func Execute(opt *Options) (err error) {
	kindHandlers := buildKindHandlers(opt.NoCache, opt.TransferPool)
	executorsMap, err := execute.GetExecutorsFromYAML(opt.InputFile, opt.Namespace, kindHandlers, false)
	if err != nil {
		return err
	}

	// Create any AgentConfig executor missing.
	// Each Agent requires a corresponding Agent Config to be created with Controller.
	// CP system agents are registered by LocalControlPlane deploy (systemAgent path), not here.
	appendedAgentExecs := append(executorsMap[config.LocalAgentKind], executorsMap[config.RemoteAgentKind]...)
	for _, agentGenericExecutor := range appendedAgentExecs {
		agentExecutor, ok := agentGenericExecutor.(deployagent.AgentDeployExecutor)
		if !ok {
			return util.NewInternalError("Could not convert agent deploy executor\n")
		}

		specHost := agentExecutor.GetHost()
		tags := agentExecutor.GetTags()
		deployConfig := agentExecutor.GetConfig()

		apiHost := specHost
		if deployConfig != nil && deployConfig.Host != nil && strings.TrimSpace(*deployConfig.Host) != "" {
			apiHost = strings.TrimSpace(*deployConfig.Host)
		}

		found := false
		for _, configGenericExecutor := range executorsMap[config.AgentConfigKind] {
			configExecutor, ok := configGenericExecutor.(deployagentconfig.AgentConfigExecutor)
			if !ok {
				return util.NewInternalError("Could not convert agent config executor\n")
			}
			if agentExecutor.GetName() == configExecutor.GetName() {
				found = true
				configExecutor.SetHost(apiHost)
				configExecutor.SetTags(tags)
				break
			}
		}
		if found {
			continue
		}

		agentConfig := buildSyntheticAgentConfiguration(deployConfig, apiHost)
		syntheticConfig := &rsc.AgentConfiguration{
			Name:               agentExecutor.GetName(),
			AgentConfiguration: agentConfig,
		}
		if err := deployagentconfig.Validate(syntheticConfig); err != nil {
			return err
		}

		executorsMap[config.AgentConfigKind] = append(executorsMap[config.AgentConfigKind],
			deployagentconfig.NewRemoteExecutor(
				agentExecutor.GetName(),
				syntheticConfig,
				opt.Namespace,
				tags,
			))
	}

	if localCPExes, exists := executorsMap[config.LocalControlPlaneKind]; exists {
		namespaceHasOtherCP := false
		if ns, nsErr := config.GetNamespace(opt.Namespace); nsErr == nil {
			if existingCP, cpErr := ns.GetControlPlane(); cpErr == nil {
				if _, ok := existingCP.(*rsc.LocalControlPlane); !ok {
					namespaceHasOtherCP = true
				}
			}
		}
		if err := deployvalidate.LocalControlPlaneDeploy(len(localCPExes), namespaceHasOtherCP); err != nil {
			return err
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
	if err := deployvalidate.RemoteControllerDeploy(opt.InputFile); err != nil {
		return err
	}
	if errs := execute.RunExecutors(executorsMap[config.RemoteControllerKind], "deploy remote controller"); len(errs) > 0 {
		return execute.CoalesceErrors(errs)
	}
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

func buildKindHandlers(noCache bool, transferPool int) map[config.Kind]func(*execute.KindHandlerOpt) (execute.Executor, error) {
	handlers := map[config.Kind]func(*execute.KindHandlerOpt) (execute.Executor, error){
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
		config.SecretKind:                 deploySecret,
		config.ConfigMapKind:              deployConfigMap,
		config.RoleKind:                   deployRole,
		config.RoleBindingKind:            deployRoleBinding,
		config.ServiceAccountKind:         deployServiceAccount,
		config.NatsAccountRuleKind:        deployNatsAccountRule,
		config.NatsUserRuleKind:           deployNatsUserRule,
		config.ServiceKind:                deployService,
		config.VolumeMountKind:            deployVolumeMount,
		config.CertificateAuthorityKind:   deployCertificate,
		config.CertificateKind:            deployCertificate,
	}
	handlers[config.OfflineImageKind] = func(opt *execute.KindHandlerOpt) (execute.Executor, error) {
		return deployofflineimage.NewExecutor(deployofflineimage.Options{
			Namespace: opt.Namespace,
			Yaml:      opt.YAML,
			Name:      opt.Name,
			NoCache:   noCache,
			PoolSize:  transferPool,
		})
	}
	return handlers
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
	// List agents on Controller with auth retry
	var listAgentReponse client.ListAgentsResponse
	err := clientutil.ExecuteWithAuthRetry(namespace, func(ctrlClient *client.Client) error {
		var err error
		listAgentReponse, err = ctrlClient.ListAgents(client.ListAgentsRequest{})
		return err
	})
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

	// Add default nats server
	agentByName[iofog.VanillaNatsAgentName] = &client.AgentInfo{Name: iofog.VanillaNatsAgentName}

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
		dependencies := getDependencies(configuration.UpstreamRouters, configuration.NetworkRouter, configuration.UpstreamNatsServers)
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
				uuidDependencies := getDependencies(agent.UpstreamRouters, agent.NetworkRouter, agent.UpstreamNatsServers)
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

func getDependencies(upstreamRouters *[]string, networkRouter *string, upstreamNatsServers *[]string) []string {
	dependencies := []string{}
	if upstreamRouters != nil {
		dependencies = append(dependencies, *upstreamRouters...)
	}
	if networkRouter != nil {
		dependencies = append(dependencies, *networkRouter)
	}
	if upstreamNatsServers != nil {
		dependencies = append(dependencies, *upstreamNatsServers...)
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

func buildSyntheticAgentConfiguration(deployConfig *rsc.AgentConfiguration, apiHost string) client.AgentConfiguration {
	host := apiHost
	isSystem := false

	if deployConfig == nil {
		return client.AgentConfiguration{
			Host:     &host,
			IsSystem: &isSystem,
		}
	}

	cfg := deployConfig.AgentConfiguration
	cfg.Host = &host
	cfg.IsSystem = &isSystem
	return cfg
}
