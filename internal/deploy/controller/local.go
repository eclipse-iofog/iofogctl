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
	"os/user"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
)

type localExecutor struct {
	opt                   *Options
	client                *iofog.LocalContainer
	localControllerConfig *iofog.LocalControllerConfig
	localUserConfig       *iofog.LocalUserConfig
	containersNames       []string
}

func newLocalExecutor(opt *Options, client *iofog.LocalContainer) *localExecutor {
	return &localExecutor{
		opt:                   opt,
		client:                client,
		localControllerConfig: iofog.NewLocalControllerConfig(opt.Name),
		localUserConfig:       iofog.GetLocalUserConfig(opt.Namespace, opt.Name),
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
	controllerImg, exists := exe.opt.Images["controller"]
	if !exists {
		controllerImg = exe.localControllerConfig.DefaultImages["controller"]
	}
	connectorImg, exists := exe.opt.Images["connector"]
	if !exists {
		connectorImg = exe.localControllerConfig.DefaultImages["connector"]
	}

	controllerContainerName := exe.localControllerConfig.ContainerNames["controller"]
	connectorContainerName := exe.localControllerConfig.ContainerNames["connector"]

	// Deploy controller image
	controllerPortMap := make(map[string]*iofog.LocalContainerPort)
	controllerPortMap[exe.localControllerConfig.ControllerPort.Host] = exe.localControllerConfig.ControllerPort.Container // 51121:51121/tcp
	_, err := exe.client.DeployContainer(controllerImg, controllerContainerName, controllerPortMap)
	if err != nil {
		return err
	}

	exe.containersNames = append(exe.containersNames, controllerContainerName)

	// Deploy connector image
	connectorPortMap := make(map[string]*iofog.LocalContainerPort)
	connectorPortMap[exe.localControllerConfig.ConnectorPort.Host] = exe.localControllerConfig.ConnectorPort.Container
	if _, err := exe.client.DeployContainer(connectorImg, connectorContainerName, connectorPortMap); err != nil {
		// Remove previously deployed Controller
		if errClean := exe.client.CleanContainer(controllerContainerName); errClean != nil {
			fmt.Printf("Could not clean container %v", errClean)
		}
		return err
	}

	exe.containersNames = append(exe.containersNames, connectorContainerName)

	return nil
}

func (exe *localExecutor) install() error {
	ctrlIP := fmt.Sprintf("%s:%s", exe.localControllerConfig.Host, exe.localControllerConfig.ControllerPort.Host)
	ctrl := iofog.NewController(ctrlIP)
	// Assign user
	user := iofog.User{
		Name:     exe.localUserConfig.Name,
		Surname:  exe.localUserConfig.Surname,
		Email:    exe.localUserConfig.Email,
		Password: exe.localUserConfig.Password,
	}
	// Create user
	if err := ctrl.CreateUser(user); err != nil {
		if !strings.Contains(err.Error(), "already an account associated") {
			return err
		}
	}
	// Login
	loginResponse, err := ctrl.Login(iofog.LoginRequest{Email: user.Email, Password: user.Password})
	if err != nil {
		return err
	}
	// Provision Connector
	connectorIP := fmt.Sprintf("%s:%s", exe.localControllerConfig.Host, exe.localControllerConfig.ConnectorPort.Host)
	connectorName := exe.localControllerConfig.ContainerNames["connector"]
	return ctrl.AddConnector(iofog.ConnectorInfo{
		IP:      connectorIP,
		Name:    connectorName,
		Domain:  exe.localControllerConfig.Host,
		DevMode: true,
	}, loginResponse.AccessToken)
}

func (exe *localExecutor) Execute() error {
	// Get current user
	currUser, err := user.Current()
	if err != nil {
		return err
	}

	// Deploy Controller and Connector images
	err = exe.deployContainers()
	if err != nil {
		return err
	}

	// Create user, login, provision connector
	if err = exe.install(); err != nil {
		exe.cleanContainers()
		return err
	}

	// Update configuration
	configEntry := config.Controller{
		Name:      exe.opt.Name,
		User:      currUser.Username,
		Host:      exe.localControllerConfig.Host,
		Images:    exe.opt.Images,
		IofogUser: exe.localUserConfig.IofogUser,
	}
	err = config.AddController(exe.opt.Namespace, configEntry)
	if err != nil {
		exe.cleanContainers()
		return err
	}

	if err = config.Flush(); err != nil {
		exe.cleanContainers()
		return err
	}

	fmt.Printf("\nController %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)
	return nil
}
