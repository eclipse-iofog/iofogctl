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

package deploycatalogitem

import (
	"fmt"
	"strconv"

	apps "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type remoteExecutor struct {
	catalogItem apps.CatalogItem
	namespace   string
}

func (exe *remoteExecutor) GetName() string {
	return exe.catalogItem.Name
}

func (exe *remoteExecutor) updateCatalogItem(clt *client.Client) (err error) {
	currentItem, err := clt.GetCatalogItem(exe.catalogItem.ID)
	if err != nil {
		return err
	}

	request := client.CatalogItemUpdateRequest{
		ID:          currentItem.ID,
		Name:        exe.catalogItem.Name,
		Images:      []client.CatalogImage{},
		Description: exe.catalogItem.Description,
	}

	if exe.catalogItem.Registry != "" {
		registryID, ok := client.RegistryTypeRegistryTypeIDDict[exe.catalogItem.Registry]
		if !ok {
			registryID, err = strconv.Atoi(exe.catalogItem.Registry)
			if err != nil {
				return err
			}
		}
		request.RegistryID = registryID
	}

	if exe.catalogItem.X86 != "" {
		request.Images = append(request.Images, client.CatalogImage{
			ContainerImage: exe.catalogItem.X86,
			AgentTypeID:    client.AgentTypeAgentTypeIDDict["x86"],
		})
	}

	if exe.catalogItem.ARM != "" {
		request.Images = append(request.Images, client.CatalogImage{
			ContainerImage: exe.catalogItem.ARM,
			AgentTypeID:    client.AgentTypeAgentTypeIDDict["arm"],
		})
	}

	if _, err = clt.UpdateCatalogItem(&request); err != nil {
		return err
	}

	return nil
}

func (exe *remoteExecutor) createCatalogItem(clt *client.Client) (err error) {
	if _, err = clt.CreateCatalogItem(&client.CatalogItemCreateRequest{
		Name: exe.catalogItem.Name,
		Images: []client.CatalogImage{
			{ContainerImage: exe.catalogItem.X86, AgentTypeID: client.AgentTypeAgentTypeIDDict["x86"]},
			{ContainerImage: exe.catalogItem.ARM, AgentTypeID: client.AgentTypeAgentTypeIDDict["arm"]},
		},
		RegistryID:  client.RegistryTypeRegistryTypeIDDict[exe.catalogItem.Registry],
		Description: exe.catalogItem.Description,
	}); err != nil {
		return err
	}
	return nil
}

func (exe *remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying catalog item %s", exe.GetName()))
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}
	if exe.catalogItem.ID == 0 {
		return exe.createCatalogItem(clt)
	}
	return exe.updateCatalogItem(clt)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	// Check the namespace exists
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return exe, err
	}

	// Check Controller exists
	if len(ns.GetControllers()) == 0 {
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Applications")
	}

	// Unmarshal file
	var catalogItem apps.CatalogItem
	if err = yaml.UnmarshalStrict(opt.Yaml, &catalogItem); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	if len(opt.Name) > 0 {
		catalogItem.Name = opt.Name
	}

	// Validate catalog item definition
	if err := validate(&catalogItem); err != nil {
		return nil, err
	}

	return &remoteExecutor{
		namespace:   opt.Namespace,
		catalogItem: catalogItem,
	}, nil
}

func validate(opt *apps.CatalogItem) error {
	if opt.Name == "" {
		return util.NewInputError("Name must be specified")
	}
	if err := util.IsLowerAlphanumeric("CatalogItem", opt.Name); err != nil {
		return err
	}

	if opt.ARM == "" && opt.X86 == "" {
		return util.NewInputError("At least one image must be specified")
	}

	if opt.Registry != "remote" && opt.Registry != "local" {
		return util.NewInputError("Registry must be either 'remote' or 'local'")
	}

	return nil
}
