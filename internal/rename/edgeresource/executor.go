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

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

func Execute(namespace, name, newName string) error {
	if err := util.IsLowerAlphanumeric("Edge Resource", newName); err != nil {
		return err
	}

	util.SpinStart(fmt.Sprintf("Renaming edgeResource %s", name))

	// Init remote resources
	clt, err := iutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	// Check capability
	if err := iutil.IsEdgeResourceCapable(namespace); err != nil {
		return err
	}

	// List all edge resources
	listResponse, err := clt.ListEdgeResources()
	if err != nil {
		return err
	}
	// Record the resources we want to rename
	renamedResources := []client.EdgeResourceMetadata{}
	for _, edge := range listResponse.EdgeResources {
		if edge.Name == name {
			edge.Name = newName
			renamedResources = append(renamedResources, edge)
		}
	}
	// Validate exists
	if len(renamedResources) == 0 {
		return util.NewNotFoundError(fmt.Sprintf("%s does not exist", name))
	}

	// Update all versions
	for _, edge := range renamedResources {
		if err := clt.UpdateHttpEdgeResource(name, edge); err != nil {
			return err
		}
	}

	return nil
}
