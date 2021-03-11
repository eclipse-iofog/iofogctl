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

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
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
	table, err := generateControllerOutput(exe.namespace)
	if err != nil {
		return err
	}
	printNamespace(exe.namespace)
	return print(table)
}

func generateControllerOutput(namespace string) (table [][]string, err error) {
	// Get controller config details
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return
	}

	podStatuses := []string{}
	// Handle k8s
	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		if rsc.IsNoControlPlaneError(err) {
			err = nil
		} else {
			return
		}
	}
	if controlPlane, ok := baseControlPlane.(*rsc.KubernetesControlPlane); ok {
		if err = updateControllerPods(controlPlane, namespace); err != nil {
			return
		}
		ns.SetControlPlane(controlPlane)
		if err = config.Flush(); err != nil {
			return
		}
		for idx := range controlPlane.ControllerPods {
			podStatuses = append(podStatuses, controlPlane.ControllerPods[idx].Status)
		}
	}

	// Handle remote and local
	controllers := ns.GetControllers()

	// Generate table and headers
	table = make([][]string, len(controllers)+1)
	headers := []string{"CONTROLLER", "STATUS", "AGE", "UPTIME", "VERSION", "ADDR", "PORT"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, ctrlConfig := range controllers {
		// Instantiate connection to controller
		ctrl, err := clientutil.NewControllerClient(namespace)
		if err != nil {
			return table, err
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
			ctrlStatus.Versions.Controller,
			addr,
			port,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	return table, err
}

func updateControllerPods(controlPlane *rsc.KubernetesControlPlane, namespace string) (err error) {
	// Clear existing
	controlPlane.ControllerPods = []rsc.KubernetesController{}
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
