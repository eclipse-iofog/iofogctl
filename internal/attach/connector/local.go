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

func (exe executor) localAttach() (err error) {

	// Try to add connector to Controller database
	ctrlClient, err := internal.NewControllerClient(exe.opt.Namespace)
	if err != nil {
		return err
	}

	containerName := install.GetLocalContainerName("connector")
	containerClient, err := install.NewLocalContainerClient()
	if err != nil {
		return err
	}
	containerIP, err := containerClient.GetContainerIP(containerName)
	if err != nil {
		return err
	}

	if err := ctrlClient.AddConnector(client.ConnectorInfo{
		IP:      containerIP,
		Domain:  containerIP,
		Name:    exe.cnct.Name,
		DevMode: true,
	}); err != nil {
		return err
	}
	return
}
