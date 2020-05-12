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
	"fmt"
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
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
	route     rsc.Route
}

func (exe executor) GetName() string {
	return "deploying Route " + exe.name
}

func (exe executor) Execute() (err error) {
	if _, err = config.GetNamespace(exe.namespace); err != nil {
		return
	}

	// Connect to Controller
	clt, err := iutil.NewControllerClient(exe.namespace)
	if err != nil {
		return
	}

	// Convert route details
	srcMsvcUUID, err := iutil.GetMicroserviceUUID(exe.namespace, exe.route.From)
	if err != nil {
		return
	}
	destMsvcUUID, err := iutil.GetMicroserviceUUID(exe.namespace, exe.route.To)
	if err != nil {
		return
	}

	if err = clt.CreateRoute(client.Route{
		Name:                   exe.name,
		SourceMicroserviceUUID: srcMsvcUUID,
		DestMicroserviceUUID:   destMsvcUUID,
	}); err != nil {
		return
	}
	return
}

func NewExecutor(opt Options) (execute.Executor, error) {
	// Unmarshal file
	var route rsc.Route
	if err := yaml.UnmarshalStrict(opt.Yaml, &route); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return nil, err
	}
	// Validate input
	if route.Name == "" && opt.Name == "" {
		return nil, util.NewInputError("Did not specify metadata.name or spec.name")
	}
	if route.Name == "" {
		route.Name = opt.Name
	}
	if opt.Name == "" {
		opt.Name = route.Name
	}
	if route.Name != opt.Name {
		return nil, util.NewInputError(fmt.Sprintf("Mismatch between metadata.name [%s] and spec.name [%s]", opt.Name, route.Name))
	}

	return executor{
		namespace: opt.Namespace,
		name:      opt.Name,
		route:     route,
	}, nil
}
