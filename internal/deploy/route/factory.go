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
	appName   string
	route     rsc.Route
}

func (exe *executor) GetName() string {
	return "deploying Route " + exe.name
}

func (exe *executor) Execute() (err error) {
	if _, err = config.GetNamespace(exe.namespace); err != nil {
		return
	}

	// Connect to Controller
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return
	}

	if err = clt.UpdateRoute(&client.Route{
		Name:        exe.name,
		From:        exe.route.From,
		To:          exe.route.To,
		Application: exe.appName,
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

	appName, routeName, err := clientutil.ParseFQName(opt.Name, "Route")
	if err != nil {
		return nil, err
	}

	if route.Name == "" {
		route.Name = routeName
	}

	if route.Name != routeName {
		return nil, util.NewInputError(fmt.Sprintf("Mismatch between metadata.name [%s] and spec.name [%s]", opt.Name, route.Name))
	}

	return &executor{
		namespace: opt.Namespace,
		name:      routeName,
		appName:   appName,
		route:     route,
	}, nil
}
