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

package describe

import (
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type routeExecutor struct {
	namespace string
	name      string
	filename  string
}

func newRouteExecutor(namespace, name, filename string) *routeExecutor {
	return &routeExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *routeExecutor) GetName() string {
	return exe.name
}

func (exe *routeExecutor) Execute() error {
	_, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}

	// Connect to Controller
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	appName, routeName, err := clientutil.ParseFQName(exe.name, "Route")
	if err != nil {
		return err
	}

	// Get Route
	route, err := clt.GetRoute(appName, routeName)
	if err != nil {
		return err
	}

	// Convert route details
	from, err := clientutil.GetMicroserviceName(exe.namespace, route.SourceMicroserviceUUID)
	if err != nil {
		return err
	}
	to, err := clientutil.GetMicroserviceName(exe.namespace, route.DestMicroserviceUUID)
	if err != nil {
		return err
	}

	// Convert to YAML
	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.RouteKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: rsc.Route{
			From: from,
			To:   to,
			Name: routeName,
		},
	}

	if exe.filename == "" {
		if err := util.Print(header); err != nil {
			return err
		}
	} else {
		if err := util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
