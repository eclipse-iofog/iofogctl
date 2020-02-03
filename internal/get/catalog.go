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

package get

import (
	"strconv"

	apps "github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type catalogExecutor struct {
	namespace string
}

func newCatalogExecutor(namespace string) *catalogExecutor {
	a := &catalogExecutor{}
	a.namespace = namespace
	return a
}

func (exe *catalogExecutor) Execute() error {
	printNamespace(exe.namespace)
	if err := generateCatalogOutput(exe.namespace); err != nil {
		return err
	}
	return nil
}

func (exe *catalogExecutor) GetName() string {
	return ""
}

func generateCatalogOutput(namespace string) error {
	// Get Config
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}

	var items []apps.CatalogItem

	// Connect to Controller if it is ready
	endpoint, err := ns.ControlPlane.GetControllerEndpoint()
	if err == nil {
		// Instantiate client
		// Log into Controller
		ctrlClient, err := client.NewAndLogin(endpoint, ns.ControlPlane.IofogUser.Email, ns.ControlPlane.IofogUser.Password)
		if err != nil {
			return tabulateCatalogItems(items)
		}

		// Get catalog from Controller
		listCatalogResponse, err := ctrlClient.GetCatalog()
		if err != nil {
			return err
		}
		for _, item := range listCatalogResponse.CatalogItems {
			catalogItem := apps.CatalogItem{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				Registry:    client.RegistryTypeIDRegistryTypeDict[item.RegistryID],
			}
			for _, image := range item.Images {
				switch client.AgentTypeIDAgentTypeDict[image.AgentTypeID] {
				case "x86":
					catalogItem.X86 = image.ContainerImage
					break
				case "arm":
					catalogItem.ARM = image.ContainerImage
					break
				default:
					break
				}
			}
			items = append(items, catalogItem)
		}
	}

	return tabulateCatalogItems(items)
}

func tabulateCatalogItems(catalogItems []apps.CatalogItem) error {
	// Generate table and headers
	table := make([][]string, len(catalogItems)+1)
	headers := []string{
		"ID",
		"NAME",
		"DESCRIPTION",
		"REGISTRY",
		"X86",
		"ARM",
	}
	table[0] = append(table[0], headers...)
	// Populate rows
	idx := 0
	for _, item := range catalogItems {
		row := []string{
			strconv.Itoa(item.ID),
			item.Name,
			item.Description,
			item.Registry,
			item.X86,
			item.ARM,
		}
		table[idx+1] = append(table[idx+1], row...)
		idx = idx + 1
	}

	// Print table
	return print(table)
}
