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

package iofog

import (
	"fmt"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"regexp"
	"strings"
	"time"
)

type ControllerInstaller struct {
	ssh  *util.SecureShellClient
	host string
}

func NewControllerInstaller(user, host string, port int, privKeyFilename string) *ControllerInstaller {
	ssh := util.NewSecureShellClient(user, host, privKeyFilename)
	ssh.SetPort(port)
	return &ControllerInstaller{
		ssh:  ssh,
		host: host,
	}
}

func (instlr *ControllerInstaller) Install() (err error) {
	defer util.SpinStop()

	util.SpinStart("Connecting to remote server " + instlr.host)
	// Connect to server
	if err = instlr.ssh.Connect(); err != nil {
		return
	}
	defer instlr.ssh.Disconnect()

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
		_, err = instlr.ssh.Run(cmd)
		if err != nil {
			return
		}
	}

	// Wait for Controller
	util.SpinStart("Waiting for Controller")
	if err = instlr.ssh.RunUntil(
		regexp.MustCompile("\"status\":\"online\""),
		"curl --request GET --url http://localhost:54421/api/v3/status", // TODO: replace hardcode
	); err != nil {
		return
	}

	// Wait for Connector
	util.SpinStart("Waiting for Connector")
	if err = instlr.ssh.RunUntil(
		regexp.MustCompile("\"status\":\"running\""),
		"curl --request POST --url http://localhost:8080/api/v2/status --header 'Content-Type: application/x-www-form-urlencoded' --data mappingid=all",
	); err != nil {
		return
	}

	return
}

func (instlr *ControllerInstaller) Configure(user User) (err error) {
	_, err = configureController(instlr.host+":54421", instlr.host, user) // TODO: change hardcode
	return
}

func configureController(ctrlEndpoint string, connectorIP string, user User) (token string, err error) {
	defer util.SpinStop()

	ctrl := NewController(ctrlEndpoint)

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
	loginRequest := LoginRequest{
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
	connectorRequest := ConnectorInfo{
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
