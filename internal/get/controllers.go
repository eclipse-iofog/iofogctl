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
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type controllerExecutor struct {
	namespace string
}

func newControllerExecutor(namespace string) *controllerExecutor {
	c := &controllerExecutor{}
	c.namespace = namespace
	return c
}

func (exe *controllerExecutor) GetName() string {
	return ""
}

func (exe *controllerExecutor) Execute() error {
	return generateControllerOutput(exe.namespace, true)
}

func generateControllerOutput(namespace string, printNS bool) error {
	// Get controller config details
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}
	controllers := ns.GetControllers()

	// Generate table and headers
	table := make([][]string, len(controllers)+1)
	headers := []string{"CONTROLLER", "STATUS", "AGE", "UPTIME", "ADDR", "PORT"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, ctrlConfig := range controllers {
		// Instantiate connection to controller
		ctrl, err := iutil.NewControllerClient(namespace)
		if err != nil {
			return err
		}

		// Ping status
		ctrlStatus, err := ctrl.GetStatus()
		uptime := "-"
		status := "Failing"
		if err == nil {
			uptime = util.FormatDuration(time.Duration(int64(ctrlStatus.UptimeSeconds)) * time.Second)
			status = ctrlStatus.Status
		}

		// Get age
		age := "-"
		if ctrlConfig.GetCreatedTime() != "" {
			age, _ = util.ElapsedUTC(ctrlConfig.GetCreatedTime(), util.NowUTC())
		}
		addr, port := getAddressAndPort(ctrlConfig.GetEndpoint(), client.ControllerPortString)
		row := []string{
			ctrlConfig.GetName(),
			status,
			age,
			uptime,
			addr,
			port,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	if printNS {
		printNamespace(namespace)
	}
	// Print table
	err = print(table)
	if err != nil {
		return err
	}

	return nil
}
