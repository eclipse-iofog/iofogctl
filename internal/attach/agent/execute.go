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

func (exe executor) Execute() (err error) {
	util.SpinStart("Attaching Agent")

	ns, err := config.GetNamespace(exe.opt.Namespace)
	if err != nil {
		return
	}

	// Update local cache based on Controller
	if err = iutil.UpdateAgentCache(exe.opt.Namespace); err != nil {
		return
	}

	var baseAgent rsc.Agent
	if exe.opt.UseDetached {
		baseAgent, err = config.GetDetachedAgent(exe.opt.Name)
		if err != nil {
			return
		}
	} else {
		// Check Agent does not exist
		if _, err := ns.GetAgent(exe.opt.Name); err == nil {
			return util.NewConflictError(exe.opt.Namespace + "/" + exe.opt.Name)
		}
		// Determine type of ECN
		controlPlane, err := ns.GetControlPlane()
		if err != nil {
			return err
		}
		switch controlPlane.(type) {
		case *rsc.LocalControlPlane:
			baseAgent = &rsc.LocalAgent{
				Name: exe.opt.Name,
				Host: exe.opt.Host,
			}
		default:
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
	if baseAgent == nil {
		return util.NewInternalError("Failed to convert options to Agent")
	}

	// Create Agent
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
		return
	}

	var executor execute.Executor
	switch agent := baseAgent.(type) {
	case *rsc.LocalAgent:
		executor, err = deploy.NewLocalExecutor(exe.opt.Namespace, agent, false)
		if err != nil {
			return
		}
	case *rsc.RemoteAgent:
		executor, err = deploy.NewRemoteExecutor(exe.opt.Namespace, agent, false)
		if err != nil {
			return
		}
	}
	deployExecutor, ok := executor.(execute.ProvisioningExecutor)
	if !ok {
		return util.NewInternalError("Attach: Could not convert executor")
	}

	UUID, err := deployExecutor.ProvisionAgent()
	if err != nil {
		return
	}

	baseAgent.SetUUID(UUID)
	if baseAgent.GetCreatedTime() == "" {
		baseAgent.SetCreatedTime(util.NowUTC())
	}

	if exe.opt.UseDetached {
		if err = config.AttachAgent(exe.opt.Namespace, exe.opt.Name, UUID); err != nil {
			return
		}
	} else {
		if err = ns.UpdateAgent(baseAgent); err != nil {
			return
		}
	}

	return config.Flush()
}
