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

package deployroute

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	"github.com/eclipse-iofog/iofogctl/v3/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	yaml "gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Name      string
	Yaml      []byte
}

type executor struct {
	namespace string
	name      string
	edge      rsc.EdgeResource
}

func (exe *executor) GetName() string {
	return "deploying Edge Resource " + exe.name
}

func (exe *executor) Execute() (err error) {
	if _, err = config.GetNamespace(exe.namespace); err != nil {
		return
	}

	// Translate edge resource to client type
	edge := &client.EdgeResourceMetadata{
		Name:              exe.name,
		Description:       exe.edge.Description,
		Version:           exe.edge.Version,
		InterfaceProtocol: exe.edge.InterfaceProtocol,
		Display:           exe.edge.Display,
		OrchestrationTags: exe.edge.OrchestrationTags,
		Interface: client.HTTPEdgeResource{
			Endpoints: exe.edge.Interface.Endpoints,
		},
		Custom: exe.edge.Custom,
	}
	// Connect to Controller
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return
	}

	// Create the resource
	if err = clt.UpdateHTTPEdgeResource(edge.Name, edge); err != nil {
		return
	}

	return
}

func NewExecutor(opt Options) (execute.Executor, error) {
	// Unmarshal file
	var edge rsc.EdgeResource
	if err := yaml.UnmarshalStrict(opt.Yaml, &edge); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return nil, err
	}
	// Validate input
	if opt.Name == "" {
		return nil, util.NewInputError("Did not specify metadata.name")
	}
	if err := util.IsLowerAlphanumeric("Edge Resource", opt.Name); err != nil {
		return nil, err
	}
	// Return executor
	return &executor{
		namespace: opt.Namespace,
		name:      opt.Name,
		edge:      edge,
	}, nil
}
