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

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type applicationExecutor struct {
	namespace           string
	client              *client.Client
	flows               []client.FlowInfo
	msvcsPerApplication map[int][]*client.MicroserviceInfo
}

func newApplicationExecutor(namespace string) *applicationExecutor {
	c := &applicationExecutor{}
	c.namespace = namespace
	c.msvcsPerApplication = make(map[int][]*client.MicroserviceInfo)
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
	table := exe.generateApplicationOutput()
	return print(table)
}

func (exe *applicationExecutor) init() (err error) {
	exe.client, err = clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		if rsc.IsNoControlPlaneError(err) {
			return nil
		}
		return err
	}
	applications, err := exe.client.GetAllApplications()
	// Try legacy if error is "not found"
	if _, ok := err.(*client.NotFoundError); ok {
		if err := exe.initLegacy(); err != nil {
			return err
		}
		// Successful legacy
		return nil
	}
	if err != nil {
		// Return errors that are not "not found"
		return err
	}
	// Execute non-legacy
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
		for idx := range listMsvcs.Microservices {
			msvc := &listMsvcs.Microservices[idx]
			if util.IsSystemMsvc(msvc) {
				continue
			}
			exe.msvcsPerApplication[application.ID] = append(exe.msvcsPerApplication[application.ID], msvc)
		}
	}
	return err
}

func (exe *applicationExecutor) generateApplicationOutput() (table [][]string) {
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
		for idx := range exe.msvcsPerApplication[flow.ID] {
			msvc := exe.msvcsPerApplication[flow.ID][idx]
			if first {
				msvcs += msvc.Name
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

	return table
}
