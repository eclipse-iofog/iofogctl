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
	"regexp"
	"strings"

	pb "github.com/schollz/progressbar"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
)

type localExecutor struct {
	opt                   *Options
	client                *iofog.LocalContainer
	localControllerConfig *iofog.LocalControllerConfig
	localUserConfig       *iofog.LocalUserConfig
	containersNames       []string
	pb                    *pb.ProgressBar
}

func newLocalExecutor(opt *Options, client *iofog.LocalContainer) *localExecutor {
	return &localExecutor{
		opt:                   opt,
		client:                client,
		localControllerConfig: iofog.NewLocalControllerConfig(opt.Name, opt.Images),
		localUserConfig:       iofog.GetLocalUserConfig(opt.Namespace, opt.Name),
		pb:                    pb.New(100),
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
	controllerContainerConfig := exe.localControllerConfig.ContainerMap["controller"]
	connectorContainerConfig := exe.localControllerConfig.ContainerMap["connector"]
	controllerContainerName := controllerContainerConfig.ContainerName
	connectorContainerName := connectorContainerConfig.ContainerName

	// Deploy controller image
	_, err := exe.client.DeployContainer(controllerContainerConfig)
	if err != nil {
		return err
	}
	exe.pb.Add(10)

	exe.containersNames = append(exe.containersNames, controllerContainerName)
	// Wait for public API
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
	exe.pb.Add(6)

	// Deploy connector image
	if _, err := exe.client.DeployContainer(connectorContainerConfig); err != nil {
		// Remove previously deployed Controller
		if errClean := exe.client.CleanContainer(controllerContainerName); errClean != nil {
			fmt.Printf("Could not clean container %v", errClean)
		}
		return err
	}
	exe.pb.Add(10)

	exe.containersNames = append(exe.containersNames, connectorContainerName)
	// Wait for public API
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
	exe.pb.Add(7)

	return nil
}

func (exe *localExecutor) install() error {
	controllerContainerConfig := exe.localControllerConfig.ContainerMap["controller"]
	connectorContainerConfig := exe.localControllerConfig.ContainerMap["connector"]

	ctrlIP := fmt.Sprintf("%s:%s", controllerContainerConfig.Host, controllerContainerConfig.Ports[0].Host)
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
	exe.pb.Add(10)
	// Login
	loginResponse, err := ctrl.Login(iofog.LoginRequest{Email: user.Email, Password: user.Password})
	if err != nil {
		return err
	}
	exe.pb.Add(10)
	// Provision Connector
	connectorIP := connectorContainerConfig.Host
	connectorName := connectorContainerConfig.ContainerName
	err = ctrl.AddConnector(iofog.ConnectorInfo{
		IP:      connectorIP,
		Name:    connectorName,
		Domain:  connectorContainerConfig.Host,
		DevMode: true,
	}, loginResponse.AccessToken)
	exe.pb.Add(13)
	return err
}

func (exe *localExecutor) Execute() error {
	exe.pb.Add(1)
	defer exe.pb.Clear()
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
		Name:      exe.opt.Name,
		User:      currUser.Username,
		Endpoint:  fmt.Sprintf("%s:%s", controllerContainerConfig.Host, controllerContainerConfig.Ports[0].Host),
		Host:      controllerContainerConfig.Host,
		Images:    exe.opt.Images,
		IofogUser: exe.localUserConfig.IofogUser,
	}
	err = config.AddController(exe.opt.Namespace, configEntry)
	if err != nil {
		exe.cleanContainers()
		return err
	}

	exe.pb.Add(33)
	if err = config.Flush(); err != nil {
		exe.cleanContainers()
		return err
	}

	fmt.Printf("\nController %s/%s successfully deployed.\n", exe.opt.Namespace, exe.opt.Name)
	return nil
}
