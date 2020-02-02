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
	GetConfiguration() config.AgentConfiguration
}

type remoteExecutor struct {
	name        string
	agentConfig config.AgentConfiguration
	namespace   string
}

func (exe remoteExecutor) GetConfiguration() config.AgentConfiguration {
	return exe.agentConfig
}

func (exe remoteExecutor) GetName() string {
	return exe.name
}

func (exe remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying agent %s configuration", exe.GetName()))

	// Check controller is reachable
	clt, err := internal.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	agent, err := clt.GetAgentByName(exe.name)
	if err != nil {
		return err
	}

	// Process needs to be done at execute time because agent might have been created during deploy
	exe.agentConfig, err = ProcessAgentNames(exe.agentConfig, clt)
	if err != nil {
		return err
	}

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

	return remoteExecutor{
		name:        opt.Name,
		agentConfig: agentConfig,
		namespace:   opt.Namespace,
	}, nil
}
