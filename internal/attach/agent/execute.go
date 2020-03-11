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

package attachagent

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"

	deploy "github.com/eclipse-iofog/iofogctl/v2/internal/deploy/agent"
	deployagentconfig "github.com/eclipse-iofog/iofogctl/v2/internal/deploy/agent_config"
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

func NewExecutor(opt Options) (execute.Executor, error) {
	return executor{opt: opt}, nil
}

func (exe executor) GetName() string {
	return exe.opt.Name
}

func (exe executor) Execute() error {
	util.SpinStart("Attaching Agent")

	var agent config.Agent
	var err error
	if exe.opt.UseDetached {
		agent, err = config.GetDetachedAgent(exe.opt.Name)
	} else {
		agent = config.Agent{
			Name: exe.opt.Name,
			Host: exe.opt.Host,
			SSH: config.SSH{
				User:    exe.opt.User,
				KeyFile: exe.opt.KeyFile,
				Port:    exe.opt.Port,
			},
		}
	}

	if err != nil {
		return err
	}

	// Create fog
	configExecutor := deployagentconfig.NewRemoteExecutor(
		exe.opt.Name,
		config.AgentConfiguration{
			Name: exe.opt.Name,
			AgentConfiguration: client.AgentConfiguration{
				Host: &agent.Host,
			},
		}, exe.opt.Namespace)
	if err = configExecutor.Execute(); err != nil {
		return err
	}

	executor, err := deploy.NewDeployExecutor(exe.opt.Namespace, &agent, false)
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

	agent.UUID = UUID
	if agent.Created == "" {
		agent.Created = util.NowUTC()
	}

	if exe.opt.UseDetached {
		if err = config.AttachAgent(exe.opt.Namespace, exe.opt.Name, UUID); err != nil {
			return err
		}
	} else {
		if err = config.UpdateAgent(exe.opt.Namespace, agent); err != nil {
			return err
		}
	}

	return config.Flush()
}
