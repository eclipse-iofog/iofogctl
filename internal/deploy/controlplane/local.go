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

package deploycontrolplane

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"os/user"
	"regexp"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type localExecutor struct {
	namespace             string
	ctrl                  config.Controller
	client                *install.LocalContainer
	localControllerConfig *install.LocalControllerConfig
	localUserConfig       *install.LocalUserConfig
	containersNames       []string
}

func newLocalExecutor(namespace string, ctrl config.Controller, client *install.LocalContainer) *localExecutor {
	if ctrl.IofogUser.Email == "" {
		ctrl.IofogUser = config.NewRandomUser()
	}
	return &localExecutor{
		namespace:             namespace,
		ctrl:                  ctrl,
		client:                client,
		localControllerConfig: install.NewLocalControllerConfig(ctrl.Name, ctrl.Images),
		localUserConfig:       &install.LocalUserConfig{ctrl.IofogUser},
	}
}

func (exe *localExecutor) cleanContainers() {
	for _, name := range exe.containersNames {
		if errClean := exe.client.CleanContainer(name); errClean != nil {
			fmt.Printf("Could not clean Controller container %v", errClean)
		}
	}
}

func (exe *localExecutor) deployContainers() error {
	defer util.SpinStop()

	controllerContainerConfig := exe.localControllerConfig.ContainerMap["controller"]
	connectorContainerConfig := exe.localControllerConfig.ContainerMap["connector"]
	controllerContainerName := controllerContainerConfig.ContainerName
	connectorContainerName := connectorContainerConfig.ContainerName

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

	// Deploy connector image
	util.SpinStart("Deploying Connector")
	if _, err := exe.client.DeployContainer(connectorContainerConfig); err != nil {
		// Remove previously deployed Controller
		if errClean := exe.client.CleanContainer(controllerContainerName); errClean != nil {
			fmt.Printf("Could not clean container %v", errClean)
		}
		return err
	}

	exe.containersNames = append(exe.containersNames, connectorContainerName)
	// Wait for public API
	util.SpinStart("Waiting for Connector API")
	if err = exe.client.WaitForCommand(
		regexp.MustCompile("\"status\":\"running\""),
		"curl",
		"--request",
		"POST",
		"--url",
		fmt.Sprintf("http://%s:%s/api/v2/status", connectorContainerConfig.Host, connectorContainerConfig.Ports[0].Host),
		"--header",
		"'Content-Type: application/x-www-form-urlencoded'",
		"--data",
		"mappingid=all",
	); err != nil {
		return err
	}

	return nil
}

func (exe *localExecutor) install() error {
	defer util.SpinStop()

	controllerContainerConfig := exe.localControllerConfig.ContainerMap["controller"]
	connectorContainerConfig := exe.localControllerConfig.ContainerMap["connector"]

	ctrlIP := fmt.Sprintf("%s:%s", controllerContainerConfig.Host, controllerContainerConfig.Ports[0].Host)
	ctrl := client.New(ctrlIP)
	// Assign user
	user := client.User{
		Name:     exe.localUserConfig.Name,
		Surname:  exe.localUserConfig.Surname,
		Email:    exe.localUserConfig.Email,
		Password: exe.localUserConfig.Password,
	}
	// Create user
	util.SpinStart("Creating new user")
	if err := ctrl.CreateUser(user); err != nil {
		if !strings.Contains(err.Error(), "already an account associated") {
			return err
		}
	}
	// Login
	if err := ctrl.Login(client.LoginRequest{Email: user.Email, Password: user.Password}); err != nil {
		return err
	}
	// Provision Connector
	util.SpinStart("Provisioning Connector to Controller")
	connectorIP := connectorContainerConfig.Host
	connectorName := connectorContainerConfig.ContainerName
	if err := ctrl.AddConnector(client.ConnectorInfo{
		IP:      connectorIP,
		Name:    connectorName,
		Domain:  connectorContainerConfig.Host,
		DevMode: true,
	}); err != nil {
		return err
	}

	return nil
}

func (exe *localExecutor) GetName() string {
	return exe.ctrl.Name
}

func (exe *localExecutor) Execute() error {
	defer util.SpinStop()
	controllerContainerConfig := exe.localControllerConfig.ContainerMap["controller"]

	// Get current user
	currUser, err := user.Current()
	if err != nil {
		return err
	}

	// Deploy Controller and Connector images
	err = exe.deployContainers()
	if err != nil {
		exe.cleanContainers()
		return err
	}

	// Create user, login, provision connector
	if err = exe.install(); err != nil {
		fmt.Printf("Cleaning containers... %v", err)
		exe.cleanContainers()
		return err
	}

	// Update configuration
	configEntry := config.Controller{
		Name:      exe.ctrl.Name,
		User:      currUser.Username,
		Endpoint:  fmt.Sprintf("%s:%s", controllerContainerConfig.Host, controllerContainerConfig.Ports[0].Host),
		Host:      controllerContainerConfig.Host,
		Images:    exe.ctrl.Images,
		IofogUser: exe.localUserConfig.IofogUser,
	}
	err = config.AddController(exe.namespace, configEntry)
	if err != nil {
		exe.cleanContainers()
		return err
	}

	if err = config.Flush(); err != nil {
		exe.cleanContainers()
		return err
	}

	return nil
}
