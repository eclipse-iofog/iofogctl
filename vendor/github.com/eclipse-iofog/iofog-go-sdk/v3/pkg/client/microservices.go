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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
)

// GetMicroserviceByName retrieves a microservice information using Controller REST API
func (clt *Client) GetMicroserviceByName(appName, name string) (response *MicroserviceInfo, err error) {
	listMsvcs, err := clt.GetMicroservicesByApplication(appName)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(listMsvcs.Microservices); i++ {
		if listMsvcs.Microservices[i].Name == name {
			return &listMsvcs.Microservices[i], nil
		}
	}
	return nil, NewNotFoundError(fmt.Sprintf("Could not find a microservice named %s/%s", appName, name))
}

// GetMicroserviceByID retrieves a microservice information using Controller REST API
func (clt *Client) GetMicroserviceByID(uuid string) (response *MicroserviceInfo, err error) {
	body, err := clt.doRequest("GET", fmt.Sprintf("/microservices/%s", uuid), nil)
	if err != nil {
		return
	}

	response = new(MicroserviceInfo)
	if err = json.Unmarshal(body, response); err != nil {
		return
	}
	return
}

// CreateMicroserviceFromYAML creates a new microservice using the Controller REST API
// It sends the yaml file to Controller REST API
func (clt *Client) CreateMicroserviceFromYAML(file io.Reader) (*MicroserviceInfo, error) {
	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	part, _ := writer.CreateFormFile("microservice", "microservice.yaml")
	_, err := io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	writer.Close()

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}
	body, err := clt.doRequestWithHeaders("POST", "/microservices/yaml", requestBody, headers)

	if err != nil {
		return nil, err
	}
	response := MicroserviceCreateResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return clt.GetMicroserviceByID(response.UUID)
}

// GetMicroservicesPerFlow (DEPRECATED) returns a list of microservices in a specific flow using Controller REST API
func (clt *Client) GetMicroservicesPerFlow(flowID int) (response *MicroserviceListResponse, err error) {
	body, err := clt.doRequest("GET", fmt.Sprintf("/microservices?flowId=%d", flowID), nil)
	if err != nil {
		return
	}
	response = new(MicroserviceListResponse)
	err = json.Unmarshal(body, response)
	return
}

// GetMicroservicesByApplication returns a list of microservices in a specific application using Controller REST API
func (clt *Client) GetMicroservicesByApplication(application string) (response *MicroserviceListResponse, err error) {
	body, err := clt.doRequest("GET", fmt.Sprintf("/microservices?application=%s", application), nil)
	if err != nil {
		return
	}
	response = new(MicroserviceListResponse)
	err = json.Unmarshal(body, response)
	return
}

// GetAllMicroservices returns all microservices on the Controller by listing all flows,
// then getting a list of microservices per flow.
func (clt *Client) getAllMicroservicesDeprecated() (response *MicroserviceListResponse, err error) {
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

// GetAllMicroservices returns all microservices on the Controller across all (non-system) flows
func (clt *Client) getAllMicroservices() (response *MicroserviceListResponse, err error) {
	body, err := clt.doRequest("GET", "/microservices", nil)
	if err != nil {
		return
	}
	response = new(MicroserviceListResponse)
	err = json.Unmarshal(body, response)
	return
}

func (clt *Client) GetAllMicroservices() (response *MicroserviceListResponse, err error) {
	major, minor, patch, err := clt.GetVersionNumbers()
	if err != nil {
		return
	}
	isCapable := (major >= 2 && minor >= 0 && patch >= 2)
	if strings.Contains(clt.status.version, "dev") || isCapable {
		return clt.getAllMicroservices()
	}
	return clt.getAllMicroservicesDeprecated()
}

// GetMicroservicePortMapping retrieves a microservice port mappings using Controller REST API
func (clt *Client) GetMicroservicePortMapping(uuid string) (response *MicroservicePortMappingListResponse, err error) {
	body, err := clt.doRequest("GET", fmt.Sprintf("/microservices/%s/port-mapping", uuid), nil)
	if err != nil {
		return
	}

	response = new(MicroservicePortMappingListResponse)
	err = json.Unmarshal(body, response)
	return
}

// DeleteMicroservicePortMapping deletes a microservice port mapping using Controller REST API
func (clt *Client) DeleteMicroservicePortMapping(uuid string, portMapping *MicroservicePortMappingInfo) (err error) {
	_, err = clt.doRequest("DELETE", fmt.Sprintf("/microservices/%s/port-mapping/%v", uuid, portMapping.Internal), nil)
	return
}

// CreateMicroservicePortMapping creates a microservice port mapping using Controller REST API
func (clt *Client) CreateMicroservicePortMapping(uuid string, portMapping *MicroservicePortMappingInfo) (err error) {
	_, err = clt.doRequest("POST", fmt.Sprintf("/microservices/%s/port-mapping", uuid), portMapping)
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
func (clt *Client) CreateMicroserviceRoute(uuid, destUUID string) (err error) {
	_, err = clt.doRequest("POST", fmt.Sprintf("/microservices/%s/routes/%s", uuid, destUUID), nil)
	return
}

// DeleteMicroserviceRoute deletes a microservice route using Controller REST API
func (clt *Client) DeleteMicroserviceRoute(uuid, destUUID string) (err error) {
	_, err = clt.doRequest("DELETE", fmt.Sprintf("/microservices/%s/routes/%s", uuid, destUUID), nil)
	return
}

func (clt *Client) UpdateMicroserviceRoutes(uuid string, currentRoutes, newRoutes []string) (err error) {
	currentRouteMap := mapFromArray(currentRoutes)
	newRouteMap := mapFromArray(newRoutes)

	// Remove unused routes
	for _, currentRouteDest := range currentRoutes {
		_, found := newRouteMap[currentRouteDest]
		if !found {
			if err = clt.DeleteMicroserviceRoute(uuid, currentRouteDest); err != nil {
				return
			}
		}
	}

	// Create missing routes
	for _, newRouteDest := range newRoutes {
		_, found := currentRouteMap[newRouteDest]
		if !found {
			if err = clt.CreateMicroserviceRoute(uuid, newRouteDest); err != nil {
				return
			}
		}
	}
	return
}

// UpdateMicroserviceFromYAML updates a microservice using the Controller REST API
// It sends the yaml file to Controller REST API
func (clt *Client) UpdateMicroserviceFromYAML(uuid string, file io.Reader) (*MicroserviceInfo, error) {
	requestBody := &bytes.Buffer{}
	writer := multipart.NewWriter(requestBody)
	part, _ := writer.CreateFormFile("microservice", "microservice.yaml")
	_, err := io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	writer.Close()

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	_, err = clt.doRequestWithHeaders("PATCH", fmt.Sprintf("/microservices/yaml/%s", uuid), requestBody, headers)
	if err != nil {
		return nil, err
	}
	return clt.GetMicroserviceByID(uuid)
}

// DeleteMicroservice deletes a microservice using Controller REST API
func (clt *Client) DeleteMicroservice(uuid string) (err error) {
	_, err = clt.doRequest("DELETE", fmt.Sprintf("/microservices/%s", uuid), nil)
	return
}
