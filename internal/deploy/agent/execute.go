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

package deployagent

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
)

type Options struct {
	Namespace string
	Name      string
	Yaml      []byte
	IsSystem  bool
	Tags      *[]string
}

func NewRemoteExecutorYAML(opt Options) (exe execute.Executor, err error) {
	// Read the input file
	agent, err := rsc.UnmarshallRemoteAgent(opt.Yaml)
	if err != nil {
		return exe, err
	}

	if len(opt.Name) > 0 {
		agent.Name = opt.Name
	}

	// Validate
	if err = ValidateRemoteAgent(&agent); err != nil {
		return
	}

	remoteExe := newRemoteExecutor(opt.Namespace, &agent)
	return newFacadeExecutor(remoteExe, opt.Namespace, &agent, opt.IsSystem, opt.Tags), nil
}

func NewLocalExecutorYAML(opt Options) (exe execute.Executor, err error) {
	// Read the input file
	agent, err := rsc.UnmarshallLocalAgent(opt.Yaml)
	if err != nil {
		return exe, err
	}

	if len(opt.Name) > 0 {
		agent.Name = opt.Name
	}

	localExe, err := newLocalExecutor(opt.Namespace, &agent, opt.IsSystem)
	if err != nil {
		return nil, err
	}
	return newFacadeExecutor(localExe, opt.Namespace, &agent, opt.IsSystem, opt.Tags), nil
}
