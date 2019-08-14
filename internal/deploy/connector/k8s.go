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

func newKubernetesExecutor(namespace string, cnct *config.Connector, controlPlane config.ControlPlane) *kubernetesExecutor {
	k := &kubernetesExecutor{}
	k.namespace = namespace
	k.cnct = cnct
	k.controlPlane = controlPlane
	return k
}

func (exe *kubernetesExecutor) GetName() string {
	return exe.cnct.Name
}

func (exe *kubernetesExecutor) Execute() (err error) {
	defer util.SpinStop()
	util.SpinStart("Deploying Connector " + exe.cnct.Name)

	// Get Kubernetes installer
	installer, err := install.NewKubernetes(exe.cnct.KubeConfig, exe.namespace)
	if err != nil {
		return
	}

	// Configure deploy
	if err = installer.SetImages(map[string]string{"connector": exe.cnct.Image}); err != nil {
		return err
	}

	// Create connector on cluster
	if err = installer.CreateConnector(install.IofogUser(exe.controlPlane.IofogUser)); err != nil {
		return
	}

	// Update connector (its a pointer, this is returned to caller)
	if exe.cnct.Endpoint, err = installer.GetConnectorEndpoint(); err != nil {
		return
	}

	return
}
