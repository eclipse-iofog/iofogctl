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
	"regexp"
	"strings"
	"time"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type ControllerOptions struct {
	User              string
	Host              string
	Port              int
	PrivKeyFilename   string
	Version           string
	PackageCloudToken string
	IofogUser         IofogUser
}

type Controller struct {
	*ControllerOptions
	ssh *util.SecureShellClient
}

func NewController(options *ControllerOptions) *Controller {
	ssh := util.NewSecureShellClient(options.User, options.Host, options.PrivKeyFilename)
	ssh.SetPort(options.Port)
	return &Controller{
		ControllerOptions: options,
		ssh:               ssh,
	}
}

func (ctrl *Controller) Install() (err error) {
	// Connect to server
	verbose("Connecting to server")
	if err = ctrl.ssh.Connect(); err != nil {
		return
	}
	defer ctrl.ssh.Disconnect()

	// Copy installation scripts to remote host
	verbose("Copying install files to server")
	installControllerScript := util.GetStaticFile("install_controller.sh")
	reader := strings.NewReader(installControllerScript)
	if err := ctrl.ssh.CopyTo(reader, "/tmp/", "install_controller.sh", "0775", len(installControllerScript)); err != nil {
		return err
	}

	// Define commands
	cmds := []string{
		fmt.Sprintf("/tmp/install_controller.sh %s %s", ctrl.Version, ctrl.PackageCloudToken),
	}

	// Execute commands
	for _, cmd := range cmds {
		verbose("Running command: " + cmd)
		_, err = ctrl.ssh.Run(cmd)
		if err != nil {
			return
		}
	}

	// Specify errors to ignore while waiting
	ignoredErrors := []string{
		"Process exited with status 7", // curl: (7) Failed to connect to localhost port 8080: Connection refused
	}
	// Wait for Controller
	verbose("Waiting for Controller")
	if err = ctrl.ssh.RunUntil(
		regexp.MustCompile("\"status\":\"online\""),
		fmt.Sprintf("curl --request GET --url http://localhost:%s/api/v3/status", iofog.ControllerPortString),
		ignoredErrors,
	); err != nil {
		return
	}

	// Create Iofog user
	endpoint := fmt.Sprintf("%s:%s", ctrl.Host, iofog.ControllerPortString)
	if err = createUser(endpoint, ctrl.IofogUser); err != nil {
		return
	}

	return
}

func (ctrl *Controller) Stop() (err error) {
	// Connect to server
	if err = ctrl.ssh.Connect(); err != nil {
		return
	}
	defer ctrl.ssh.Disconnect()

	// TODO: Clear the database
	// Define commands
	cmds := []string{
		"sudo iofog-controller stop",
	}

	// Execute commands
	for _, cmd := range cmds {
		_, err = ctrl.ssh.Run(cmd)
		if err != nil {
			return
		}
	}

	return
}

func createUser(endpoint string, user IofogUser) (err error) {
	ctrlClient := client.New(endpoint)

	// Create user (this is the first API call and the service might need to resolve IP to new pods so we retry)
	connected := false
	iter := 0
	for !connected {
		// Time out
		if iter > 60 {
			err = util.NewInternalError("Failed to create new user with Controller")
			return
		}
		// Try to create the user
		if err = ctrlClient.CreateUser(client.User(user)); err != nil {
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

	return
}
