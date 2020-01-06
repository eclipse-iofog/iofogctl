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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func (exe *executor) localRemove() error {
	containerName := install.GetLocalContainerName("connector")
	// Get IP
	client, err := install.NewLocalContainerClient()
	if err != nil {
		return err
	}
	_, err = client.GetContainerIP(containerName)
	if err != nil {
		return err
	}

	// Clean container
	if errClean := client.CleanContainer(containerName); errClean != nil {
		util.PrintNotify(fmt.Sprintf("Could not clean Connector container: %v", errClean))
	}

	return nil
}
