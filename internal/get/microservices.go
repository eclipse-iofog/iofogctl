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

package get

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type microserviceExecutor struct {
	namespace string
	client    *client.Client
	flows     []client.FlowInfo
	msvcPerID map[string]*client.MicroserviceInfo
}

func newMicroserviceExecutor(namespace string) *microserviceExecutor {
	a := &microserviceExecutor{}
	a.namespace = namespace
	a.msvcPerID = make(map[string]*client.MicroserviceInfo)
	return a
}

func (exe *microserviceExecutor) init(controller *config.Controller) (err error) {
	exe.client = client.New(controller.Endpoint)
	if err = exe.client.Login(client.LoginRequest{Email: controller.IofogUser.Email, Password: controller.IofogUser.Password}); err != nil {
		return
	}
	flows, err := exe.client.GetAllFlows()
	if err != nil {
		return
	}
	exe.flows = flows.Flows
	for _, flow := range exe.flows {
		listMsvcs, err := exe.client.GetMicroservicesPerFlow(flow.ID)
		if err != nil {
			return err
		}
		for i := 0; i < len(listMsvcs.Microservices); i++ {
			exe.msvcPerID[listMsvcs.Microservices[i].UUID] = &listMsvcs.Microservices[i]
		}
	}
	return
}

func (exe *microserviceExecutor) Execute() error {
	// Get controller config details
	controllers, err := config.GetControllers(exe.namespace)
	if err != nil {
		return err
	}
	if len(controllers) == 0 {
		// Generate empty output
		return exe.generateMicroserviceOutput()
	}
	if len(controllers) > 1 {
		errMessage := fmt.Sprintf("This namespace contains %d Controller(s), you must have one, and only one.", len(controllers))
		return util.NewInputError(errMessage)
	}
	// Fetch data
	if err = exe.init(&controllers[0]); err != nil {
		return err
	}

	return exe.generateMicroserviceOutput()
}

func (exe *microserviceExecutor) generateMicroserviceOutput() (err error) {

	// Generate table and headers
	table := make([][]string, len(exe.msvcPerID)+1)
	headers := []string{"MICROSERVICE", "STATUS", "CONFIG", "ROUTES", "VOLUMES", "PORTS"}
	table[0] = append(table[0], headers...)

	// Populate rows
	count := 0
	for _, ms := range exe.msvcPerID {
		routes := ""
		for idx, route := range ms.Routes {
			if idx == 0 {
				routes += exe.msvcPerID[route].Name
			} else {
				routes += fmt.Sprintf(", %s", exe.msvcPerID[route].Name)
			}
		}
		volumes := ""
		for idx, volume := range ms.Volumes {
			if idx == 0 {
				volumes += fmt.Sprintf("%s:%s", volume.HostDestination, volume.ContainerDestination)
			} else {
				volumes += fmt.Sprintf(", %s:%s", volume.HostDestination, volume.ContainerDestination)
			}
		}
		ports := ""
		for idx, port := range ms.Ports {
			if idx == 0 {
				ports += fmt.Sprintf("%d:%d", port.External, port.Internal)
			} else {
				ports += fmt.Sprintf(", %d:%d", port.External, port.Internal)
			}
		}
		row := []string{
			ms.Name,
			"-",
			ms.Config,
			routes,
			volumes,
			ports,
		}
		table[count+1] = append(table[count+1], row...)
		count++
	}

	// Print the table
	err = print(table)
	if err != nil {
		return err
	}

	return nil
}
