/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package deletevolume

import (
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

func deleteLocal(agent *rsc.LocalAgent, volume rsc.Volume) error {
	client, err := install.NewLocalContainerClient()
	if err != nil {
		return err
	}

	// Delete
	if _, err := client.ExecuteCmd(install.GetLocalContainerName("agent", false), []string{"sh", "-c", "rm -rf " + util.AddTrailingSlash(volume.Destination) + "*"}); err != nil {
		return err
	}
	return nil
}
