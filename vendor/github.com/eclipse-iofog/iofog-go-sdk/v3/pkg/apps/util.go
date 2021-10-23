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

package apps

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

func validateRoutes(routes []string, microserviceByName map[string]*client.MicroserviceInfo) (routesUUIDs []string, err error) { // nolint:deadcode,unused
	// Validate routes
	for _, route := range routes {
		msvc, foundTo := microserviceByName[route]
		if !foundTo {
			return routesUUIDs, NewNotFoundError(fmt.Sprintf("Could not find microservice [%s] required by a route", route))
		}
		routesUUIDs = append(routesUUIDs, msvc.UUID)
	}
	return routesUUIDs, nil
}

func createRoutes(routes []Route, microserviceByName map[string]*client.MicroserviceInfo, clt *client.Client) error { // nolint:deadcode,unused
	for _, route := range routes {
		fromMsvc := microserviceByName[route.From]
		toMsvc := microserviceByName[route.To]
		if err := clt.CreateMicroserviceRoute(fromMsvc.UUID, toMsvc.UUID); err != nil {
			return err
		}
	}
	return nil
}
