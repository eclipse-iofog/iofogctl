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

package deleteall

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deleteagent "github.com/eclipse-iofog/iofogctl/internal/delete/agent"
	deleteconnector "github.com/eclipse-iofog/iofogctl/internal/delete/connector"
	deletecontroller "github.com/eclipse-iofog/iofogctl/internal/delete/controller"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace string) error {
	// Get namespace
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}

	// Delete Agents
	util.SpinStart("Deleting Agents")
	var executors []execute.Executor
	for _, agent := range ns.Agents {
		exe, err := deleteagent.NewExecutor(namespace, agent.Name)
		if err != nil {
			return err
		}
		executors = append(executors, exe)
	}
	if err := runExecutors(executors); err != nil {
		return err
	}
	for _, agent := range ns.Agents {
		if err = config.DeleteAgent(namespace, agent.Name); err != nil {
			return err
		}
	}

	// Delete routes (which would prevent connector from being deleted)
	util.SpinStart("Deleting Routes")
	executors = executors[:0]
	for _, ctrl := range ns.ControlPlane.Controllers {
		clt, err := client.NewAndLogin(ctrl.Endpoint, ns.ControlPlane.IofogUser.Email, ns.ControlPlane.IofogUser.Password)
		if err != nil {
			return err
		}
		msvcs, err := clt.GetAllMicroservices()
		if err != nil {
			return err
		}
		for _, msvc := range msvcs.Microservices {
			for _, destUUID := range msvc.Routes {
				if err = clt.DeleteMicroserviceRoute(msvc.UUID, destUUID); err != nil {
					return err
				}
			}
		}
	}

	// Delete Connectors
	util.SpinStart("Deleting Connectors")
	executors = executors[:0]
	for _, cnct := range ns.Connectors {
		exe, err := deleteconnector.NewExecutor(namespace, cnct.Name)
		if err != nil {
			return err
		}
		// TODO: Replace serial execution when CRD updated
		// Kubernetes Connectors cannot run in parallel
		if cnct.KubeConfig != "" {
			if err = exe.Execute(); err != nil {
				return err
			}
		} else {
			executors = append(executors, exe)
		}
	}
	if err := runExecutors(executors); err != nil {
		return err
	}
	for _, cnct := range ns.Connectors {
		if err = config.DeleteConnector(namespace, cnct.Name); err != nil {
			return err
		}
	}

	// Delete Controllers
	util.SpinStart("Deleting Controllers")
	executors = executors[:0]
	for _, ctrl := range ns.ControlPlane.Controllers {
		exe, err := deletecontroller.NewExecutor(namespace, ctrl.Name)
		if err != nil {
			return err
		}
		executors = append(executors, exe)
	}
	if err := runExecutors(executors); err != nil {
		return err
	}
	for _, ctrl := range ns.ControlPlane.Controllers {
		if err = config.DeleteController(namespace, ctrl.Name); err != nil {
			return err
		}
	}

	// Delete Control Plane
	if err = config.DeleteControlPlane(namespace); err != nil {
		return err
	}

	return config.Flush()
}

func runExecutors(executors []execute.Executor) error {
	if errs, failedExes := execute.ForParallel(executors); len(errs) > 0 {
		for idx := range errs {
			util.PrintNotify("Error from " + failedExes[idx].GetName() + ": " + errs[idx].Error())
		}
		return util.NewError("Failed to delete")
	}
	return nil
}
