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
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
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

	podStatuses := make([]string, 0)
	// Handle k8s
	baseControlPlane, err := ns.GetControlPlane()
	if controlPlane, ok := baseControlPlane.(*rsc.KubernetesControlPlane); ok {
		if err := updateControllerPods(controlPlane, namespace); err != nil {
			return err
		}
		ns.SetControlPlane(controlPlane)
		if err := config.Flush(); err != nil {
			return err
		}
		for idx := range controlPlane.ControllerPods {
			podStatuses = append(podStatuses, controlPlane.ControllerPods[idx].Status)
		}
	}

	// Handle remote and local
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
		// Handle k8s pod statuses
		if len(podStatuses) != 0 && idx < len(podStatuses) {
			status = podStatuses[idx]
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

func updateControllerPods(controlPlane *rsc.KubernetesControlPlane, namespace string) (err error) {
	// Clear existing
	controlPlane.ControllerPods = make([]rsc.KubernetesController, 0)
	// Get pods
	installer, err := install.NewKubernetes(controlPlane.KubeConfig, namespace)
	if err != nil {
		return
	}
	pods, err := installer.GetControllerPods()
	if err != nil {
		return
	}
	// Add pods
	for idx := range pods {
		k8sPod := rsc.KubernetesController{
			Endpoint: controlPlane.Endpoint,
			PodName:  pods[idx].Name,
			Status:   pods[idx].Status,
		}
		if err := controlPlane.AddController(&k8sPod); err != nil {
			return err
		}
	}
	return
}
