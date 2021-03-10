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

package logs

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

func NewExecutor(resourceType, namespace, name string) (execute.Executor, error) {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}
	switch resourceType {
	case "controller":
		baseControlPlane, err := ns.GetControlPlane()
		if err != nil {
			return nil, util.NewError("Could not get Control Plane for namespace " + namespace)
		}
		switch controlPlane := baseControlPlane.(type) {
		case *rsc.KubernetesControlPlane:
			return newKubernetesControllerExecutor(controlPlane, namespace, name), nil
		case *rsc.RemoteControlPlane:
			return newRemoteControllerExecutor(controlPlane, namespace, name), nil
		case *rsc.LocalControlPlane:
			return newLocalControllerExecutor(controlPlane, namespace, name), nil
		}
	case "agent":
		return newAgentExecutor(namespace, name), nil
	case "microservice":
		if len(ns.GetControllers()) == 0 {
			return nil, util.NewError("No Controllers found in namespace " + namespace)
		}
		return newRemoteMicroserviceExecutor(namespace, name), nil
	}
	msg := "Unknown resource: '" + resourceType + "'"
	return nil, util.NewInputError(msg)
}
