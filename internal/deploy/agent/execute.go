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

package deployagent

import (
	"github.com/eclipse-iofog/iofogctl/internal/execute"
)

type Options struct {
	Namespace string
	Name      string
	Yaml      []byte
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Read the input file
	agent, err := UnmarshallYAML(opt.Yaml)
	if err != nil {
		return exe, err
	}

	if len(opt.Name) > 0 {
		agent.Name = opt.Name
	}

	// Validate
	if err = Validate(agent); err != nil {
		return
	}

	return newExecutor(opt.Namespace, &agent)
}
