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

package edgeresource

import (
	"fmt"

	clientutil "github.com/eclipse-iofog/iofogctl/v2/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

func Execute(namespace, name, newName string) error {
	if err := util.IsLowerAlphanumeric("Edge Resource", newName); err != nil {
		return err
	}

	util.SpinStart(fmt.Sprintf("Renaming edgeResource %s", name))

	// Init remote resources
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	// Check capability
	if err := clientutil.IsEdgeResourceCapable(namespace); err != nil {
		return err
	}

	// List all edge resources
	listResponse, err := clt.ListEdgeResources()
	if err != nil {
		return err
	}
	// Validate exists
	if len(listResponse.EdgeResources) == 0 {
		return util.NewNotFoundError(fmt.Sprintf("%s does not exist", name))
	}

	// Get full resource contents and update
	for _, meta := range listResponse.EdgeResources {
		if meta.Name != name {
			continue
		}
		// Get versioned resource
		oldEdge, err := clt.GetHttpEdgeResourceByName(meta.Name, meta.Version)
		if err != nil {
			return err
		}
		// Update versioned resource
		oldEdge.Name = newName
		if err := clt.UpdateHttpEdgeResource(name, oldEdge); err != nil {
			return err
		}
	}

	return nil
}
