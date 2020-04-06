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
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
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

func (exe executor) Execute() error {
	util.SpinStart("Attaching Agent")

	var baseAgent rsc.Agent
	var err error
	if exe.opt.UseDetached {
		baseAgent, err = config.GetDetachedAgent(exe.opt.Name)
	} else {
		switch baseAgent.(type) {
		case *rsc.LocalAgent:
			baseAgent = &rsc.LocalAgent{
				Name: exe.opt.Name,
				Host: exe.opt.Host,
			}
		case *rsc.RemoteAgent:
			baseAgent = &rsc.RemoteAgent{
				Name: exe.opt.Name,
				Host: exe.opt.Host,
				SSH: rsc.SSH{
					User:    exe.opt.User,
					KeyFile: exe.opt.KeyFile,
					Port:    exe.opt.Port,
				},
			}
		}
	}

	if err != nil {
		return err
	}

	// Create fog
	host := baseAgent.GetHost()
	configExecutor := deployagentconfig.NewRemoteExecutor(
		exe.opt.Name,
		rsc.AgentConfiguration{
			Name: exe.opt.Name,
			AgentConfiguration: client.AgentConfiguration{
				Host: &host,
			},
		}, exe.opt.Namespace)
	if err = configExecutor.Execute(); err != nil {
		return err
	}

	var executor execute.Executor
	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		executor, err = deploy.NewLocalExecutor(exe.opt.Namespace, agent, false)
	case *rsc.RemoteAgent:
		executor, err = deploy.NewRemoteExecutor(exe.opt.Namespace, agent, false)
	}
	if err != nil {
		return err
	}
	deployExecutor, ok := executor.(execute.ProvisioningExecutor)
	if !ok {
		return util.NewInternalError("Attach: Could not convert executor")
	}

	UUID, err := deployExecutor.ProvisionAgent()
	if err != nil {
		return err
	}

	baseAgent.SetUUID(UUID)
	if baseAgent.GetCreatedTime() == "" {
		baseAgent.SetCreatedTime(util.NowUTC())
	}

	if exe.opt.UseDetached {
		if err = config.AttachAgent(exe.opt.Namespace, exe.opt.Name, UUID); err != nil {
			return err
		}
	} else {
		if err = config.UpdateAgent(exe.opt.Namespace, baseAgent); err != nil {
			return err
		}
	}

	return config.Flush()
}
