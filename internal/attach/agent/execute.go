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

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"

	deploy "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/agent"
	deployagentconfig "github.com/eclipse-iofog/iofogctl/v3/internal/deploy/agentconfig"
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
	opt *Options
}

func NewExecutor(opt *Options) execute.Executor {
	return &executor{opt: opt}
}

func (exe *executor) GetName() string {
	return exe.opt.Name
}

func (exe *executor) fail(inErr error) error {
	// Get a client
	iofogclient, err := clientutil.NewControllerClient(exe.opt.Namespace)
	if err != nil {
		return fmt.Errorf("%s\nFailed to create Controller API client: %s", inErr.Error(), err.Error())
	}
	agent, err := iofogclient.GetAgentByName(exe.opt.Name, false)
	if err != nil {
		msg := "%s\nFailed to get newly created Agent by name: %s"
		return fmt.Errorf(msg, inErr.Error(), err.Error())
	}
	if err := iofogclient.DeleteAgent(agent.UUID); err != nil {
		msg := "%s\nFailed to delete newly created Agent: %s"
		return fmt.Errorf(msg, inErr.Error(), err.Error())
	}
	return inErr
}

func (exe *executor) Execute() error {
	util.SpinStart("Attaching Agent")

	// Update local cache based on Controller
	if err := clientutil.SyncAgentInfo(exe.opt.Namespace); err != nil {
		return err
	}

	baseAgent, err := config.GetDetachedAgent(exe.opt.Name)
	if err != nil {
		errStr := fmt.Sprintf("%s\nIs Agent %s detached? Use `iofogctl get agents --detached` to check.", err.Error(), exe.opt.Name)
		return errors.New(errStr)
	}

	// Create Agent Config
	agentConfig := baseAgent.GetConfig()
	if agentConfig == nil {
		agentConfig = &rsc.AgentConfiguration{}
	}
	host := baseAgent.GetHost()
	agentConfig.Host = &host
	configExecutor := deployagentconfig.NewRemoteExecutor(
		exe.opt.Name,
		agentConfig,
		exe.opt.Namespace, nil)
	if err := configExecutor.Execute(); err != nil {
		return err
	}

	// Create Agent
	var executor execute.Executor
	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		executor, err = deploy.NewLocalExecutor(exe.opt.Namespace, agent, false)
		if err != nil {
			return exe.fail(err)
		}
	case *rsc.RemoteAgent:
		executor, err = deploy.NewRemoteExecutor(exe.opt.Namespace, agent, false)
		if err != nil {
			return exe.fail(err)
		}
	}
	deployExecutor, ok := executor.(execute.ProvisioningExecutor)
	if !ok {
		msg := "attach: Could not convert Executor"
		return exe.fail(errors.New(msg))
	}
	UUID, err := deployExecutor.ProvisionAgent()
	if err != nil {
		return exe.fail(err)
	}
	// TODO: Remove this additional config deploy step when Agent no longer posts config on provision
	// Deploy config again
	if err := configExecutor.Execute(); err != nil {
		return err
	}

	// Update local config
	baseAgent.SetUUID(UUID)
	if baseAgent.GetCreatedTime() == "" {
		baseAgent.SetCreatedTime(util.NowUTC())
	}
	if err = config.AttachAgent(exe.opt.Namespace, exe.opt.Name, UUID); err != nil {
		return exe.fail(err)
	}

	return config.Flush()
}
