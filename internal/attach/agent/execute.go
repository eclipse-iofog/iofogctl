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

package attachagent

import (
	"errors"
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"

	deploy "github.com/eclipse-iofog/iofogctl/v2/internal/deploy/agent"
	deployagentconfig "github.com/eclipse-iofog/iofogctl/v2/internal/deploy/agentconfig"
)

type Options struct {
	Name        string
	Namespace   string
	Host        string
	User        string
	Port        int
	KeyFile     string
	UseDetached bool
}

type executor struct {
	opt Options
}

func NewExecutor(opt Options) execute.Executor {
	return executor{opt: opt}
}

func (exe executor) GetName() string {
	return exe.opt.Name
}

func (exe executor) onFailure(inErr error) error {
	// Get a client
	iofogclient, err := iutil.NewControllerClient(exe.opt.Namespace)
	if err != nil {
		return errors.New(fmt.Sprintf("%s\nFailed to create Controller API client: %s", inErr.Error(), err.Error()))
	}
	agent, err := iofogclient.GetAgentByName(exe.opt.Name, false)
	if err != nil {
		return errors.New(fmt.Sprintf("%s\nFailed to get newly created Agent by name: %s", inErr.Error(), err.Error()))
	}
	if err := iofogclient.DeleteAgent(agent.UUID); err != nil {
		return errors.New(fmt.Sprintf("%s\nFailed to delete newly created Agent: %s", inErr.Error(), err.Error()))
	}
	return inErr
}

func (exe executor) Execute() error {
	util.SpinStart("Attaching Agent")

	// Update local cache based on Controller
	if err := iutil.UpdateAgentCache(exe.opt.Namespace); err != nil {
		return err
	}

	baseAgent, err := config.GetDetachedAgent(exe.opt.Name)
	if err != nil {
		errStr := fmt.Sprintf("%s\nIs Agent %s detached? Use `iofogctl get agents --detached` to check.", err.Error(), exe.opt.Name)
		return errors.New(errStr)
	}

	// Create Agent
	host := baseAgent.GetHost()
	configExecutor := deployagentconfig.NewRemoteExecutor(exe.opt.Name,
		rsc.AgentConfiguration{
			Name: exe.opt.Name,
			AgentConfiguration: client.AgentConfiguration{
				Host: &host,
			},
		}, exe.opt.Namespace, nil)
	if err = configExecutor.Execute(); err != nil {
		return err
	}

	// Create executor to provision the Agent w/ Controller
	var executor execute.Executor
	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		executor, err = deploy.NewLocalExecutor(exe.opt.Namespace, agent, false)
		if err != nil {
			return exe.onFailure(err)
		}
	case *rsc.RemoteAgent:
		executor, err = deploy.NewRemoteExecutor(exe.opt.Namespace, agent, false)
		if err != nil {
			return exe.onFailure(err)
		}
	}
	deployExecutor, ok := executor.(execute.ProvisioningExecutor)
	if !ok {
		return exe.onFailure(errors.New("Attach: Could not convert executor"))
	}
	UUID, err := deployExecutor.ProvisionAgent()
	if err != nil {
		return exe.onFailure(err)
	}

	// Update config
	baseAgent.SetUUID(UUID)
	if baseAgent.GetCreatedTime() == "" {
		baseAgent.SetCreatedTime(util.NowUTC())
	}
	if err = config.AttachAgent(exe.opt.Namespace, exe.opt.Name, UUID); err != nil {
		return exe.onFailure(err)
	}

	return config.Flush()
}
