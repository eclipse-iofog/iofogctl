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

package client

import (
	"encoding/json"
	"fmt"
)

// GetMicroserviceByName retrieves a microservice information using Controller REST API
func (clt *Client) GetMicroserviceByName(name string) (response *MicroserviceInfo, err error) {
	listMsvcs, err := clt.GetAllMicroservices()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(listMsvcs.Microservices); i++ {
		if listMsvcs.Microservices[i].Name == name {
			return &listMsvcs.Microservices[i], nil
		}
	}
	return nil, NewNotFoundError(fmt.Sprintf("Could not find a microservice named %s", name))
}

// GetMicroserviceByID retrieves a microservice information using Controller REST API
func (clt *Client) GetMicroserviceByID(UUID string) (response *MicroserviceInfo, err error) {
	body, err := clt.doRequest("GET", fmt.Sprintf("/microservices/%s", UUID), nil)
	if err != nil {
		return
	}

	response = new(MicroserviceInfo)
	if err = json.Unmarshal(body, response); err != nil {
		return
	}
	return
}

// CreateMicroservice creates a microservice using Controller REST API
func (clt *Client) CreateMicroservice(request MicroserviceCreateRequest) (*MicroserviceInfo, error) {
	// Init empty arrays
	if request.Volumes == nil {
		request.Volumes = []MicroserviceVolumeMapping{}
	}
	if request.Env == nil {
		request.Env = []MicroserviceEnvironment{}
	}
	if request.Ports == nil {
		request.Ports = []MicroservicePortMapping{}
	}
	if request.Routes == nil {
		request.Routes = []string{}
	}
	if request.Commands == nil {
		request.Commands = []string{}
	}

	// Make request
	body, err := clt.doRequest("POST", "/microservices", request)
	if err != nil {
		return nil, err
	}
	response := &MicroserviceCreateResponse{}
	if err = json.Unmarshal(body, response); err != nil {
		return nil, err
	}
	return clt.GetMicroserviceByID(response.UUID)
}

// GetMicroservicesPerFlow returns a list of microservices in a specific flow using Controller REST API
func (clt *Client) GetMicroservicesPerFlow(flowID int) (response *MicroserviceListResponse, err error) {
	body, err := clt.doRequest("GET", fmt.Sprintf("/microservices?flowId=%d", flowID), nil)
	if err != nil {
		return
	}
	response = new(MicroserviceListResponse)
	err = json.Unmarshal(body, response)
	return
}

// GetAllMicroservices returns all microservices on the Controller by listing all flows,
// then getting a list of microservices per flow.
func (clt *Client) GetAllMicroservices() (response *MicroserviceListResponse, err error) {
	flows, err := clt.GetAllFlows()
	if err != nil {
		return nil, err
	}
	response = new(MicroserviceListResponse)

	for _, flow := range flows.Flows {
		listPerFlow, err := clt.GetMicroservicesPerFlow(flow.ID)
		if err != nil {
			continue
		}
		response.Microservices = append(response.Microservices, listPerFlow.Microservices...)
	}
	return
}

// GetMicroservicePortMapping retrieves a microservice port mappings using Controller REST API
func (clt *Client) GetMicroservicePortMapping(UUID string) (response *MicroservicePortMappingListResponse, err error) {
	body, err := clt.doRequest("GET", fmt.Sprintf("/microservices/%s/port-mapping", UUID), nil)
	if err != nil {
		return
	}

	response = new(MicroservicePortMappingListResponse)
	err = json.Unmarshal(body, response)
	return
}

// DeleteMicroservicePortMapping deletes a microservice port mapping using Controller REST API
func (clt *Client) DeleteMicroservicePortMapping(UUID string, portMapping MicroservicePortMapping) (err error) {
	_, err = clt.doRequest("DELETE", fmt.Sprintf("/microservices/%s/port-mapping/%d", UUID, portMapping.Internal), nil)
	return
}

// CreateMicroservicePortMapping creates a microservice port mapping using Controller REST API
func (clt *Client) CreateMicroservicePortMapping(UUID string, portMapping MicroservicePortMapping) (err error) {
	_, err = clt.doRequest("POST", fmt.Sprintf("/microservices/%s/port-mapping", UUID), portMapping)
	return
}

func portMappingsToMap(mappings []MicroservicePortMapping) map[int]MicroservicePortMapping {
	response := make(map[int]MicroservicePortMapping)
	for _, port := range mappings {
		response[port.Internal] = port
	}
	return response
}

