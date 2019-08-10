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
)

type applicationExecutor struct {
	namespace    string
	client       *client.Client
	flows        []client.FlowInfo
	msvcsPerFlow map[int][]client.MicroserviceInfo
}

func newApplicationExecutor(namespace string) *applicationExecutor {
	c := &applicationExecutor{}
	c.namespace = namespace
	c.msvcsPerFlow = make(map[int][]client.MicroserviceInfo)
	return c
}

func (exe *applicationExecutor) GetName() string {
	return ""
}

func (exe *applicationExecutor) Execute() error {
	// Get controller config details
	controllers, err := config.GetControllers(exe.namespace)
	if err != nil {
		return err
	}
	if len(controllers) == 0 {
		// Generate empty output
		return exe.generateApplicationOutput()
	}
	// Fetch data
	if err = exe.init(&controllers[0]); err != nil {
		return err
	}
	return exe.generateApplicationOutput()
}

func (exe *applicationExecutor) init(controller *config.Controller) (err error) {
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
		exe.msvcsPerFlow[flow.ID] = append(exe.msvcsPerFlow[flow.ID], listMsvcs.Microservices...)
	}
	return
}

func (exe *applicationExecutor) generateApplicationOutput() error {
	// Generate table and headers
	table := make([][]string, len(exe.flows)+1)
	headers := []string{"APPLICATION", "STATUS", "MICROSERVICES"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, flow := range exe.flows {
		status := "Inactive"
		if flow.IsActivated == true {
			status = "Running"
		}
		msvcs := ""
		first := true
		if len(exe.msvcsPerFlow[flow.ID]) > 5 {
			msvcs = fmt.Sprintf("%d microservices", len(exe.msvcsPerFlow[flow.ID]))
		} else {
			for _, msvc := range exe.msvcsPerFlow[flow.ID] {
				if first == true {
					msvcs += fmt.Sprintf("%s", msvc.Name)
				} else {
					msvcs += fmt.Sprintf(", %s", msvc.Name)
				}
				first = false
			}
		}
		row := []string{
			flow.Name,
			status,
			msvcs,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	// Print table
	err := print(table)
	if err != nil {
		return err
	}

	return nil
}
