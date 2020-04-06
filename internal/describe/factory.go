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
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

func NewExecutor(resourceType, namespace, name, filename string, useDetached bool) (execute.Executor, error) {
	switch resourceType {
	case "namespace":
		return newNamespaceExecutor(namespace, filename), nil
	case "controlplane":
		return newControlPlaneExecutor(namespace, filename), nil
	case "controller":
		return newControllerExecutor(namespace, name, filename), nil
	case "agent":
		return newAgentExecutor(namespace, name, filename, useDetached), nil
	case "registry":
		return newRegistryExecutor(namespace, name, filename, useDetached)
	case "volume":
		return newVolumeExecutor(namespace, name, filename), nil
	case "agent-config":
		return newAgentConfigExecutor(namespace, name, filename), nil
	case "microservice":
		return newMicroserviceExecutor(namespace, name, filename), nil
	case "application":
		return newApplicationExecutor(namespace, name, filename), nil
	default:
		msg := "Unknown resourceType: '" + resourceType + "'"
		return nil, util.NewInputError(msg)
	}
}
