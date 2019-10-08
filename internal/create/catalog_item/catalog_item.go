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

package createcatalogitem

import (
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deploy "github.com/eclipse-iofog/iofogctl/pkg/iofog/deploy"
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

	if _, err = clt.CreateCatalogItem(&client.CatalogItemCreateRequest{
		Name: opt.Name,
		Images: []client.CatalogImage{
			{ContainerImage: opt.X86, AgentTypeID: client.AgentTypeAgentTypeIDDict["x86"]},
			{ContainerImage: opt.ARM, AgentTypeID: client.AgentTypeAgentTypeIDDict["arm"]},
		},
		RegistryID:  client.RegistryTypeRegistryTypeIDDict[opt.Registry],
		Description: opt.Description,
	}); err != nil {
		return err
	}

	return nil
}

func validate(opt deploy.CatalogItem) error {
	if opt.Name == "" {
		return util.NewInputError("Name must be specified")
	}

	if opt.ARM == "" && opt.X86 == "" {
		return util.NewInputError("At least one image must be specified")
	}

	if opt.Registry != "remote" && opt.Registry != "local" {
		return util.NewInputError("Registry must be either 'remote' or 'local'")
	}

	return nil
}
