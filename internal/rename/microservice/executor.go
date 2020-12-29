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

package microservice

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/v2/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

func Execute(namespace, name, newName string) error {
	if err := util.IsLowerAlphanumeric("Microservice", newName); err != nil {
		return err
	}
	// Init remote resources
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	msvc, err := clt.GetMicroserviceByName(name)
	if err != nil {
		return err
	}

	util.SpinStart(fmt.Sprintf("Renaming microservice %s", name))

	if _, err = clt.UpdateMicroservice(&client.MicroserviceUpdateRequest{
		UUID: msvc.UUID,
		Name: &newName,
		// Bug in Controller, fails if empty because images should be an array
		Images: msvc.Images,
		// Ports and Routes get automatically updated by the SDK, to avoid deletion of port mapping or route, those fields are mandatory
		Ports: msvc.Ports,
	}); err != nil {
		return err
	}

	return err
}
