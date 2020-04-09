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

package deployagentconfig

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/eclipse-iofog/iofogctl/v2/internal"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"gopkg.in/yaml.v2"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type AgentConfigExecutor interface {
	execute.Executor
	GetAgentUUID() string
	SetHost(string)
	GetConfiguration() rsc.AgentConfiguration
	GetNamespace() string
}

type remoteExecutor struct {
	name        string
	uuid        string
	agentConfig rsc.AgentConfiguration
	namespace   string
}

func NewRemoteExecutor(name string, config rsc.AgentConfiguration, namespace string) *remoteExecutor {
	return &remoteExecutor{
		name:        name,
		agentConfig: config,
		namespace:   namespace,
	}
}

func (exe *remoteExecutor) GetNamespace() string {
	return exe.namespace
}

func (exe *remoteExecutor) GetConfiguration() rsc.AgentConfiguration {
	return exe.agentConfig
}

func (exe *remoteExecutor) SetHost(host string) {
	exe.agentConfig.Host = &host
}

func (exe *remoteExecutor) GetAgentUUID() string {
	return exe.uuid
}

func (exe *remoteExecutor) GetName() string {
	return exe.name
}

func isOverridingSystemAgent(controllerHost, agentHost string, isSystem bool) (err error) {
	// Generate controller endpoint
	controllerURL, err := url.Parse(controllerHost)
	if err != nil || controllerURL.Host == "" {
		controllerURL, err = url.Parse("//" + controllerHost) // Try to see if controllerEndpoint is an IP, in which case it needs to be pefixed by //
		if err != nil {
			return err
		}
	}
	agentURL, err := url.Parse(agentHost)
	if err != nil || agentURL.Host == "" {
		agentURL, err = url.Parse("//" + agentHost) // Try to see if controllerEndpoint is an IP, in which case it needs to be pefixed by //
		if err != nil {
			return err
		}
	}
	if agentURL.Hostname() == controllerURL.Hostname() && !isSystem {
		return util.NewConflictError("Cannot deploy an agent on the same host than the Controller\n")
	}
	return nil
}

func (exe *remoteExecutor) Execute() error {
	isSystem := internal.IsSystemAgent(exe.agentConfig)
	if !isSystem || install.IsVerbose() {
		util.SpinStart(fmt.Sprintf("Deploying agent %s configuration", exe.GetName()))
	}

	// Check controller is reachable
	clt, err := internal.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Check we are not about to override Vanilla system agent
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil || len(controlPlane.GetControllers()) == 0 {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return err
	}
	host := ""
	if exe.agentConfig.Host != nil {
		host = *exe.agentConfig.Host
	}
	endpoint, err := controlPlane.GetEndpoint()
	if err != nil {
		return err
	}
	if err := isOverridingSystemAgent(endpoint, host, isSystem); err != nil {
		return err
	}

	// Get the Agent in question
	agent, err := clt.GetAgentByName(exe.name, isSystem)
	// TODO: replace this check with built-in IsNewNotFound() func from go-sdk
	if err != nil && !strings.Contains(err.Error(), "not find agent") {
		return err
	}
	ip := ""
	if agent != nil {
		ip = agent.IPAddressExternal
	}
	// Get all other non-system Agents
	agentList, err := clt.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return err
	}
	// Process needs to be done at execute time because agent might have been created during deploy
	exe.agentConfig, err = Process(exe.agentConfig, exe.name, ip, agentList.Agents)
	if err != nil {
		return err
	}

	// Create if Agent does not exist
	if agent == nil {
		uuid, err := createAgentFromConfiguration(exe.agentConfig, exe.name, clt)
		exe.uuid = uuid
		return err
	}
	// Update existing Agent
	exe.uuid = agent.UUID
	return updateAgentConfiguration(&exe.agentConfig, agent.UUID, clt)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Unmarshal file
	agentConfig := rsc.AgentConfiguration{}
	if err = yaml.UnmarshalStrict(opt.Yaml, &agentConfig); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	if len(agentConfig.Name) == 0 {
		agentConfig.Name = opt.Name
	}

	if err = Validate(agentConfig); err != nil {
		return
	}

	return &remoteExecutor{
		name:        opt.Name,
		agentConfig: agentConfig,
		namespace:   opt.Namespace,
	}, nil
}

func findAgent(agents []client.AgentInfo, name string) *client.AgentInfo {
	for idx := range agents {
		if agents[idx].Name == name {
			return &agents[idx]
		}
	}
	return nil
}
