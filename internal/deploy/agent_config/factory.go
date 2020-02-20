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

package deployagentconfig

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal"
	"gopkg.in/yaml.v2"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
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
	GetConfiguration() config.AgentConfiguration
	GetNamespace() string
}

type remoteExecutor struct {
	name        string
	uuid        string
	agentConfig config.AgentConfiguration
	namespace   string
}

func NewRemoteExecutor(name string, config config.AgentConfiguration, namespace string) *remoteExecutor {
	return &remoteExecutor{
		name:        name,
		agentConfig: config,
		namespace:   namespace,
	}
}

func (exe *remoteExecutor) GetNamespace() string {
	return exe.namespace
}

func (exe *remoteExecutor) GetConfiguration() config.AgentConfiguration {
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

func (exe *remoteExecutor) Execute() error {
	fmt.Printf("Attaching agent config: %v\n", exe.agentConfig)
	fmt.Printf("Host: %s\n", *exe.agentConfig.Host)

	isSystem := internal.IsSystemAgent(exe.agentConfig)
	if !isSystem || install.IsVerbose() {
		util.SpinStart(fmt.Sprintf("Deploying agent %s configuration", exe.GetName()))
	}

	// Check controller is reachable
	clt, err := internal.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	// Process needs to be done at execute time because agent might have been created during deploy
	exe.agentConfig, err = Process(exe.agentConfig, exe.name, clt)
	if err != nil {
		return err
	}

	agent, err := clt.GetAgentByName(exe.name)
	if err != nil {
		if strings.Contains(err.Error(), "Could not find agent") {
			uuid, err := install.CreateAgentFromConfiguration(exe.agentConfig, exe.name, clt)
			exe.uuid = uuid
			return err
		}
		return err
	}
	exe.uuid = agent.UUID
	return install.UpdateAgentConfiguration(&exe.agentConfig, agent.UUID, clt)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Unmarshal file
	agentConfig := config.AgentConfiguration{}
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
