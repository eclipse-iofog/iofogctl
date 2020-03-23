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

package deploycontroller

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type facadeExecutor struct {
	exe        execute.Executor
	controller *rsc.Controller
	namespace  string
}

func (facade facadeExecutor) Execute() (err error) {
	util.SpinStart(fmt.Sprintf("Deploying controller %s", facade.GetName()))
	if err = facade.exe.Execute(); err != nil {
		return
	}
	install.Verbose(fmt.Sprintf("Controller is running at %s", facade.controller.Endpoint))
	if err = config.UpdateController(facade.namespace, *facade.controller); err != nil {
		return
	}
	return config.Flush()
}

func (facade facadeExecutor) GetName() string {
	return facade.exe.GetName()
}

func newFacadeExecutor(exe execute.Executor, namespace string, controller *rsc.Controller) execute.Executor {
	return facadeExecutor{
		exe:        exe,
		namespace:  namespace,
		controller: controller,
	}
}

func newExecutor(namespace string, ctrl *rsc.Controller, controlPlane rsc.ControlPlane) (execute.Executor, error) {
	if err := util.IsLowerAlphanumeric(ctrl.Name); err != nil {
		return nil, err
	}

	if controlPlane.IofogUser.Email == "" || controlPlane.IofogUser.Password == "" {
		return nil, util.NewError("Cannot deploy Controller because ioFog user is not specified")
	}
	// Local executor
	if util.IsLocalHost(ctrl.Host) {
		// Check the namespace does not contain a Controller yet
		nbControllers := len(controlPlane.Controllers)
		if nbControllers != 1 {
			return nil, util.NewInputError("Cannot deploy more than a single Controller locally")
		}
		cli, err := install.NewLocalContainerClient()
		if err != nil {
			return nil, err
		}
		exe, err := newLocalExecutor(namespace, ctrl, controlPlane, cli)
		if err != nil {
			return nil, err
		}
		return newFacadeExecutor(exe, namespace, ctrl), nil
	}

	// Kubernetes executor
	if controlPlane.Kube.Config != "" {
		return newFacadeExecutor(newKubernetesExecutor(namespace, ctrl, &controlPlane), namespace, ctrl), nil
	}

	// Default executor
	if ctrl.Host == "" || ctrl.SSH.KeyFile == "" || ctrl.SSH.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}
	return newFacadeExecutor(newRemoteExecutor(namespace, ctrl, controlPlane), namespace, ctrl), nil
}
