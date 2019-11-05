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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type kubernetesExecutor struct {
	namespace    string
	cnct         *config.Connector
	controlPlane config.ControlPlane
}

func newKubernetesExecutor(namespace string, cnct *config.Connector) *kubernetesExecutor {
	k := &kubernetesExecutor{}
	k.namespace = namespace
	k.cnct = cnct
	return k
}

func (exe *kubernetesExecutor) GetName() string {
	return exe.cnct.Name
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Get Control Plane
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil || len(controlPlane.Controllers) == 0 {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Connector")
		return err
	}
	exe.controlPlane = controlPlane

	// Get Kubernetes installer
	installer, err := install.NewKubernetes(exe.cnct.Kube.Config, exe.namespace)
	if err != nil {
		return
	}

	// Configure deploy
	installer.SetConnectorImage(exe.cnct.Container.Image)

	// Create connector on cluster
	if err = installer.CreateConnector(exe.cnct.Name, install.IofogUser(exe.controlPlane.IofogUser)); err != nil {
		return
	}

	// Update connector (its a pointer, this is returned to caller)
	endpoint, err := installer.GetConnectorEndpoint(exe.cnct.Name)
	if err != nil {
		return
	}
	exe.cnct.Endpoint = endpoint
	exe.cnct.Created = util.NowUTC()

	return
}
