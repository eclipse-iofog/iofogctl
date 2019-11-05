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

package deployconnector

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type facadeExecutor struct {
	exe       execute.Executor
	connector *config.Connector
	namespace string
}

func (facade facadeExecutor) Execute() (err error) {
	// Get the Control Plane
	controlPlane, err := config.GetControlPlane(facade.namespace)
	if err != nil {
		return
	}

	// Must contain Controller
	if len(controlPlane.Controllers) == 0 {
		return util.NewError("There are no Controllers in this namespace. You must first deploy one or more Controllers.")
	}

	// Must contain an ioFog User
	if controlPlane.IofogUser.Email == "" || controlPlane.IofogUser.Password == "" {
		return util.NewError("The Control Plane in this namespace does not have a valid ioFog user")
	}
	util.SpinStart(fmt.Sprintf("Deploying connector %s", facade.GetName()))
	if err = facade.exe.Execute(); err != nil {
		return
	}
	if err = config.UpdateConnector(facade.namespace, *facade.connector); err != nil {
		return
	}
	return config.Flush()
}

func (facade facadeExecutor) GetName() string {
	return facade.exe.GetName()
}

func newFacadeExecutor(exe execute.Executor, namespace string, connector *config.Connector) execute.Executor {
	return facadeExecutor{
		exe:       exe,
		namespace: namespace,
		connector: connector,
	}
}

func newExecutor(namespace string, cnct *config.Connector) (execute.Executor, error) {
	if err := util.IsLowerAlphanumeric(cnct.Name); err != nil {
		return nil, err
	}

	// Local executor
	if util.IsLocalHost(cnct.SSH.Host) {
		cli, err := install.NewLocalContainerClient()
		if err != nil {
			return nil, err
		}
		exe, err := newLocalExecutor(namespace, cnct, cli)
		if err != nil {
			return exe, err
		}
		return newFacadeExecutor(exe, namespace, cnct), nil
	}

	if cnct.Kube.Config != "" {
		return newFacadeExecutor(newKubernetesExecutor(namespace, cnct), namespace, cnct), nil
	}

	// Default executor
	if cnct.SSH.Host == "" || cnct.SSH.KeyFile == "" || cnct.SSH.User == "" {
		return nil, util.NewInputError("Must specify user, host, and key file flags for remote deployment")
	}

	return newFacadeExecutor(newRemoteExecutor(namespace, cnct), namespace, cnct), nil
}
