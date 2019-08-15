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

package deploycontroller

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"os/user"
	"regexp"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type localExecutor struct {
	namespace             string
	ctrl                  *config.Controller
	client                *install.LocalContainer
	localControllerConfig *install.LocalControllerConfig
	iofogUser             config.IofogUser
	containersNames       []string
}

func newLocalExecutor(namespace string, ctrl *config.Controller, controlPlane config.ControlPlane, client *install.LocalContainer) (*localExecutor, error) {
	return &localExecutor{
		namespace:             namespace,
		ctrl:                  ctrl,
		client:                client,
		localControllerConfig: install.NewLocalControllerConfig(ctrl.Images),
		iofogUser:             controlPlane.IofogUser,
	}, nil
}

func (exe *localExecutor) cleanContainers() {
	for _, name := range exe.containersNames {
		if errClean := exe.client.CleanContainer(name); errClean != nil {
			fmt.Printf("Could not clean Controller container %v", errClean)
		}
	}
}

func (exe *localExecutor) deployContainers() error {

	controllerContainerConfig := exe.localControllerConfig.ContainerMap["controller"]
	controllerContainerName := controllerContainerConfig.ContainerName

	// Deploy controller image
	util.SpinStart("Deploying Controller container")
	_, err := exe.client.DeployContainer(controllerContainerConfig)
	if err != nil {
		return err
	}

	exe.containersNames = append(exe.containersNames, controllerContainerName)
	// Wait for public API
	util.SpinStart("Waiting for Controller API")
	if err = exe.client.WaitForCommand(
		regexp.MustCompile("\"status\":\"online\""),
		"curl",
		"--request",
		"GET",
		"--url",
		fmt.Sprintf("http://%s:%s/api/v3/status", controllerContainerConfig.Host, controllerContainerConfig.Ports[0].Host),
	); err != nil {
		return err
	}

	return nil
}

func (exe *localExecutor) GetName() string {
	return exe.ctrl.Name
}

func (exe *localExecutor) Execute() error {
	// Deploy Controller and Connector images
	if err := exe.deployContainers(); err != nil {
		exe.cleanContainers()
		return err
	}

	// Get current user
	currUser, err := user.Current()
	if err != nil {
		return err
	}

	// Update controller (its a pointer, this is returned to caller)
	controllerContainerConfig := exe.localControllerConfig.ContainerMap["controller"]
	exe.ctrl.Endpoint = fmt.Sprintf("%s:%s", controllerContainerConfig.Host, controllerContainerConfig.Ports[0].Host)
	exe.ctrl.Host = controllerContainerConfig.Host
	exe.ctrl.User = currUser.Username
	exe.ctrl.Created = util.NowUTC()

	return nil
}
