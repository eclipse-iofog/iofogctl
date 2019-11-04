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

package configure

import (
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	ResourceType string
	Namespace    string
	Name         string
	KubeConfig   string
	KeyFile      string
	User         string
	Port         int
}

func NewExecutor(opt Options) (execute.Executor, error) {
	switch opt.ResourceType {
	case "controller":
		return newControllerExecutor(opt), nil
	case "connector":
		return newConnectorExecutor(opt), nil
	case "agent":
		return newAgentExecutor(opt), nil
	default:
		msg := "Unsupported resource type: '" + opt.ResourceType + "'"
		return nil, util.NewInputError(msg)
	}
}