func samePortMapping(currentMapping, newMapping MicroservicePortMapping) bool {
	if newMapping.Host == "" {
		newMapping.Host = DefaultRouterName
	}
	if newMapping.Protocol == "" {
		newMapping.Protocol = "http"
	}
	return (currentMapping.Internal == newMapping.Internal &&
		currentMapping.Public == newMapping.Public &&
		currentMapping.Protocol == newMapping.Protocol &&
		currentMapping.External == newMapping.External &&
		currentMapping.Host == newMapping.Host)
}

func (clt *Client) updateMicroservicePortMapping(UUID string, newPortMappings []MicroservicePortMapping) (err error) {
	currentPortMappings, err := clt.GetMicroservicePortMapping(UUID)
	if err != nil {
		return
	}

	currentPortMappingMap := portMappingsToMap(currentPortMappings.PortMappings)
	newPortMappingMap := portMappingsToMap(newPortMappings)

	// Remove outdated ports
	for _, currentMapping := range currentPortMappings.PortMappings {
		if newPortMapping, found := newPortMappingMap[currentMapping.Internal]; !found || (found && !samePortMapping(currentMapping, newPortMapping)) {
			if err = clt.DeleteMicroservicePortMapping(UUID, currentMapping); err != nil {
				return
			}
		}
	}

	// Create missing mappings
	for _, newMapping := range newPortMappings {
		if currentMapping, found := currentPortMappingMap[newMapping.Internal]; !found || (found && !samePortMapping(currentMapping, newMapping)) {
			if err = clt.CreateMicroservicePortMapping(UUID, newMapping); err != nil {
				return
			}
		}
	}

	return
}

func (clt *Client) GetAllMicroservicePublicPorts() (response []MicroservicePublicPort, err error) {
	body, err := clt.doRequest("GET", "/microservices/public-ports", nil)
	if err != nil {
		return
	}

	response = make([]MicroservicePublicPort, 0)
	err = json.Unmarshal(body, &response)
	return
}

func mapFromArray(arr []string) map[string]bool {
	result := make(map[string]bool)
	for _, str := range arr {
		result[str] = true
	}
	return result
}

// CreateMicroserviceRoute creates a microservice route using Controller REST API
func (clt *Client) CreateMicroserviceRoute(UUID, destUUID string) (err error) {
	_, err = clt.doRequest("POST", fmt.Sprintf("/microservices/%s/routes/%s", UUID, destUUID), nil)
	return
}

// DeleteMicroserviceRoute deletes a microservice route using Controller REST API
func (clt *Client) DeleteMicroserviceRoute(UUID, destUUID string) (err error) {
	_, err = clt.doRequest("DELETE", fmt.Sprintf("/microservices/%s/routes/%s", UUID, destUUID), nil)
	return
}

func (clt *Client) UpdateMicroserviceRoutes(UUID string, currentRoutes, newRoutes []string) (err error) {
	currentRouteMap := mapFromArray(currentRoutes)
	newRouteMap := mapFromArray(newRoutes)

	// Remove unused routes
	for _, currentRouteDest := range currentRoutes {
		_, found := newRouteMap[currentRouteDest]
		if !found {
			if err = clt.DeleteMicroserviceRoute(UUID, currentRouteDest); err != nil {
				return
			}
		}
	}

	// Create missing routes
	for _, newRouteDest := range newRoutes {
		_, found := currentRouteMap[newRouteDest]
		if !found {
			if err = clt.CreateMicroserviceRoute(UUID, newRouteDest); err != nil {
				return
			}
		}
	}
	return
}

// UpdateMicroservice patches a microservice using the Controller REST API
func (clt *Client) UpdateMicroservice(request MicroserviceUpdateRequest) (*MicroserviceInfo, error) {
	// Get current routes
	currentMsvc, err := clt.GetMicroserviceByID(request.UUID)
	if err != nil {
		return nil, err
	}

	// Update microservice
	_, err = clt.doRequest("PATCH", fmt.Sprintf("/microservices/%s", request.UUID), request)
	if err != nil {
		return nil, err
	}

	// Update routing
	if err = clt.UpdateMicroserviceRoutes(request.UUID, currentMsvc.Routes, request.Routes); err != nil {
		return nil, err
	}

	// Update Ports mapping
	if err = clt.updateMicroservicePortMapping(request.UUID, request.Ports); err != nil {
		return nil, err
	}

	return clt.GetMicroserviceByID(request.UUID)
}

// DeleteMicroservice deletes a microservice using Controller REST API
func (clt *Client) DeleteMicroservice(UUID string) (err error) {
	_, err = clt.doRequest("DELETE", fmt.Sprintf("/microservices/%s", UUID), nil)
	return
}
