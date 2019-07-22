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

package install

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"regexp"
	"strings"
	"time"
)

type Controller struct {
	ssh  *util.SecureShellClient
	host string
}

func NewController(user, host string, port int, privKeyFilename string) *Controller {
	ssh := util.NewSecureShellClient(user, host, privKeyFilename)
	ssh.SetPort(port)
	return &Controller{
		ssh:  ssh,
		host: host,
	}
}

func (ctrl *Controller) Install() (err error) {
	defer util.SpinStop()

	util.SpinStart("Connecting to remote server " + ctrl.host)
	// Connect to server
	if err = ctrl.ssh.Connect(); err != nil {
		return
	}
	defer ctrl.ssh.Disconnect()

	// Specify install script
	branch := util.GetVersion().Branch
	installURL := fmt.Sprintf("https://raw.githubusercontent.com/eclipse-iofog/iofogctl/%s/script/install_controller.sh", branch)

	// Define commands
	cmds := []string{
		"curl " + installURL + " | tee /tmp/install_controller.sh",
		"chmod +x /tmp/install_controller.sh",
		"/tmp/install_controller.sh",
	}

	// Execute commands
	util.SpinStart("Installing Controller and Connector on remote server")
	for _, cmd := range cmds {
		_, err = ctrl.ssh.Run(cmd)
		if err != nil {
			return
		}
	}

	// Wait for Controller
	util.SpinStart("Waiting for Controller")
	if err = ctrl.ssh.RunUntil(
		regexp.MustCompile("\"status\":\"online\""),
		fmt.Sprintf("curl --request GET --url http://localhost:%s/api/v3/status", iofog.ControllerPortString),
	); err != nil {
		return
	}

	// Wait for Connector
	util.SpinStart("Waiting for Connector")
	if err = ctrl.ssh.RunUntil(
		regexp.MustCompile("\"status\":\"running\""),
		fmt.Sprintf("curl --request POST --url http://localhost:%s/api/v2/status --header 'Content-Type: application/x-www-form-urlencoded' --data mappingid=all", iofog.ControllerPortString),
	); err != nil {
		return
	}

	return
}

func (ctrl *Controller) Configure(user client.User) (err error) {
	ctrlEndpoint := fmt.Sprintf("%s:%s", ctrl.host, iofog.ControllerPortString)
	connectorIP := "localhost"
	_, err = configureController(ctrlEndpoint, connectorIP, user)
	return
}

func configureController(ctrlEndpoint string, connectorIP string, user client.User) (token string, err error) {
	defer util.SpinStop()

	ctrl := client.New(ctrlEndpoint)

	// Create user (this is the first API call and the service might need to resolve IP to new pods so we retry)
	util.SpinStart("Creating ioFog user on Controller")
	connected := false
	iter := 0
	for !connected {
		// Time out
		if iter > 60 {
			err = util.NewInternalError("Failed to create new user with Controller")
			return
		}
		// Try to create the user
		if err = ctrl.CreateUser(user); err != nil {
			// Retry if connection is refused, this is usually only necessary on K8s Controller
			if strings.Contains(err.Error(), "connection refused") {
				time.Sleep(time.Millisecond * 1000)
				iter = iter + 1
				continue
			}
			// Account already exists, proceed to login
			if strings.Contains(err.Error(), "already an account associated") {
				connected = true
				continue
			}
			// Return the error otherwise
			return
		}
		// No error, connected
		connected = true
		continue
	}

	// Get token
	util.SpinStart("Logging in to retrieve access token")
	loginRequest := client.LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	}
	loginResponse, err := ctrl.Login(loginRequest)
	if err != nil {
		return
	}
	token = loginResponse.AccessToken

	// Connect Controller with Connector
	util.SpinStart("Provisioning Connector")
	connectorRequest := client.ConnectorInfo{
		IP:      connectorIP,
		DevMode: true,
		Domain:  connectorIP,
		Name:    "gke",
	}
	if err = ctrl.AddConnector(connectorRequest, token); err != nil {
		return
	}

	return
}
