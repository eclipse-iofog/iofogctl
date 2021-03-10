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

package describe

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type Options struct {
	Resource   string
	Name       string
	Namespace  string
	Filename   string
	IsDetached bool
	Version    string
}

func NewExecutor(opt *Options) (execute.Executor, error) {
	switch opt.Resource {
	case "namespace":
		return newNamespaceExecutor(opt.Namespace, opt.Filename), nil
	case "controlplane":
		return newControlPlaneExecutor(opt.Namespace, opt.Filename), nil
	case "controller":
		return newControllerExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "agent":
		return newAgentExecutor(opt.Namespace, opt.Name, opt.Filename, opt.IsDetached), nil
	case "registry":
		return newRegistryExecutor(opt.Namespace, opt.Name, opt.Filename)
	case "agent-config":
		return newAgentConfigExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "microservice":
		return newMicroserviceExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "application-template":
		return newApplicationTemplateExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "application":
		return newApplicationExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "volume":
		return newVolumeExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "route":
		return newRouteExecutor(opt.Namespace, opt.Name, opt.Filename), nil
	case "edge-resource":
		return newEdgeResourceExecutor(opt.Namespace, opt.Name, opt.Version, opt.Filename), nil
	default:
		return nil, util.NewInputError(fmt.Sprintf("Unknown resources: %s", opt.Resource))
	}
}
