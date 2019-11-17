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
	Host         string
	KeyFile      string
	User         string
	Port         int
}

var multipleResources = map[string]bool{
	"all":         true,
	"controllers": true,
	"connectors":  true,
	"agents":      true,
}

func NewExecutor(opt Options) (execute.Executor, error) {
	switch opt.ResourceType {
	case "default-namespace":
		return newDefaultNamespaceExecutor(opt), nil
	case "controller":
		return newControllerExecutor(opt), nil
	case "connector":
		return newConnectorExecutor(opt), nil
	case "agent":
		return newAgentExecutor(opt), nil
	default:
		if _, exists := multipleResources[opt.ResourceType]; !exists {
			return nil, util.NewInputError("Unsupported resource: " + opt.ResourceType)
		}
		return newMultipleExecutor(opt), nil
	}
}
