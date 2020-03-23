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
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type facadeExecutor struct {
	exe        execute.Executor
	controller rsc.Controller
	namespace  string
}

func (facade facadeExecutor) Execute() (err error) {
	util.SpinStart(fmt.Sprintf("Deploying controller %s", facade.GetName()))
	if err = facade.exe.Execute(); err != nil {
		return
	}
	install.Verbose(fmt.Sprintf("Controller is running at %s", facade.controller.GetEndpoint))
	if err = config.UpdateController(facade.namespace, facade.controller); err != nil {
		return
	}
	return config.Flush()
}

func (facade facadeExecutor) GetName() string {
	return facade.exe.GetName()
}

func newFacadeExecutor(exe execute.Executor, namespace string, controller rsc.Controller) execute.Executor {
	return facadeExecutor{
		exe:        exe,
		namespace:  namespace,
		controller: controller,
	}
}

func newExecutor(namespace string, baseController rsc.Controller) (execute.Executor, error) {
	if err := util.IsLowerAlphanumeric(baseController.GetName()); err != nil {
		return nil, err
	}

	baseControlPlane, err := config.GetControlPlane(namespace)
	if err != nil {
		return nil, err
	}
	user := baseControlPlane.GetUser()
	if user.Email == "" || user.Password == "" {
		return nil, util.NewError("Cannot deploy Controller because ioFog user is not specified")
	}
	switch controlPlane := baseControlPlane.(type) {
	case *rsc.KubernetesControlPlane:
		return newFacadeExecutor(newKubernetesExecutor(namespace, controlPlane), namespace, nil), nil
	case *rsc.RemoteControlPlane:
		remoteController, ok := baseController.(*rsc.RemoteController)
		if !ok {
			return nil, util.NewInputError("Tried to deploy wrong type of Controller in a Remote Control Plane")
		}
		return newFacadeExecutor(newRemoteExecutor(namespace, remoteController, controlPlane), namespace, remoteController), nil
	case *rsc.LocalControlPlane:
		// Check the namespace does not contain a Controller yet
		cli, err := install.NewLocalContainerClient()
		if err != nil {
			return nil, err
		}
		localController, ok := baseController.(*rsc.LocalController)
		if !ok {
			return nil, util.NewInputError("Tried to deploy wrong type of Controller in a Remote Control Plane")
		}
		exe, err := newLocalExecutor(namespace, localController, controlPlane, cli)
		if err != nil {
			return nil, err
		}
		return newFacadeExecutor(exe, namespace, localController), nil
	}

	return nil, util.NewError("Could not determine Control Plane type")
}
