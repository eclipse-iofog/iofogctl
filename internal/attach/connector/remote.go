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
)

func (exe executor) remoteAttach() (err error) {

	ctrlClient, err := internal.NewControllerClient(exe.opt.Namespace)
	if err != nil {
		return err
	}

	if err = ctrlClient.AddConnector(client.ConnectorInfo{
		IP:      exe.cnct.Host,
		Domain:  exe.cnct.Host,
		Name:    exe.cnct.Name,
		DevMode: true,
	}); err != nil {
		return
	}
	return
}
