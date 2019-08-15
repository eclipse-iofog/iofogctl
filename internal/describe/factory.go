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

package describe

import (
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func NewExecutor(resourceType, namespace, name, filename string) (execute.Executor, error) {
	switch resourceType {
	case "namespace":
		return newNamespaceExecutor(namespace, filename), nil
	case "controlplane":
		return newControlPlaneExecutor(namespace, filename), nil
	case "controller":
		return newControllerExecutor(namespace, name, filename), nil
	case "connector":
		return newConnectorExecutor(namespace, name, filename), nil
	case "agent":
		return newAgentExecutor(namespace, name, filename), nil
	case "microservice":
		return newMicroserviceExecutor(namespace, name, filename), nil
	case "application":
		return newApplicationExecutor(namespace, name, filename), nil
	default:
		msg := "Unknown resourceType: '" + resourceType + "'"
		return nil, util.NewInputError(msg)
	}
}
