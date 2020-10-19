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

package get

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type applicationExecutor struct {
	namespace           string
	client              *client.Client
	flows               []client.FlowInfo
	msvcsPerApplication map[int][]client.MicroserviceInfo
}

func newApplicationExecutor(namespace string) *applicationExecutor {
	c := &applicationExecutor{}
	c.namespace = namespace
	c.msvcsPerApplication = make(map[int][]client.MicroserviceInfo)
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
	printNamespace(exe.namespace)
	table, err := exe.generateApplicationOutput()
	if err != nil {
		return err
	}
	return print(table)
}

func (exe *applicationExecutor) init() (err error) {
	exe.client, err = iutil.NewControllerClient(exe.namespace)
	if err != nil {
		if rsc.IsNoControlPlaneError(err) {
			return nil
		}
		return err
	}
	applications, err := exe.client.GetAllApplications()
	// If notfound error, try legacy
	if _, ok := err.(*client.NotFoundError); err != nil && ok {
		if err = exe.initLegacy(); err != nil {
			return err
		}
	} else {
		// Map applications to flow
		// TODO: Use Application instead of flow
		exe.flows = []client.FlowInfo{}
		for _, application := range applications.Applications {
			exe.flows = append(exe.flows, client.FlowInfo{
				Name:        application.Name,
				IsActivated: application.IsActivated,
				Description: application.Description,
				IsSystem:    application.IsSystem,
				UserID:      application.UserID,
				ID:          application.ID,
			})
			listMsvcs, err := exe.client.GetMicroservicesByApplication(application.Name)
			if err != nil {
				return err
			}

			// Filter System microservices
			for _, ms := range listMsvcs.Microservices {
				if util.IsSystemMsvc(ms) {
					continue
				}
				exe.msvcsPerApplication[application.ID] = append(exe.msvcsPerApplication[application.ID], ms)
			}
		}
	}
	return
}

func (exe *applicationExecutor) generateApplicationOutput() (table [][]string, err error) {
	// Generate table and headers
	table = make([][]string, len(exe.flows)+1)
	headers := []string{"APPLICATION", "RUNNING", "MICROSERVICES"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, flow := range exe.flows {
		nbMsvcs := len(exe.msvcsPerApplication[flow.ID])
		runningMsvcs := 0
		msvcs := ""
		first := true
		for _, msvc := range exe.msvcsPerApplication[flow.ID] {
			if first == true {
				msvcs += fmt.Sprintf("%s", msvc.Name)
			} else {
				msvcs += fmt.Sprintf(", %s", msvc.Name)
			}
			first = false
			if msvc.Status.Status == "RUNNING" {
				runningMsvcs++
			}
		}

		if nbMsvcs > 5 {
			msvcs = fmt.Sprintf("%d microservices", len(exe.msvcsPerApplication[flow.ID]))
		}

		status := fmt.Sprintf("%d/%d", runningMsvcs, nbMsvcs)

		row := []string{
			flow.Name,
			status,
			msvcs,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	return
}
