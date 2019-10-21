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

package deleteconnector

import (
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal"
)

func deleteConnectorFromController(namespace, connectorIP string) error {
	ctrlClient, err := internal.NewControllerClient(namespace)
	if err != nil {
		return err
	}
	if err = ctrlClient.DeleteConnector(connectorIP); err != nil {
		if !strings.Contains(err.Error(), "NotFoundError") {
			return err
		}
	}

	return nil
}
