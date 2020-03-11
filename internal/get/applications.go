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

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
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
	// Fetch data
	if err := exe.init(); err != nil {
		return err
	}
	return exe.generateApplicationOutput()
}

func (exe *applicationExecutor) init() (err error) {
	exe.client, err = internal.NewControllerClient(exe.namespace)
	if err != nil {
		if err.Error() == "This control plane does not have controller" {
			return nil
		}
		return err
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

		// Filter System microservices
		for _, ms := range listMsvcs.Microservices {
			if util.IsSystemMsvc(ms) {
				continue
			}
			exe.msvcsPerFlow[flow.ID] = append(exe.msvcsPerFlow[flow.ID], ms)
		}
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
		status := "INACTIVE"
		if flow.IsActivated == true {
			status = "RUNNING"
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
