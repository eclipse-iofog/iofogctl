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

package updatecatalogitem

import (
	deploy "github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(opt deploy.CatalogItem, namespace string) error {
	// Get Control Plane
	controlPlane, err := config.GetControlPlane(namespace)
	if err != nil || len(controlPlane.Controllers) == 0 {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return err
	}

	// Validate catalog item definition
	if err = validate(opt); err != nil {
		return err
	}

	// Init remote resources
	clt, err := client.NewAndLogin(controlPlane.Controllers[0].Endpoint, controlPlane.IofogUser.Email, controlPlane.IofogUser.Password)
	if err != nil {
		return err
	}

	currentItem, err := clt.GetCatalogItemByName(opt.Name)
	if err != nil {
		return err
	}

	request := client.CatalogItemUpdateRequest{
		ID:          currentItem.ID,
		Name:        opt.Name,
		Images:      []client.CatalogImage{},
		Description: opt.Description,
	}

	if opt.Registry != "" {
		request.RegistryID = client.RegistryTypeRegistryTypeIDDict[opt.Registry]
	}

	if opt.X86 != "" {
		request.Images = append(request.Images, client.CatalogImage{
			ContainerImage: opt.X86,
			AgentTypeID:    client.AgentTypeAgentTypeIDDict["x86"],
		})
	}

	if opt.ARM != "" {
		request.Images = append(request.Images, client.CatalogImage{
			ContainerImage: opt.ARM,
			AgentTypeID:    client.AgentTypeAgentTypeIDDict["arm"],
		})
	}

	if _, err = clt.UpdateCatalogItem(&request); err != nil {
		return err
	}

	return nil
}

func validate(opt deploy.CatalogItem) error {
	if opt.Name == "" {
		return util.NewInputError("Name must be specified")
	}

	if len(opt.Registry) > 0 && opt.Registry != "remote" && opt.Registry != "local" {
		return util.NewInputError("Registry must be either 'remote' or 'local'")
	}

	return nil
}
