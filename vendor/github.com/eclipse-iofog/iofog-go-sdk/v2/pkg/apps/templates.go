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
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
)

type applicationTemplateExecutor struct {
	controller         IofogController
	template           ApplicationTemplate
	microserviceByName map[string]*client.MicroserviceInfo
	client             *client.Client
	agentsByName       map[string]*client.AgentInfo
}

func newApplicationTemplateExecutor(controller IofogController, template ApplicationTemplate) *applicationTemplateExecutor {
	exe := &applicationTemplateExecutor{
		controller:         controller,
		template:           template,
		microserviceByName: microserviceArrayToClientMap(template.Application.Microservices),
	}

	return exe
}

func (exe *applicationTemplateExecutor) execute() (err error) {
	// Init remote resources
	if err = exe.init(); err != nil {
		return
	}

	// Deploy application
	err = exe.deploy()

	return nil
}

func (exe *applicationTemplateExecutor) init() (err error) {
	if exe.controller.Token != "" {
		exe.client, err = client.NewWithToken(client.Options{Endpoint: exe.controller.Endpoint}, exe.controller.Token)
	} else {
		exe.client, err = client.NewAndLogin(client.Options{Endpoint: exe.controller.Endpoint}, exe.controller.Email, exe.controller.Password)
	}
	if err != nil {
		return
	}
	listAgents, err := exe.client.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return
	}

	exe.agentsByName = make(map[string]*client.AgentInfo)
	for i := 0; i < len(listAgents.Agents); i++ {
		exe.agentsByName[listAgents.Agents[i].Name] = &listAgents.Agents[i]
	}

	return
}

func (exe *applicationTemplateExecutor) deploy() (err error) {
	microservices, err := mapMicroservicesToClientMicroserviceRequests(exe.template.Application.Microservices, exe.agentsByName)
	if err != nil {
		return err
	}
	routes := mapRoutesToClientRouteRequests(exe.template.Application.Routes)
	variables := mapVariablesToClientVariables(exe.template.Variables)
	request := &client.ApplicationTemplateUpdateRequest{
		ApplicationTemplate: client.ApplicationTemplate{
			Description: exe.template.Description,
			Name:        exe.template.Name,
			Application: client.ApplicationTemplateInfo{
				Microservices: microservices,
				Routes:        routes,
			},
			Variables: variables,
		},
	}

	if _, err = exe.client.UpdateApplicationTemplate(request); err != nil {
		return err
	}
	return nil
}
