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
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

func validatePortMapping(port *MicroservicePortMapping, agentsByName map[string]*client.AgentInfo) error {
	if port.Protocol != "" {
		port.Protocol = strings.ToLower(port.Protocol)
		protocol := port.Protocol
		if protocol != "tcp" && protocol != "http" {
			return NewInputError(fmt.Sprintf("Protocol %s is not supported. Valid protocols are tcp and http\n", protocol))
		}
	}
	if port.Public != nil && port.Public.Router != nil {
		if port.Public.Router.Host != "" {
			if port.Public.Router.Host != client.DefaultRouterName {
				agent, found := agentsByName[port.Public.Router.Host]
				if !found {
					return NewNotFoundError(fmt.Sprintf("Could not find port host %s\n", port.Public.Router.Host))
				}
				port.Public.Router.Host = agent.UUID
			}
		}
	}
	return nil
}

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

func createRoutes(routes []Route, microserviceByName map[string]*client.MicroserviceInfo, clt *client.Client) (err error) { // nolint:deadcode,unused
	for _, route := range routes {
		fromMsvc := microserviceByName[route.From]
		toMsvc := microserviceByName[route.To]
		if err = clt.CreateMicroserviceRoute(fromMsvc.UUID, toMsvc.UUID); err != nil {
			return
		}
	}
	return nil
}

func mapRouteToClientRouteRequest(route Route) client.ApplicationRouteCreateRequest {
	return client.ApplicationRouteCreateRequest{
		From: route.From,
		To:   route.To,
		Name: route.Name,
	}
}

func mapVariablesToClientVariables(variables []TemplateVariable) (result []client.TemplateVariable) {
	if len(variables) == 0 {
		return
	}
	for _, variable := range variables {
		clientVariable := client.TemplateVariable{
			Key:          variable.Key,
			Description:  variable.Description,
			DefaultValue: variable.DefaultValue,
			Value:        variable.Value,
		}
		result = append(result, clientVariable)
	}
	return
}

func mapTemplateToClientTemplate(template *ApplicationTemplate) (result *client.ApplicationTemplate) {
	if template != nil {
		result = &client.ApplicationTemplate{
			Name: template.Name,
		}
		result.Variables = mapVariablesToClientVariables(template.Variables)
	}
	return
}
