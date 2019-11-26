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
	"os/user"
	"regexp"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
)

type localExecutor struct {
	namespace            string
	name                 string
	cnct                 *config.Connector
	client               *install.LocalContainer
	localConnectorConfig *install.LocalContainerConfig
	iofogUser            config.IofogUser
	containersNames      []string
}

func newLocalExecutor(namespace string, cnct *config.Connector, client *install.LocalContainer) (*localExecutor, error) {
	imageMap := make(map[string]string)
	imageMap["connector"] = cnct.Container.Image
	return &localExecutor{
		namespace: namespace,
		name:      cnct.Name,
		cnct:      cnct,
		client:    client,
		localConnectorConfig: install.NewLocalConnectorConfig(cnct.Container.Image, install.Credentials{
			User:     cnct.Container.Credentials.User,
			Password: cnct.Container.Credentials.Password,
		}),
	}, nil
}

func (exe *localExecutor) GetName() string {
	return exe.name
}

func (exe *localExecutor) cleanContainers() {
	for _, name := range exe.containersNames {
		if errClean := exe.client.CleanContainer(name); errClean != nil {
			util.PrintNotify(fmt.Sprintf("Could not clean Controller container: %v", errClean))
		}
	}
}

func (exe *localExecutor) deployContainers() error {
	// Deploy connector image
	util.SpinStart("Deploying Connector")

	// If container already exists, clean it
	connectorContainerName := exe.localConnectorConfig.ContainerName
	if _, err := exe.client.GetContainerByName(connectorContainerName); err == nil {
		if err := exe.client.CleanContainer(connectorContainerName); err != nil {
			return err
		}
	}

	if _, err := exe.client.DeployContainer(exe.localConnectorConfig); err != nil {
		return err
	}

	exe.containersNames = append(exe.containersNames, exe.localConnectorConfig.ContainerName)
	// Wait for public API
	util.SpinStart("Waiting for Connector API")
	if err := exe.client.WaitForCommand(
		install.GetLocalContainerName("connector"),
		regexp.MustCompile("iofog-connector is up and running."),
		"iofog-connector",
		"status",
	); err != nil {
		return err
	}

	// Provision the Connector with Controller
	IP, err := exe.client.GetContainerIP(install.GetLocalContainerName("controller"))
	if err != nil {
		return err
	}
	controllerEndpoint := fmt.Sprintf("%s:%s", IP, iofog.ControllerPortString)
	ctrlClient := client.New(controllerEndpoint)
	loginRequest := client.LoginRequest{
		Email:    exe.iofogUser.Email,
		Password: exe.iofogUser.Password,
	}
	if err := ctrlClient.Login(loginRequest); err != nil {
		return err
	}
	if err := ctrlClient.AddConnector(client.ConnectorInfo{
		IP:     exe.localConnectorConfig.Host,
		Domain: exe.localConnectorConfig.Host,
		Name:   exe.name,
	}); err != nil {
		return err
	}

	return nil
}

func (exe *localExecutor) Execute() error {
	// Get Control Plane
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil || len(controlPlane.Controllers) == 0 {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Connector")
		return err
	}
	exe.iofogUser = controlPlane.IofogUser

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
	connectorContainerConfig := exe.localConnectorConfig
	exe.cnct.Endpoint = fmt.Sprintf("%s:%s", connectorContainerConfig.Host, connectorContainerConfig.Ports[0].Host)
	exe.cnct.Host = connectorContainerConfig.Host
	exe.cnct.SSH.User = currUser.Username
	exe.cnct.Created = util.NowUTC()

	return nil
}
