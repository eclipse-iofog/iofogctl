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
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"gopkg.in/yaml.v2"
)

const (
	errParseControllerURL = "failed to parse Controller endpoint as URL: %s"
)

// ApplicationData is data fetched from controller at init time
type ApplicationData struct {
	MicroserviceByName map[string]*client.MicroserviceInfo
	AgentsByName       map[string]*client.AgentInfo
	CatalogByID        map[int]*client.CatalogItemInfo
	RegistryByID       map[int]*client.RegistryInfo
	CatalogByName      map[string]*client.CatalogItemInfo
	FlowInfo           *client.FlowInfo
}

type microserviceExecutor struct {
	controller IofogController
	msvc       interface{}
	name       string
	appName    string
	uuid       string
	client     *client.Client
}

func ParseFQMsvcName(fqName string) (appName, name string, err error) {
	if fqName == "" {
		return "", "", NewInputError(fmt.Sprintf("Invalid microservice name %s", fqName))
	}
	splittedName := strings.Split(fqName, "/")
	switch len(splittedName) {
	case 1:
		return "", splittedName[0], nil
	case 2:
		return splittedName[0], splittedName[1], nil
	default:
		return "", "", NewInputError(fmt.Sprintf("Invalid microservice name %s", fqName))
	}
}

func newMicroserviceExecutor(controller IofogController, msvc interface{}, appName, name string) *microserviceExecutor {
	exe := &microserviceExecutor{
		controller: controller,
		msvc:       msvc,
		name:       name,
		appName:    appName,
	}

	return exe
}

func (exe *microserviceExecutor) execute() error {
	// Init remote resources
	if err := exe.init(); err != nil {
		return err
	}

	// Deploy microservice
	if _, err := exe.deploy(); err != nil {
		return err
	}
	return nil
}

func (exe *microserviceExecutor) init() (err error) {
	baseURL, err := url.Parse(exe.controller.Endpoint)
	if err != nil {
		return fmt.Errorf(errParseControllerURL, err.Error())
	}
	if exe.controller.Token != "" {
		exe.client, err = client.NewWithToken(client.Options{BaseURL: baseURL}, exe.controller.Token)
	} else {
		exe.client, err = client.NewAndLogin(client.Options{BaseURL: baseURL}, exe.controller.Email, exe.controller.Password)
	}
	if err != nil {
		return err
	}
	if exe.appName == "" {
		return NewInputError(fmt.Sprintf("Application name missing for microservice %s", exe.name))
	}
	listMsvcs, err := exe.client.GetMicroservicesByApplication(exe.appName)
	if err != nil {
		return err
	}

	for i := 0; i < len(listMsvcs.Microservices); i++ {
		// If msvc already exists, set UUID
		if listMsvcs.Microservices[i].Name == exe.name {
			if exe.uuid == "" {
				exe.uuid = listMsvcs.Microservices[i].UUID
			}
		}
	}
	return err
}

func (exe *microserviceExecutor) deploy() (newMsvc *client.MicroserviceInfo, err error) {
	if exe.uuid != "" {
		// Update microservice
		return exe.update()
	}
	// Create microservice
	return exe.create()
}

func (exe *microserviceExecutor) create() (newMsvc *client.MicroserviceInfo, err error) {
	file := IofogHeader{
		APIVersion: "iofog.org/v3",
		Kind:       MicroserviceKind,
		Metadata: HeaderMetadata{
			Name: strings.Join([]string{exe.appName, exe.name}, "/"),
		},
		Spec: exe.msvc,
	}
	yamlBytes, err := yaml.Marshal(file)
	if err != nil {
		return nil, err
	}
	return exe.client.CreateMicroserviceFromYAML(bytes.NewReader(yamlBytes))
}

func (exe *microserviceExecutor) update() (newMsvc *client.MicroserviceInfo, err error) {
	file := IofogHeader{
		APIVersion: "iofog.org/v3",
		Kind:       MicroserviceKind,
		Metadata: HeaderMetadata{
			Name: strings.Join([]string{exe.appName, exe.name}, "/"),
		},
		Spec: exe.msvc,
	}
	yamlBytes, err := yaml.Marshal(file)
	if err != nil {
		return nil, err
	}
	return exe.client.UpdateMicroserviceFromYAML(exe.uuid, bytes.NewReader(yamlBytes))
}
