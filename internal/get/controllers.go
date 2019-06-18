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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type controllerExecutor struct {
	namespace string
}

func newControllerExecutor(namespace string) *controllerExecutor {
	c := &controllerExecutor{}
	c.namespace = namespace
	return c
}

func (exe *controllerExecutor) Execute() error {
	// Get controller config details
	controllers, err := config.GetControllers(exe.namespace)
	if err != nil {
		return err
	}

	// Generate table and headers
	table := make([][]string, len(controllers)+1)
	headers := []string{"CONTROLLER", "STATUS", "AGE", "UPTIME"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, ctrlConfig := range controllers {
		// Instantiate connection to controller
		ctrl := iofog.NewController(ctrlConfig.Endpoint)

		// Ping status
		ctrlStatus, err := ctrl.GetStatus()
		uptime := "-"
		status := "Failing"
		if err == nil {
			uptimeSec := ctrlStatus.UptimeTimeMsUTC / int64(1000)
			uptime, _ = util.ElapsedUTC(util.FromIntUTC(uptimeSec), util.NowUTC())
			status = ctrlStatus.Status
		}

		// Get age
		age := "-"
		if ctrlConfig.Created != "" {
			age, _ = util.ElapsedUTC(ctrlConfig.Created, util.NowUTC())
		}
		row := []string{
			ctrlConfig.Name,
			status,
			age,
			uptime,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	// Print table
	err = print(table)
	if err != nil {
		return err
	}

	return nil
}
