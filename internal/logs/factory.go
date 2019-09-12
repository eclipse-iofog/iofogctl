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

package logs

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func NewExecutor(resourceType, namespace, name string) (execute.Executor, error) {
	switch resourceType {
	case "controller":
		return newControllerExecutor(namespace, name), nil
	case "agent":
		return newAgentExecutor(namespace, name), nil
	case "microservice":
		controlPlane, err := config.GetControlPlane(namespace)
		if err != nil {
			return nil, util.NewError("Could not get Control Plane for namespace " + namespace)
		}
		if len(controlPlane.Controllers) == 0 {
			return nil, util.NewError("No Controllers found in namespace " + namespace)
		}
		if util.IsLocalHost(controlPlane.Controllers[0].Host) {
			return nil, util.NewInputError("Microservice logs for local deploys is not supported. Use docker logs directly instead.")
		}
		return newRemoteMicroserviceExecutor(namespace, name), nil
	default:
		msg := "Unknown resource: '" + resourceType + "'"
		return nil, util.NewInputError(msg)
	}
}
