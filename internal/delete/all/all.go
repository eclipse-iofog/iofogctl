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
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deleteagent "github.com/eclipse-iofog/iofogctl/internal/delete/agent"
	deleteconnector "github.com/eclipse-iofog/iofogctl/internal/delete/connector"
	deletecontroller "github.com/eclipse-iofog/iofogctl/internal/delete/controller"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace string) error {
	// Make sure to update config despite failure
	defer config.Flush()

	// Get namespace
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}

	// Delete Agents
	if len(ns.Agents) > 0 {
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
	}

	// Delete routes (which would prevent connector from being deleted)
	if len(ns.ControlPlane.Controllers) > 0 {
		// Get list of microservices from backend
		endpoint := ns.ControlPlane.Controllers[0].Endpoint
		clt, err := client.NewAndLogin(endpoint, ns.ControlPlane.IofogUser.Email, ns.ControlPlane.IofogUser.Password)
		if err != nil {
			return err
		}
		msvcs, err := clt.GetAllMicroservices()
		if err != nil {
			return err
		}
		// Delete routes
		if len(msvcs.Microservices) > 0 {
			util.SpinStart("Deleting Routes")

			for _, msvc := range msvcs.Microservices {
				for _, destUUID := range msvc.Routes {
					if err = clt.DeleteMicroserviceRoute(msvc.UUID, destUUID); err != nil {
						return err
					}
				}
			}
		}
	}

	// Delete Connectors
	if len(ns.Connectors) > 0 {
		util.SpinStart("Deleting Connectors")

		var executors []execute.Executor
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
	}

	// Delete Controllers
	if len(ns.ControlPlane.Controllers) > 0 {
		util.SpinStart("Deleting Controllers")

		var executors []execute.Executor
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
	}

	// Delete Control Plane
	if err = config.DeleteControlPlane(namespace); err != nil {
		return err
	}

	return nil
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
