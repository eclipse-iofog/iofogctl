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

package deployconnector

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"os/user"
	"regexp"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type localExecutor struct {
	namespace             string
	name                  string
	cnct                  *config.Connector
	client                *install.LocalContainer
	localControllerConfig *install.LocalControllerConfig
	iofogUser             config.IofogUser
	containersNames       []string
}

func newLocalExecutor(namespace string, cnct *config.Connector, controlPlane config.ControlPlane, client *install.LocalContainer) (*localExecutor, error) {
	return &localExecutor{
		namespace:             namespace,
		name:                  cnct.Name,
		cnct:                  cnct,
		client:                client,
		localControllerConfig: install.NewLocalControllerConfig(make(map[string]string)),
		iofogUser:             controlPlane.IofogUser,
	}, nil
}

func (exe *localExecutor) GetName() string {
	return exe.name
}

func (exe *localExecutor) cleanContainers() {
	for _, name := range exe.containersNames {
		if errClean := exe.client.CleanContainer(name); errClean != nil {
			fmt.Printf("Could not clean Controller container %v", errClean)
		}
	}
}

func (exe *localExecutor) deployContainers() error {
	connectorContainerConfig := exe.localControllerConfig.ContainerMap["connector"]
	connectorContainerName := connectorContainerConfig.ContainerName

	// Deploy connector image
	util.SpinStart("Deploying Connector")
	if _, err := exe.client.DeployContainer(connectorContainerConfig); err != nil {
		// Remove previously deployed Controller
		if errClean := exe.client.CleanContainer(connectorContainerName); errClean != nil {
			fmt.Printf("Could not clean container %v", errClean)
		}
		return err
	}

	exe.containersNames = append(exe.containersNames, connectorContainerName)
	// Wait for public API
	util.SpinStart("Waiting for Connector API")
	if err := exe.client.WaitForCommand(
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

	// Provision the Connector with Controller
	ctrlClient := client.New("localhost:" + iofog.ControllerPortString)
	loginRequest := client.LoginRequest{
		Email:    exe.iofogUser.Email,
		Password: exe.iofogUser.Password,
	}
	if err := ctrlClient.Login(loginRequest); err != nil {
		return err
	}
	if err := ctrlClient.AddConnector(client.ConnectorInfo{
		IP:     exe.localControllerConfig.ContainerMap["connector"].Host,
		Domain: exe.localControllerConfig.ContainerMap["connector"].Host,
		Name:   exe.name,
	}); err != nil {
		return err
	}

	return nil
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
	// Update Connector (return through pointer)
	connectorContainerConfig := exe.localControllerConfig.ContainerMap["connector"]
	exe.cnct.Endpoint = fmt.Sprintf("%s:%s", connectorContainerConfig.Host, connectorContainerConfig.Ports[0].Host)
	exe.cnct.Host = connectorContainerConfig.Host
	exe.cnct.User = currUser.Username
	exe.cnct.Created = util.NowUTC()

	return nil
}
