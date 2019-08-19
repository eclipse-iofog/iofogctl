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

package deletecatalogitem

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace, name string) error { // Get Control Plane
	controlPlane, err := config.GetControlPlane(namespace)
	if err != nil || len(controlPlane.Controllers) == 0 {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return err
	}

	util.SpinStart("Deleting Catalog item")
	// Init remote resources
	clt, err := client.NewAndLogin(controlPlane.Controllers[0].Endpoint, controlPlane.IofogUser.Email, controlPlane.IofogUser.Password)
	if err != nil {
		return err
	}

	// Get all catalog items
	catalog, err := clt.GetCatalog()
	if err != nil {
		return err
	}

	// Find catalog item
	for _, item := range catalog.CatalogItems {
		if item.Name == name {
			return clt.DeleteCatalogItem(item.ID)
		}
	}
	return util.NewNotFoundError("Could not find the catalog item to delete")
}
