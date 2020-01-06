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
	deletecontrolplane "github.com/eclipse-iofog/iofogctl/internal/delete/controlplane"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace string, useDetached, soft bool) error {
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
			exe, err := deleteagent.NewExecutor(namespace, agent.Name, useDetached, soft)
			if err != nil {
				return err
			}
			executors = append(executors, exe)
		}
		if err := runExecutors(executors); err != nil {
			return err
		}
	}

	if !useDetached {
		// Delete routes (which would prevent connector from being deleted)
		if len(ns.ControlPlane.Controllers) > 0 {
			// Get list of microservices from backend
			endpoint, _ := ns.ControlPlane.GetControllerEndpoint()
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
	}

	// Delete Connectors
	if len(ns.Connectors) > 0 {
		util.SpinStart("Deleting Connectors")

		var executors []execute.Executor
		for _, cnct := range ns.Connectors {
			exe, err := deleteconnector.NewExecutor(namespace, cnct.Name, useDetached, soft)
			if err != nil {
				return err
			}
			// TODO: Replace serial execution when CRD updated
			// Kubernetes Connectors cannot run in parallel
			if cnct.Kube.Config != "" {
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

	if !useDetached {
		// Delete Controllers
		util.SpinStart("Deleting ControlPlane")
		exe, err := deletecontrolplane.NewExecutor(namespace, "controlplane", soft)
		if err != nil {
			return err
		}
		if err = exe.Execute(); err != nil {
			return err
		}
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
