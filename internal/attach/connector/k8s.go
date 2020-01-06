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

package attachconnector

import (
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

func (exe executor) k8sAttach() (err error) {
	// Get Kubernetes installer
	installer, err := install.NewKubernetes(exe.opt.KubeConfig, exe.opt.Namespace)
	if err != nil {
		return
	}

	// Update connector
	exe.cnct.Endpoint, err = installer.GetConnectorEndpoint(exe.opt.Name)
	if err != nil {
		return
	}

	ctrlClient, err := internal.NewControllerClient(exe.opt.Namespace)
	if err != nil {
		return err
	}

	if err := ctrlClient.AddConnector(client.ConnectorInfo{
		IP:      exe.cnct.Endpoint,
		Domain:  exe.cnct.Endpoint,
		Name:    exe.cnct.Name,
		DevMode: true,
	}); err != nil {
		return err
	}
	return
}
