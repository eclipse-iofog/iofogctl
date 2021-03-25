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

package configure

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
)

type multipleExecutor struct {
	opt *Options
}

func newMultipleExecutor(opt *Options) *multipleExecutor {
	return &multipleExecutor{
		opt: opt,
	}
}

func (exe *multipleExecutor) Execute() (err error) {
	// Instantiate executor list
	var executors []execute.Executor

	// Populate list
	if exe.opt.ResourceType == "agents" {
		executors, err = exe.AddAgentExecutors(executors)
		if err != nil {
			return err
		}
	}
	if exe.opt.ResourceType == "controllers" {
		executors, err = exe.AddControllerExecutors(executors)
		if err != nil {
			return err
		}
	}

	// Execute
	for _, executor := range executors {
		if err := executor.Execute(); err != nil {
			return err
		}
	}

	return nil
}

func (exe *multipleExecutor) AddAgentExecutors(executors []execute.Executor) ([]execute.Executor, error) {
	ns, err := config.GetNamespace(exe.opt.Namespace)
	if err != nil {
		return nil, err
	}
	var agents []rsc.Agent
	if exe.opt.UseDetached {
		agents = config.GetDetachedAgents()
	} else {
		agents = ns.GetAgents()
	}
	for _, agent := range agents {
		opt := exe.opt
		opt.Name = agent.GetName()
		executors = append(executors, newAgentExecutor(opt))
	}

	return executors, nil
}

func (exe *multipleExecutor) AddControllerExecutors(executors []execute.Executor) ([]execute.Executor, error) {
	ns, err := config.GetNamespace(exe.opt.Namespace)
	if err != nil {
		return nil, err
	}
	for _, controller := range ns.GetControllers() {
		opt := exe.opt
		opt.Name = controller.GetName()
		executors = append(executors, newControllerExecutor(opt))
	}

	return executors, nil
}

func (exe *multipleExecutor) GetName() string {
	return exe.opt.Name
}
